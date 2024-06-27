package link

type Msg struct {
	Data []byte
	Seq  uint64
}

type Channel interface {
	WaitInit()
	Send() chan<- Msg
	Ack() <-chan uint64
	Recv() <-chan []byte
}

type Stream interface {
	NewChannel(name string) (channel Channel, err error)
}
