package main

import (
	"bang/link"
	"bang/link/h2"
	"bang/sdk"
	"math"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func debug() {

	hlog.Infof("debug mode")

	sdk.RegisterToProxy(globalConfig.RedisServer, globalConfig.RedisPassword, globalConfig.Id, globalConfig.LinkHost)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		sdk.UnRegisterToProxy(globalConfig.RedisServer, globalConfig.RedisPassword, globalConfig.Id)
		os.Exit(1)
	}()

	// var DownSubjectName string
	// var UpSubjectName string
	// var LinkName string

	// if mode == 1 {
	// 	DownSubjectName = "down-test-p0"
	// 	UpSubjectName = "up-p0-test"
	// 	LinkName = "test-link"
	// } else {
	// 	DownSubjectName = "down-test-p1"
	// 	UpSubjectName = "up-p1-test"
	// 	LinkName = "test-link"
	// }

	// cus, err := ctrl.ms.NewConsumer(DownSubjectName)
	// if err != nil {
	// 	panic(err)
	// }
	// pro, err := ctrl.ms.NewProducer(UpSubjectName)
	// if err != nil {
	// 	panic(err)
	// }
	ls, err := h2.NewLink(&link.Config{
		Id:             globalConfig.Id,
		Host:           globalConfig.LinkHost,
		Domain:         globalConfig.Domain,
		Target:         globalConfig.Target,
		Remote:         globalConfig.Remote,
		StorePath:      "",
		WindowSize:     100, //globalConfig.WindowSize,
		RecoverSupport: false,
	})
	if err != nil {
		panic(err)
	}

	ch, err := ls.NewChannel("test-link")
	if err != nil {
		panic(err)
	}
	ch.WaitInit()

	// sub := &subject{cus: cus, pro: pro, chn: chn, consumer_seq: 0}

	// go sub.slink()

	// s, err := nats.NewStream(mq.Config{conf.MqAddress, 1})
	// if err != nil {
	// 	panic(err)
	// }

	if globalConfig.Debug == 1 {
		go func() {
			for i := 0; i < math.MaxInt; i++ {
				ack := <-ch.Ack()
				// if ack != uint64(i) {
				// 	t.Fatalf("ack error: %d != %d", ack, i)
				// }
				_ = ack
				// t.Logf("ack %d", ack)
				hlog.Infof("send %d", i)
			}
		}()
		for i := 0; i < math.MaxInt; i++ {
			data := strconv.FormatUint(uint64(i), 10)
			ch.Send() <- link.Msg{Seq: uint64(i), Data: []byte(data)}
			// time.Sleep(time.Second)
		}
		// p0p, err := s.NewProducer("down-test-p0")
		// if err != nil {
		// 	panic(err)
		// }

		// for i := 0; i < math.MaxInt; i++ {
		// 	p0p.Push() <- []byte(fmt.Sprintf("%d", i))
		// 	time.Sleep(time.Millisecond)
		// 	// t.Logf("send %d", i)
		// }
	} else {
		for i := 0; i < math.MaxInt; i++ {
			msg := <-ch.Recv()
			seq, _ := strconv.ParseUint(string(msg), 10, 64)
			if seq != uint64(i) {
				hlog.Errorf("seq error:  %d != %d", seq, i)
				panic("error")
			}
			hlog.Infof("recv %d", i)
		}
		// c1c, err := s.NewConsumer("up-p1-test")
		// if err != nil {
		// 	panic(err)
		// }
		// for i := 0; i < math.MaxInt; i++ {
		// 	msg := <-c1c.Consume()
		// 	ci, err := strconv.ParseInt(string(msg.Data), 10, 64)
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// 	// if ci != int64(i) {
		// 	// 	t.Logf("expect %d, but %d", i, ci)
		// 	// 	// panic(fmt.Sprintf("expect %d, but %d", i, ci))
		// 	// }
		// 	t.Logf("recv %d %d", ci, i)
		// 	msg.Ack()
		// }
	}

}
