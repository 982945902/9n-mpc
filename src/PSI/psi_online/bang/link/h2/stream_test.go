package h2

import (
	"bang/link"
	"math"
	"strconv"
	"sync"
	"testing"
)

var p1Cfg = link.Config{
	Id:             "test",
	Domain:         "jd",
	Host:           "127.0.0.1:8900",
	Target:         "cu",
	Remote:         "127.0.0.1:8901",
	StorePath:      "",
	WindowSize:     1,
	RecoverSupport: false,
}

var p2Cfg = link.Config{
	Id:             "test",
	Domain:         "cu",
	Host:           "127.0.0.1:8901",
	Target:         "jd",
	Remote:         "127.0.0.1:8900",
	StorePath:      "",
	WindowSize:     1,
	RecoverSupport: false,
}

func run_test(cfg *link.Config, send_or_recv int, t *testing.T) {
	s, err := NewLink(cfg)
	if err != nil {
		t.Fatal(err)
	}

	ch, err := s.NewChannel("test_ch")
	if err != nil {
		t.Fatal(err)
	}

	ch.WaitInit()

	for i := 0; i < math.MaxInt; i++ {
		if send_or_recv == 0 {
			data := strconv.FormatUint(uint64(i), 10)
			ch.Send() <- link.Msg{Seq: uint64(i), Data: []byte(data)}
			ack := <-ch.Ack()
			// if ack != uint64(i) {
			// 	t.Fatalf("ack error: %d != %d", ack, i)
			// }
			_ = ack
			// t.Logf("ack %d", ack)
			// t.Logf("send %d", i)

			// time.Sleep(time.Second)
		} else if send_or_recv == 1 {
			msg := <-ch.Recv()
			seq, _ := strconv.ParseUint(string(msg), 10, 64)
			if seq != uint64(i) {
				t.Logf("seq error:  %d != %d", seq, i)
				panic("error")
			}
			t.Logf("recv %d", i)
		}
	}

}

func TestMain(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		run_test(&p1Cfg, 0, t)
	}()

	go func() {
		defer wg.Done()
		run_test(&p2Cfg, 1, t)
	}()

	wg.Wait()
}
