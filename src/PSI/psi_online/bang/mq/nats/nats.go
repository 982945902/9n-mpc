package nats

import (
	"bang/link/util"
	"bang/mq"
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type consumer struct {
	name string
	js   jetstream.JetStream
	ss   jetstream.Stream
	cus  jetstream.Consumer
	ctx  context.Context

	buf chan mq.Msg
}

func (c *consumer) Consume() <-chan mq.Msg {
	return c.buf
}

func (c *consumer) worker() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			msg, err := c.cus.Next()
			if err != nil {
				continue
			}
			meta, err := msg.Metadata()
			if err != nil {
				_ = msg.Nak()
				continue
			}

			c.buf <- mq.Msg{
				Data: msg.Data(),
				Seq:  meta.Sequence.Consumer,
				Ack:  msg.Ack,
			}
		}

	}
}

type producer struct {
	name string
	js   jetstream.JetStream
	ctx  context.Context

	buf chan []byte
}

func (p *producer) Push() chan<- []byte {
	return p.buf
}

func (p *producer) worker() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case data := <-p.buf:
			util.RetryOnConflict(util.AlwaysRetry, func() error {
				_, err := p.js.Publish(p.ctx, p.name, data)

				return err
			}, func(err error) bool { return true })
		}
	}
}

type stream struct {
	c   *nats.Conn
	ctx context.Context

	window_size int
}

func (s *stream) NewConsumer(name string) (c mq.Consumer, err error) {
	js, err := jetstream.New(s.c)
	if err != nil {
		return
	}

	ss, err := js.CreateOrUpdateStream(s.ctx, jetstream.StreamConfig{
		Name:     name + "_stream",
		Subjects: []string{name},
	})
	if err != nil {
		return
	}

	cus, err := ss.CreateOrUpdateConsumer(s.ctx, jetstream.ConsumerConfig{
		Durable: "TestConsumerConsume",

		DeliverPolicy: jetstream.DeliverAllPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxDeliver:    -1,
		MaxAckPending: s.window_size,
	})
	if err != nil {
		return
	}

	c = &consumer{
		name: name,
		js:   js,
		ss:   ss,
		cus:  cus,
		ctx:  s.ctx,
		buf:  make(chan mq.Msg, s.window_size),
	}

	go c.(*consumer).worker()

	return
}
func (s *stream) NewProducer(name string) (p mq.Producer, err error) {
	js, err := jetstream.New(s.c)
	if err != nil {
		return
	}

	p = &producer{
		name: name,
		js:   js,
		ctx:  s.ctx,
		buf:  make(chan []byte, s.window_size),
	}

	go p.(*producer).worker()

	return
}

func NewStream(conf mq.Config) (s *stream, err error) {
	ctx := context.Background()

	c, err := nats.Connect(conf.MqAddress)
	if err != nil {
		return
	}

	s = &stream{
		c:           c,
		ctx:         ctx,
		window_size: conf.WindowSize,
	}

	return
}
