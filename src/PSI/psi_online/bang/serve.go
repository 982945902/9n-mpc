package main

import (
	"bang/link"
	"bang/link/h2"
	"bang/mq"
	"bang/mq/nats"
	"bang/sdk"
	"context"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
)

type controler struct {
	s *server.Hertz

	ls link.Stream
	ms mq.Stream

	subjects sync.Map
}

type subject struct {
	cus          mq.Consumer
	pro          mq.Producer
	chn          link.Channel
	consumer_seq uint64
}

func (s *subject) slink() {
	go func() {
		defer func() {
			recover()
		}()

		ackMap := sync.Map{}

		go func() {
			defer func() {
				recover()
			}()

			for {
				msg, ok := <-s.cus.Consume()
				if !ok {
					return
				}

				msg.Seq = s.consumer_seq
				s.consumer_seq += 1

				s.chn.Send() <- link.Msg{
					Seq:  msg.Seq,
					Data: msg.Data,
				}

				ackMap.Store(msg.Seq, msg)
			}
		}()
		for {
			seq, ok := <-s.chn.Ack()
			if !ok {
				return
			}

			v, ok := ackMap.LoadAndDelete(seq)
			if ok {
				_ = v.(mq.Msg).Ack()
			}
		}
	}()

	go func() {
		defer func() {
			recover()
		}()

		for {
			msg, ok := <-s.chn.Recv()
			if !ok {
				return
			}

			s.pro.Push() <- msg
		}
	}()
}

const create_subject_url = "/bang/create"

type createSubjectReq struct {
	DownSubjectName string
	UpSubjectName   string
	LinkName        string
}

func (ctrl *controler) createSubject(c context.Context, ctx *app.RequestContext) {
	var req createSubjectReq
	err := ctx.BindAndValidate(&req)
	if err != nil {
		ctx.AbortWithError(500, err)
	}

	cus, err := ctrl.ms.NewConsumer(req.DownSubjectName)
	if err != nil {
		ctx.AbortWithError(500, err)
	}
	pro, err := ctrl.ms.NewProducer(req.UpSubjectName)
	if err != nil {
		ctx.AbortWithError(500, err)
	}
	chn, err := ctrl.ls.NewChannel(req.LinkName)
	if err != nil {
		ctx.AbortWithError(500, err)
	}
	chn.WaitInit()

	sub := &subject{cus: cus, pro: pro, chn: chn, consumer_seq: 0}
	ctrl.subjects.Store(req.LinkName, sub)

	go sub.slink()

	ctx.SetStatusCode(200)
}

func (ctrl *controler) Run() {
	ctrl.s.Spin()
}

func newControler(conf *Config) (c *controler, err error) {
	c = &controler{
		s: server.New(server.WithHostPorts(conf.Host)),
	}

	c.ls, err = h2.NewLink(&link.Config{
		Id:             conf.Id,
		Host:           conf.LinkHost,
		Domain:         conf.Domain,
		Target:         conf.Target,
		Remote:         conf.Remote,
		StorePath:      "",
		WindowSize:     conf.WindowSize,
		RecoverSupport: false,
	})
	if err != nil {
		return
	}

	c.ms, err = nats.NewStream(mq.Config{
		MqAddress:  conf.MqAddress,
		WindowSize: conf.WindowSize,
	})
	if err != nil {
		return
	}

	c.s.POST(create_subject_url, c.createSubject)

	port, _ := strconv.Atoi(strings.Split(conf.LinkHost, ":")[1])
	sdk.RegisterToProxy(conf.RedisServer, conf.RedisPassword, conf.Id, port)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		sdk.UnRegisterToProxy(conf.RedisServer, conf.RedisPassword, conf.Id)
		os.Exit(1)
	}()

	return
}
