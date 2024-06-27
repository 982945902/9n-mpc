package main

import (
	"bang/mq"
	"bang/mq/nats"
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"
)

var p0Conf = &Config{
	Id:         "test",
	Domain:     "jd",
	Host:       "127.0.0.1:7000",
	LinkHost:   "127.0.0.1:9000",
	Target:     "cu",
	Remote:     "127.0.0.1:9001",
	MqAddress:  "127.0.0.1:4222",
	WindowSize: 1,
}

var p1Conf = &Config{
	Id:         "test",
	Domain:     "jd",
	Host:       "127.0.0.1:7001",
	LinkHost:   "127.0.0.1:9001",
	Target:     "cu",
	Remote:     "127.0.0.1:9000",
	MqAddress:  "127.0.0.1:4222",
	WindowSize: 1,
}

func run_test(t *testing.T, conf *Config, p0 bool) {
	ctrl, err := newControler(conf)
	if err != nil {
		panic(err)
	}

	go ctrl.Run()

	var req createSubjectReq
	if p0 {
		req = createSubjectReq{
			DownSubjectName: "down-test-p0",
			UpSubjectName:   "up-p0-test",
			LinkName:        "test-link",
		}
	} else {
		req = createSubjectReq{
			DownSubjectName: "down-test-p1",
			UpSubjectName:   "up-p1-test",
			LinkName:        "test-link",
		}
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}

	time.Sleep(2 * time.Second)
	_, err = http.Post(
		fmt.Sprintf("http://%s/bang/create", conf.Host),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		panic(err)
	}

	s, err := nats.NewStream(mq.Config{conf.MqAddress, 1})
	if err != nil {
		panic(err)
	}

	if p0 {
		p0p, err := s.NewProducer("down-test-p0")
		if err != nil {
			panic(err)
		}

		for i := 0; i < math.MaxInt; i++ {
			p0p.Push() <- []byte(fmt.Sprintf("%d", i))
			time.Sleep(time.Millisecond)
			// t.Logf("send %d", i)
		}
	} else {
		c1c, err := s.NewConsumer("up-p1-test")
		if err != nil {
			panic(err)
		}
		for i := 0; i < math.MaxInt; i++ {
			msg := <-c1c.Consume()
			ci, err := strconv.ParseInt(string(msg.Data), 10, 64)
			if err != nil {
				panic(err)
			}
			// if ci != int64(i) {
			// 	t.Logf("expect %d, but %d", i, ci)
			// 	// panic(fmt.Sprintf("expect %d, but %d", i, ci))
			// }
			t.Logf("recv %d %d", ci, i)
			msg.Ack()
		}
	}

	time.Sleep(time.Minute)
}

func TestMain(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		run_test(t, p0Conf, true)
	}()

	go func() {
		defer wg.Done()
		run_test(t, p1Conf, false)
	}()

	wg.Wait()
}
