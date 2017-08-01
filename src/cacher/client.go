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
		fmt.Println("cache user not in")
		return err
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
		fmt.Println("scan stuct error ", err)
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
		return err
	}

	//todo: user int <-> int32
	cu := &proto.CacheUser{
		Account: userInfo.Account,
		Name: userInfo.Name,
		Uid: int(userInfo.Userid),
		Diamond: int(userInfo.Diamond),
		RoomId: int(userInfo.Roomid),
		Gold: int64(userInfo.Gold),
		Score: int(userInfo.Score),
	}

	if _, err := cc.command("hmset", redis.Args{users(cu.Uid)}.AddFlat(cu)...); err != nil {
		log.Fatal(err)
	}

	return nil
}

func (cc *cacheClient) SetUserCidUserId(uid uint32, userId int) error {
	if _, err := cc.command("set", ciduserid(uid), userId); err != nil {
		fmt.Println("set account info err ", ciduserid(uid), uid, userId)
		return err
	}
	return nil
}

func (cc *cacheClient) GetUserCidUserId(uid uint32) int {
	userId, err := redis.Int(cc.command("get", ciduserid(uid)))
	if err != nil {
		return -1
	} else {
		return userId
	}
}

func (cc *cacheClient) UpdateUserInfo(uid uint32, prop int, value interface{}) bool {
	var (
		err error
		reply interface{}
	)
	if prop == defines.PpRoomId  {
		reply, err = cc.command("hset", users(int(uid)), "roomid", value)
	} else if prop ==  defines.PpGold {
		g := value.(int64)
		if g > 0 {
			reply, err = cc.command("hset", users(int(uid)), "gold", g)
		}
	} else if prop == defines.PpScore {
		g := value.(int64)
		if g > 0 {
			reply, err = cc.command("hset", users(int(uid)), "gold", g)
		}
	} else {
		return false
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

func (cc *cacheClient) NoticeOperation(notice *[]*proto.CacheNotice, op string) error {
	if op == "update" {
		for _, n := range *notice {
			if _, err := cc.command("hmset", redis.Args{notices(n.Id)}.AddFlat(n)...); err != nil {
				fmt.Println("set notices error", err)
			}
		}
	} else if op == "del" {
		ids := []string{}
		for _, n := range *notice {
			ids = append(ids, notices(n.Id))
		}
		if _, err := cc.command("del", redis.Args{}.AddFlat(ids)...); err != nil {
			fmt.Println("set notices error", err)
		}
	} else if op == "getall" {
		nids, err := redis.Strings(cc.ccConn.Do("keys", noticesPattern()))
		if err != nil {
			return nil
		}

		for _, key := range nids {
			values, err := redis.Values(cc.command("hgetall", key))
			if err != nil {
				fmt.Println("get notice err ", key, err)
				continue
			}

			var ns proto.CacheNotice
			if err := redis.ScanStruct(values, &ns); err != nil {
				fmt.Println("get notice scan values error", err)
				continue
			}

			fmt.Println("ns is .. ", ns)

			*notice = append(*notice, &ns)
		}
	}
	return nil
}



// ICacheLoader