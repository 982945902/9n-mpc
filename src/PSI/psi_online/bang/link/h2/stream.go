package h2

import (
	"bang/link"
	"bang/link/util"
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/middlewares/server/recovery"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/errors"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/hertz-contrib/http2/config"
	"github.com/hertz-contrib/http2/factory"
)

const push_msg_url_handle = "/link/post/:channel"
const push_msg_url_send = "http://%s/link/post/%s"
const sync_url_handle = "/sync/:channel"
const sync_url_send = "http://%s/sync/%s"

const (
	meta_id           = "id"
	meta_target       = "target"
	meta_consumer_seq = "consumer-seq"
)

const (
	channel_name_not_find_err = "channel name not find"
	channel_not_find_err      = "channel not find"
)

type send_post_typ func(ctx context.Context, msg link.Msg) (err error)

type channel struct {
	name string
	s    *stream

	consumer_seq uint64
	window_size  int

	store_path string

	lock sync.Mutex

	send_ch chan link.Msg
	ack_ch  chan uint64
	recv_ch chan []byte

	send_post send_post_typ

	buffer *util.ReceiverBuffer

	send_seq_heap *util.MinHeap[uint64]

	ctx context.Context

	sync chan struct{}
}

func (c *channel) watch(duration time.Duration) {
	timer := time.NewTimer(duration)
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-timer.C:
			util.WriteUint64ToFile(c.store_path, atomic.LoadUint64(&c.consumer_seq))
		}
	}
}

func newChannel(s *stream, ctx context.Context, name string, window_size int, store_path string, watch_duration time.Duration, send_post send_post_typ, recover bool) (c *channel, err error) {
	c = &channel{name: name, s: s, window_size: window_size, store_path: store_path}

	c.send_ch = make(chan link.Msg, window_size)
	c.recv_ch = make(chan []byte, window_size)
	c.ack_ch = make(chan uint64, window_size)
	if recover {
		c.consumer_seq, err = util.ReadUint64FromFile(c.store_path)
		c.buffer = util.NewReceiverBuffer[link.Msg](c.consumer_seq)
		go c.watch(watch_duration)
	} else {
		c.consumer_seq = 0
		c.buffer = util.NewReceiverBuffer[link.Msg](c.consumer_seq)

	}
	c.send_post = send_post
	c.send_seq_heap = util.NewMinHeap([]uint64{})
	c.ctx = ctx
	c.sync = make(chan struct{})

	go c.recv_loop()
	go c.send_loop(window_size)

	return
}

func (c *channel) inprcess(ctx *app.RequestContext) (err error) {
	consumer_seq_str := ctx.Request.Header.Get(meta_consumer_seq)
	if len(consumer_seq_str) == 0 {
		return errors.NewPublic("consumer_seq not find")
	}
	consumer_seq, err := strconv.ParseUint(consumer_seq_str, 10, 64)
	if err != nil {
		return errors.NewPublic("consumer_seq format error")
	}

	if atomic.LoadUint64(&c.consumer_seq) > consumer_seq {
		hlog.Warnf("Repeat consumer_seq")
		return
	}

	// hlog.Infof("consumer_seq:%d", consumer_seq)

	c.buffer.Push(link.Msg{Seq: consumer_seq, Data: ctx.Request.Body()})

	return
}

func (c *channel) recv_loop() {
	for {
		msg, err := c.buffer.Pop()
		if err != nil {
			return
		}

		c.recv_ch <- msg.Data
	}
}

func (c *channel) send_loop(window_size int) {
	send_ch := make(chan link.Msg, c.window_size)

	go func() {
		defer func() {
			close(send_ch)
		}()

		for {
			select {
			case <-c.ctx.Done():
				return
			case msg, ok := <-c.send_ch:
				if !ok {
					return
				}

				c.lock.Lock()
				c.send_seq_heap.Push(msg.Seq)
				c.lock.Unlock()

				send_ch <- msg
			}
		}
	}()

	for i := 0; i < window_size; i++ {
		go func() {
			for {
				select {
				case <-c.ctx.Done():
					return
				case msg, ok := <-send_ch:
					if !ok {
						return
					}
					if atomic.LoadUint64(&c.consumer_seq) > msg.Seq {
						hlog.Warnf("Repeat consumer_seq")
						return
					}

					util.RetryOnConflict(util.AlwaysRetry, func() error {
						err := c.send_post(c.ctx, msg)
						if err != nil {
							hlog.Errorf("Send msg[%d] error: %s", msg.Seq, err)
						}

						return err
					}, func(err error) bool {
						return true
					})

					c.lock.Lock()
					if c.send_seq_heap.Top() == msg.Seq {
						c.consumer_seq = msg.Seq
						c.send_seq_heap.Pop()
					} else {
						c.send_seq_heap.Remove(msg.Seq)
					}
					c.lock.Unlock()

					c.ack_ch <- msg.Seq
				}
			}
		}()
	}
}

func (c *channel) Send() chan<- link.Msg { return c.send_ch }
func (c *channel) Ack() <-chan uint64    { return c.ack_ch }
func (c *channel) Recv() <-chan []byte   { return c.recv_ch }

func (c *channel) WaitInit() {
	util.RetryOnConflict(util.AlwaysRetry, func() error {
		err := c.s.sync_channel(c.ctx, c.name)
		if err != nil {
			hlog.Errorf("sync_channel error: %s", err)
		}

		return err
	}, func(err error) bool {
		return true
	})
	<-c.sync
}

type stream struct {
	h  *server.Hertz
	c  *client.Client
	cf *link.Config

	chan_map sync.Map
	ctx      context.Context
	cancel   func()

	recover bool
}

func NewLink(cf *link.Config) (s *stream, err error) {
	s = &stream{
		cf: cf,
	}

	if cf.RecoverSupport {
		s.recover, err = util.IsRecover(s.cf.StorePath)
		if err != nil {
			return
		}
		if !s.recover {
			err = util.RunOnce(s.cf.StorePath)
			if err != nil {
				return
			}
		}
	} else {
		s.recover = false
	}

	s.h = server.New(server.WithHostPorts(cf.Host), server.WithH2C(true))
	s.h.AddProtocol("h2", factory.NewServerFactory())
	s.h.POST(push_msg_url_handle, s.handle_post)
	s.h.POST(sync_url_handle, s.handle_sync)
	s.h.Use(logAbortMiddleware())
	s.h.Use(recovery.Recovery())

	s.c, err = client.NewClient()
	if err != nil {
		return
	}
	s.c.SetClientFactory(factory.NewClientFactory(config.WithAllowHTTP(true)))

	s.ctx, s.cancel = context.WithCancel(context.Background())

	go func() {
		s.h.Spin()
	}()

	return
}

func (s *stream) NewChannel(name string) (chn link.Channel, err error) {
	chn, err = newChannel(s, s.ctx, name, s.cf.WindowSize,
		filepath.Join(s.cf.StorePath, name), time.Second, s.gen_post(name), s.recover)
	if err != nil {
		return
	}

	s.chan_map.Store(name, chn)

	return
}

func (s *stream) gen_post(ch_name string) func(ctx context.Context, msg link.Msg) (err error) {
	remote_url := fmt.Sprintf(push_msg_url_send, s.cf.Remote, ch_name)

	header := &protocol.RequestHeader{}
	header.SetMethod("POST")
	header.SetRequestURI(remote_url)
	header.Set(meta_id, s.cf.Id)
	header.Set(meta_target, s.cf.Target)

	return func(ctx context.Context, msg link.Msg) (err error) {
		req := protocol.AcquireRequest()
		rsp := protocol.AcquireResponse()

		header.CopyTo(&req.Header)
		req.SetHeader(meta_consumer_seq, strconv.FormatUint(msg.Seq, 10))
		req.SetBody(msg.Data)
		err = s.c.Do(ctx, req, rsp)
		if err != nil {
			return
		}
		if rsp.StatusCode() != 200 {
			err = errors.NewPublic(fmt.Sprintf("status code: %d", rsp.StatusCode()))
		}

		return
	}
}

func (s *stream) sync_channel(ctx context.Context, ch_name string) (err error) {
	req := protocol.AcquireRequest()
	rsp := protocol.AcquireResponse()

	req.SetMethod("POST")
	req.SetRequestURI(fmt.Sprintf(sync_url_send, s.cf.Remote, ch_name))
	req.SetHeader(meta_id, s.cf.Id)
	req.SetHeader(meta_target, s.cf.Target)

	err = s.c.Do(ctx, req, rsp)
	if err != nil {
		return
	}
	if rsp.StatusCode() != 200 {
		err = errors.NewPublic(fmt.Sprintf("status code: %d", rsp.StatusCode()))
	}

	return
}

func (s *stream) handle_post(c context.Context, ctx *app.RequestContext) {
	chan_name := ctx.Param("channel")
	if len(chan_name) == 0 {
		ctx.AbortWithMsg(channel_name_not_find_err, 500)
	}

	chan_obj, ok := s.chan_map.Load(chan_name)
	if !ok {
		ctx.AbortWithMsg(channel_not_find_err, 404)
	}

	err := chan_obj.(*channel).inprcess(ctx)
	if err != nil {
		ctx.AbortWithMsg(err.Error(), 500)
	}

	ctx.SetStatusCode(200)
}

func (s *stream) handle_sync(c context.Context, ctx *app.RequestContext) {
	chan_name := ctx.Param("channel")
	if len(chan_name) == 0 {
		ctx.AbortWithMsg(channel_name_not_find_err, 500)
	}

	chan_obj, ok := s.chan_map.Load(chan_name)
	if !ok {
		ctx.AbortWithMsg(channel_not_find_err, 404)
	}
	// func() {
	// 	defer func() {
	// 		recover()
	// 	}()
	// 	close(chan_obj.(*channel).sync)
	// }()
	close(chan_obj.(*channel).sync)

	ctx.SetStatusCode(200)
}

func logAbortMiddleware() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		ctx.Next(c)

		if ctx.IsAborted() {
			hlog.Errorf("Request aborted: %s", string(ctx.Response.Body()))
		}
	}
}
