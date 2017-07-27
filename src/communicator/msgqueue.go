package communicator

import (
	"communicator/mq/broker"
	"communicator/mq/client"
	"time"
	"fmt"
)

var brokerAddr string = ":11181"

type msgBroker struct {

}

func newMessageBroker() *msgBroker {
	return &msgBroker{}
}

func (b *msgBroker) Start() error {
	cfg := broker.NewDefaultConfig()
	cfg.Addr = brokerAddr

	app, err := broker.NewAppWithConfig(cfg)
	if err != nil {
		panic(err)
	}

	fmt.Println("broker start")
	go app.Run()
	return nil
}

func (b *msgBroker) Stop() error {
	return nil
}


type msgPublisher struct {
	c 			*client.Client
}

func newMsgPublisher() *msgPublisher {
	return &msgPublisher{
	}
}

func (mp *msgPublisher) Start() error {
	cfg := client.NewDefaultConfig()
	cfg.BrokerAddr = brokerAddr

	c, err := client.NewClientWithConfig(cfg)
	if err != nil {
		panic(err)
	}
	mp.c = c
	return nil
}

func (mp *msgPublisher) Stop() error {
	mp.c.Close()
	return nil
}


func (mp *msgPublisher) WaitPublish(channel string, key string, data interface{}) error {
	msg, err := serilize(key, data)
	if err != nil {
		err :=fmt.Errorf("serilize data err %v %v %v %v", channel, key, data, err)
		fmt.Println(err)
		return err
	}
	fmt.Println("PublishDirect publish ", channel, data, msg)
	_, err = mp.c.PublishDirect(channel, key, msg)
	return err
}

func (mp *msgPublisher) SendPublish(channel string, data interface{}) error {
	msg, err := serilize(channel, data)
	if err != nil {
		return fmt.Errorf("serilize data err %v %v %v", channel, channel, data)
	}
	fmt.Println("fanout publish ", channel, data, msg)
	_, err = mp.c.PublishFanout(channel, msg)
	return err
}

type msgConsumer struct {
	c 			*client.Client
}

func newMsgConsumer() *msgConsumer {
	return &msgConsumer{
	}
}

func (mc *msgConsumer) Start() error {
	cfg := client.NewDefaultConfig()
	cfg.BrokerAddr = brokerAddr
	c, err := client.NewClientWithConfig(cfg)
	if err != nil {
		panic(err)
	}
	mc.c = c
	return nil
}

func (mc *msgConsumer) Stop() error {
	mc.c.Close()
	return nil
}

func (mc *msgConsumer) WaitMessage(channel string, key string, t time.Duration) interface{} {
	conn, err := mc.c.Get()
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	ch, err := conn.Bind(channel, key, true)
	d := ch.WaitMsg(t)
	msg, err := deserilize("", key, d)
	if err != nil {
		fmt.Println("desierlize message error ", d, err)
		return nil
	}
	return msg
}

func (mc *msgConsumer) GetMessage(channel string, key string) interface {} {
	conn, err := mc.c.Get()
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	ch, err := conn.Bind(channel, key, true)
	d := ch.GetMsg()
	msg, err := deserilize("", key, d)
	if err != nil {
		fmt.Println("desierlize message error ", d, err)
		return nil
	}
	return msg
}

type brokerClient struct {
	*msgPublisher
	*msgConsumer
}

func newBrokerClient() *brokerClient {
	return &brokerClient{
		msgPublisher: newMsgPublisher(),
		msgConsumer: newMsgConsumer(),
	}
}

func (bc *brokerClient) Start() error {
	bc.msgPublisher.Start()
	bc.msgConsumer.Start()
	return nil
}

func (bc *brokerClient) Stop() error {
	bc.msgConsumer.Stop()
	bc.msgPublisher.Stop()
	return nil
}
