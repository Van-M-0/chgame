package cacher

import (
	"exportor/defines"
	"communicator"
	"github.com/garyburd/redigo/redis"
	"time"
	"log"
	"fmt"
	"exportor/proto"
	"dbproxy/table"
)


type cacheClient struct {
	group         string
	communicator  defines.ICommunicatorClient
	ccConn        redis.Conn
	channelNotify chan interface{}
}

func newCacheClient(gr string) *cacheClient {
	return &cacheClient{
		group:         gr,
		communicator:  communicator.NewCommunicator(nil),
		channelNotify: make(chan interface{}),
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

func (cc *cacheClient) Start() error {

	if err := cc.connectCacheServer(); err != nil {
		log.Fatalln("connect cache server err :", err)
	}

	cc.communicator.JoinChanel("dbLoadFinishChannel", false, func(data []byte) {
		cc.channelNotify <- data
	})

	go func() {
		select {
		case d := <- cc.channelNotify:
			fmt.Println("d isl", d)
		}
	}()

	return nil
}

func (cc *cacheClient) Stop() error {
	return nil
}

// ICacheClient
func (cc *cacheClient) SetCacheNotify(notify defines.ICacheNotify) {

}

func (cc *cacheClient) GetUserId(account string) (uint32,error) {
	id, err := redis.Int(cc.ccConn.Do("hget", accountId(account)))
	return uint32(id), err
}

func (cc *cacheClient) GetUserInfo(account string, user *proto.CacheUser) error {

	uid, err := cc.GetUserId(account)
	if err != nil {
		fmt.Println("get user id error", account)
		return err
	}

	values, err := redis.Values(cc.ccConn.Do("HGETALL", users(int(uid))))
	if err != nil {
		fmt.Println("get user info error", account, uid)
		return err
	}

	if err := redis.ScanStruct(values, user); err != nil {
		return err
	}

	return nil
}

func (cc *cacheClient) GetUserInfoById(uid uint32, user *proto.CacheUser) error {
	values, err := redis.Values(cc.ccConn.Do("HGETALL", users(int(uid))))
	if err != nil {
		fmt.Println("get user info error", uid)
		return err
	}

	if err := redis.ScanStruct(values, user); err != nil {
		return err
	}

	return nil
}

func (cc *cacheClient) SetUserInfo(d interface{}, dbRet bool) error {
	userInfo := d.(*table.T_Users)

	//todo: user int <-> int32
	cu := &proto.CacheUser{
		Account: userInfo.Account,
		Name: userInfo.Name,
		Uid: int(userInfo.Userid),
	}

	if _, err := cc.ccConn.Do("HMSET", redis.Args{users(cu.Uid)}.AddFlat(cu)...); err != nil {
		log.Fatal(err)
	}

	return nil
}


// ICacheLoader