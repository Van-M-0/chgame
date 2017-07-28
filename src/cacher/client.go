package cacher

import (
	"exportor/defines"
	"github.com/garyburd/redigo/redis"
	"time"
	"log"
	"fmt"
	"exportor/proto"
	"dbproxy/table"
	"errors"
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
		//communicator:  communicator.NewCommunicator(nil),
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

	/*
	cc.communicator.JoinChanel("dbLoadFinishChannel", false, defines.WaitChannelInfinite, func(data []byte) {
		cc.channelNotify <- data
	})
	*/

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

func (cc *cacheClient) command(commandName string, args ...interface{}) (reply interface{}, err error) {
	fmt.Println("")
	fmt.Println("redis command> ", commandName, args)
	fmt.Println("")
	return cc.ccConn.Do(commandName, args...)
}

func (cc *cacheClient) GetUserId(account string) (uint32,error) {
	id, err := redis.Int(cc.command("get", accountId(account)))
	return uint32(id), err
}

func (cc *cacheClient) GetUserInfo(account string, user *proto.CacheUser) error {

	uid, err := cc.GetUserId(account)

	if err == redis.ErrNil {
		return nil
	}

	if err != nil {
		fmt.Println("get user id error", err, account)
		return err
	}

	values, err := redis.Values(cc.command("HGETALL", users(int(uid))))
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
	values, err := redis.Values(cc.command("HGETALL", users(int(uid))))
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

	if _, err := cc.command("set", accountId(userInfo.Account), userInfo.Userid); err != nil {
		fmt.Println("set account info err ", accountId(userInfo.Account), userInfo.Account, userInfo.Userid)
	}

	//todo: user int <-> int32
	cu := &proto.CacheUser{
		Account: userInfo.Account,
		Name: userInfo.Name,
		Uid: int(userInfo.Userid),
	}

	if _, err := cc.command("hmset", redis.Args{users(cu.Uid)}.AddFlat(cu)...); err != nil {
		log.Fatal(err)
	}

	return nil
}

func (cc *cacheClient) UpdateUserInfo(uid uint32, prop string, value interface{}) bool {
	var (
		err error
		reply interface{}
	)
	if prop == "roomid" {
		reply, err = cc.command("hset", users(int(uid)), "roomid", value)
	} else if prop == "gold" {
		g := value.(int64)
		if g > 0 {
			reply, err = cc.command("hset", users(int(uid)), "gold", g)
		}
	}
	if err != nil {
		fmt.Println(reply, err)
		return false
	}
	return true
}

func (cc *cacheClient) SetServer(server *proto.CacheServer) error {
	serverKeys := servers(server.Id)
	if _, err := cc.command("hmset", redis.Args{serverKeys}.AddFlat(server)...); err != nil {
		log.Fatal(err)
	}
	return nil
}

func (cc *cacheClient) GetServers() ([]*proto.CacheServer, error) {
	var servers []*proto.CacheServer

	skeys, err := redis.Strings(cc.ccConn.Do("keys", serversPattern()))
	if err != nil {
		return nil, err
	}

	for _, key := range skeys {
		values, err := redis.Values(cc.command("hgetall", key))
		if err != nil {
			fmt.Println("get server err ", key, err)
			continue
		}

		var ser proto.CacheServer
		if err := redis.ScanStruct(values, ser); err != nil {
			fmt.Println("get server scan values error", err)
			continue
		}

		servers = append(servers, &ser)
	}

	return servers, nil
}

func (cc *cacheClient) UpdateServer(ser *proto.CacheServer) error {
	if ser.Id == 0 {
		return errors.New("id is empty")
	}
	key := servers(ser.Id)
	if _, err := cc.command("hmset", redis.Args{key}.AddFlat(ser)...); err != nil {
		log.Fatal(err)
	}
	return nil
}

func (cc *cacheClient) FlushAll() {
	if cc.group != "dbproxy" {
		return
	}
	cc.ccConn.Do("flushall")
}

// ICacheLoader