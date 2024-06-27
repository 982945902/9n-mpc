package mq

type Msg struct {
	Data []byte
	Seq  uint64
	Ack  func() error
}

type Consumer interface {
	Consume() <-chan Msg
}

type Producer interface {
	Push() chan<- []byte
}

type Stream interface {
	NewConsumer(name string) (Consumer, error)
	NewProducer(name string) (Producer, error)
}
