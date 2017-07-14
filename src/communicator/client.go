package communicator

import (
	"github.com/garyburd/redigo/redis"
	"exportor/defines"
	"time"
	"fmt"
)

type communicator struct {
	opt 		*defines.CommunicatorOption
}

func newCommunicator(opt *defines.CommunicatorOption) *communicator {
	return &communicator{
		opt: opt,
	}
}

func (cm *communicator) JoinChanel(chanel string, reg bool, cb defines.CommunicatorCb) error {
	c, err := cm.connectServer()
	if err != nil {
		return err
	}

	psc := redis.PubSubConn{Conn:c}
	if reg {
		psc.PSubscribe(chanel)
	} else {
		psc.Subscribe(chanel)
	}

	go func() {
		defer func() {
			fmt.Println("psc close ", chanel)
			psc.Close()
		}()

		for {
			switch n := psc.Receive().(type) {
			case redis.Message:
				fmt.Printf("Message: %s %s\n", n.Channel, n.Data)
				cb(n.Data)
			case redis.PMessage:
				fmt.Printf("PMessage: %s %s %s\n", n.Pattern, n.Channel, n.Data)
			case redis.Subscription:
				fmt.Printf("Subscription: %s %s %d\n", n.Kind, n.Channel, n.Count)
			case error:
				fmt.Printf("error: %v\n", n)
				return
			}
		}
	}()

	return nil
}

func (cm *communicator) connectServer() (redis.Conn, error) {
	conn, err := redis.Dial("tcp", ":6379", redis.DialReadTimeout(1*time.Second), redis.DialWriteTimeout(1*time.Second))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (cm *communicator) Notify(chanel string, value ... interface{}) error {
	//go func() {
		c, err := cm.connectServer()
		if err != nil {
			fmt.Println("connect server ", err)
			return nil
		}
		defer c.Close()
		r, err := c.Do("PUBLISH", chanel, value)
	fmt.Println("notify ... ", r, err)
		if err != nil {
			return err
		}

	//}()
	return nil
}