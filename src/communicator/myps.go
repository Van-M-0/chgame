package communicator

import (
	"github.com/garyburd/redigo/redis"
	"time"
	"fmt"
)

type producer struct {
	conn 	redis.Conn
}

type consumer struct {
	conn 	redis.Conn
}

func connectServer() (redis.Conn, error) {
	conn, err := redis.Dial("tcp", ":6379", redis.DialReadTimeout(1*time.Second), redis.DialWriteTimeout(1*time.Second))
	if err != nil {
		fmt.Println("connect server error ", err)
		return nil, err
	}
	fmt.Println("connect redis success")
	return conn, nil
}

func (p *producer) Start() error {
	var err error
	p.conn, err = connectServer()
	if err != nil {
		return err
	}
	return nil
}

func (p *producer) Stop() error {
	p.conn.Close()
	return nil
}

func (p *producer) WaitPublish(channel string, key string, data interface{}) error {
	panic("not suported")
	return nil
}

func (p *producer) SendPublish(channel string, data interface{}) error {
	_, err := p.conn.Do("PUBLISH", channel, data)
	return err
}

func (c *consumer) Start() error {
	return nil
}

func (c *consumer) Stop() error {
	return nil
}

func (c *consumer) WaitMessage(channel string, key string, t time.Duration) interface{} {
	conn, err := c.getConn()
	if err != nil {
		return err
	}
	psc := redis.PubSubConn{Conn: conn}
	psc.Subscribe(key)

	defer func() {
		psc.Unsubscribe(key)
		psc.Close()
	}()

	read := func() chan []byte {
		var r chan []byte
		switch n := psc.Receive().(type) {
		case redis.Message:
			fmt.Printf("Message: %s %s\n", n.Channel, n.Data)
			r <- n.Data
		case error:
			fmt.Printf("error: %v\n", n)
		}
		return r
	}

	select {
		case <- time.After(time.Duration(t) * time.Second):
			fmt.Println("time out for channel ", channel)
			return nil
		case data := <- read():
			msg, err := deserilize(channel, key, data)
			if err != nil {
				fmt.Println("desierlize message error ", data, err)
				return nil
			}
			return msg
	}

	return nil
}

func (c *consumer) GetMessage(channel string, key string) interface {} {
	fmt.Println("get message ........")
	conn, err := c.getConn()
	if err != nil {
		return err
	}
	psc := redis.PubSubConn{Conn: conn}
	psc.Subscribe(key)

	defer func() {
		psc.Unsubscribe(key)
		psc.Close()
	}()

	read := func() []byte {
		var ch chan []byte
		switch n := psc.Receive().(type) {
		case redis.Message:
			fmt.Printf("Message: %s %s\n", n.Channel, n.Data)
			ch <- n.Data
		case redis.PMessage:
			fmt.Printf("PMessage: %s %s %s\n", n.Pattern, n.Channel, n.Data)
		case redis.Subscription:
			fmt.Printf("Subscription: %s %s %d\n", n.Kind, n.Channel, n.Count)
		case error:
			fmt.Printf("error: %v\n", n)
		}
		return <- ch
	}

	data := read()
	if data == nil {
		fmt.Println("read data nil ..........")
		return nil
	}

	msg, err := deserilize(channel, key, read())
	if err != nil {
		fmt.Println("desierlize message error ", data, err)
		return nil
	}
	return msg
}

func (c *consumer) getConn() (redis.Conn, error){
	return connectServer()
}

func newProduer() *producer {
	return &producer{}
}

func newConsumer() *consumer {
	return &consumer{}
}
