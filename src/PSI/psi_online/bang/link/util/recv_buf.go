package util

import (
	"bang/link"
	"errors"
	"sync"
)

type ReceiverBuffer struct {
	startSeq uint64
	buf      *MinHeapGenerics[link.Msg]
	seq      uint64

	mu sync.Mutex

	outCond sync.Cond

	close bool
}

var ErrBufferClosed = errors.New("buffer closed")

func NewReceiverBuffer[T any](startSeq uint64) *ReceiverBuffer {
	rb := &ReceiverBuffer{
		startSeq: startSeq,
	}
	rb.buf = NewMinHeapGenerics[link.Msg]([]link.Msg{}, func(a, b any) bool {
		return a.(link.Msg).Seq < b.(link.Msg).Seq
	})
	rb.outCond = *sync.NewCond(&rb.mu)

	return rb
}

func (rb *ReceiverBuffer) Push(msg link.Msg) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.close {
		return
	}

	rb.buf.Push(msg)

	if rb.buf.Top().Seq < rb.seq {
		rb.buf.Pop()
	}

	rb.outCond.Signal()
}

func (rb *ReceiverBuffer) Pop() (msg link.Msg, err error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

RETRY:

	if !rb.close && rb.buf.Empty() {
		rb.outCond.Wait()
	}

	if rb.close && rb.buf.Empty() {
		err = ErrBufferClosed
		return
	}

	var succ bool = false

	for !rb.buf.Empty() {
		h := rb.buf.Top()
		if h.Seq < rb.seq {
			rb.buf.Pop()
		} else if h.Seq > rb.seq {
			if !rb.close {
				rb.outCond.Wait()
			} else {
				err = ErrBufferClosed
				return
			}
		} else {
			msg = h
			succ = true
			rb.seq += 1
			rb.buf.Pop()
			break
		}
	}

	if !succ {
		goto RETRY
	}

	return
}

func (rb *ReceiverBuffer) Close() {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.close = true
	rb.outCond.Broadcast()
}
