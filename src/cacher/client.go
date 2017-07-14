package cacher

import (
	"exportor/defines"
	"communicator"
	"github.com/garyburd/redigo/redis"
	"time"
	"log"
	"fmt"
	"reflect"
	"exportor/proto"
	"strconv"
)


type cacheClient struct {
	group 			string
	communicator 	defines.ICommunicatorClient
	ccConn 			redis.Conn
	notify 			chan interface{}
}

func newCacheClient(gr string) *cacheClient {
	return &cacheClient{
		group: gr,
		communicator: communicator.NewCommunicator(nil),
		notify: make(chan interface{}),
	}
}

func (cc *cacheClient) connectCacheServer() error {
	conn, err := redis.Dial("tcp", ":6379", redis.DialReadTimeout(1*time.Second), redis.DialWriteTimeout(1*time.Second))
	if err != nil {
		return err
	}
	cc.ccConn = conn
	return nil
}

func (cc *cacheClient) Start() {

	if err := cc.connectCacheServer(); err != nil {
		log.Fatalln("connect cache server err :", err)
	}

	cc.communicator.JoinChanel("dbLoadFinishChannel", false, func(data []byte) {
		cc.notify <- data
	})

	go func() {
		select {
		case d := <- cc.notify:
			fmt.Println("d isl", d)
		}
	}()
}

func (cc *cacheClient) Stop() {

}

// ICacheClient
func (cc *cacheClient) SetCacheNotify(notify defines.ICacheNotify) {

}

func (cc *cacheClient) GetUserId(name string) (uint32,error) {
	id, err := redis.Int(cc.ccConn.Do("hget", "uids."+name+":Uid"))
	return uint32(id), err
}

func (cc *cacheClient) GetUserInfo(name string, user *proto.CacheUser) error {

	uid, err := cc.GetUserId(name)
	if err != nil {
		fmt.Println("get user id error", name)
		return err
	}

	values, err := redis.Values(cc.ccConn.Do("HGETALL", "user."+strconv.Itoa(int(uid))))
	if err != nil {
		fmt.Println("get user info error", name, uid)
		return err
	}

	if err := redis.ScanStruct(values, user); err != nil {
		return err
	}

	return nil
}

func (cc *cacheClient) GetUserProps(uid uint32, props string) (interface{}, error) {

}

// ICacheLoader