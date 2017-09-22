package main

import (
	"flag"
	"communicator/mq/broker"
	"communicator/mq/client"
	"time"
	"mylog"
)

// broker
var addr = flag.String("addr", "127.0.0.1:11181", "moonmq broker listen address")

// publisher
var queue = flag.String("queue", "test_queue", "queue want to bind")
var msg = flag.String("msg", "hello world", "msg to publish")

func testBroker() {
	mylog.Debug("test broker")
	flag.Parse()

	cfg := broker.NewDefaultConfig()
	cfg.Addr = *addr

	app, err := broker.NewAppWithConfig(cfg)
	if err != nil {
		panic(err)
	}

	app.Run()
}

func testPublisher() {
	cfg := client.NewDefaultConfig()
	cfg.BrokerAddr = *addr

	c, err := client.NewClientWithConfig(cfg)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	_, err = c.PublishFanout(*queue, []byte(*msg))
	if err != nil {
		panic(err)
	}
}

func testConsumer() {
	flag.Parse()

	cfg := client.NewDefaultConfig()
	cfg.BrokerAddr = *addr

	c, err := client.NewClientWithConfig(cfg)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	var conn *client.Conn
	conn, err = c.Get()
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	var ch *client.Channel
	ch, err = conn.Bind(*queue, "", true)

	msg := ch.GetMsg()
	println("get msg: ", string(msg))

}

func testMq() {
	go testBroker()

	time.Sleep(1*time.Second)

	testPublisher()
	testConsumer()

}