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
	"sort"
	"strings"
	"strconv"
	"reflect"
)

const (
	RecordTimeOut = 60 * 60 * 24
	MaxUserRecord = 500
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
		Openid: userInfo.OpenId,
		Uid: int(userInfo.Userid),
		Sex: userInfo.Sex,
		Name: userInfo.Name,
		HeadImg: userInfo.Headimg,
		Diamond: int(userInfo.Diamond),
		RoomCard: int(userInfo.RoomCard),
		Gold: int64(userInfo.Gold),
		Score: int(userInfo.Score),
		RoomId: int(userInfo.Roomid),
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

func (cc *cacheClient) DelUserCidUserId(uid uint32) {
	if _, err := cc.command("del", ciduserid(uid)); err != nil {
		fmt.Println("del user cid user error", err)
	}
}

func (cc *cacheClient) UpdateUserInfo(uid uint32, prop int, value interface{}) bool {

	getName := func() string {
		if prop == defines.PpDiamond {
			return "diamond"
		} else if prop == defines.PpGold {
			return "gold"
		} else if prop == defines.PpScore {
			return "score"
		} else if prop == defines.PpRoomId {
			return "roomid"
		}
		return "error"
	}

	keyName := getName()
	if keyName == "error" {
		return false
	}

	updateIntProp := func(key int, val interface{}) bool {
		if reflect.TypeOf(val).Kind() == reflect.Uint32 {
			val = int(val.(uint32))
		}
		if v, ok := val.(int); ok {
			if v > 0 {
				reply, err := cc.command("hset", users(int(uid)), keyName, v)
				if err != nil {
					fmt.Println("UpdateUserInfo set cache error ", uid, key, val, reply, err)
				} else {
					return true
				}
			} else {
				fmt.Println("UpdateUserInfo update cache value ?? ", key, v)
			}
		} else {
			fmt.Println("UpdateUserInfo value not right", prop, value)
		}
		return false
	}

	updateInt64Prop := func(key int, val interface{}) bool {
		if v, ok := val.(int64); ok {
			if v > 0 {
				reply, err := cc.command("hset", users(int(uid)), keyName, v)
				if err != nil {
					fmt.Println("UpdateUserInfo set cache error ", uid, key, val, reply, err)
				} else {
					return true
				}
			} else {
				fmt.Println("UpdateUserInfo update cache value ?? ", key, v)
			}
		} else {
			fmt.Println("UpdateUserInfo value not right", prop, value)
		}
		return false
	}

	if prop == defines.PpRoomId || prop == defines.PpScore || prop == defines.PpDiamond || prop == defines.PpRoomCard {
		return updateIntProp(prop, value)
	} else if prop == defines.PpGold {
		return updateInt64Prop(prop, value)
	}

	fmt.Println("UpdateUserInfo udpate prop not exists ", prop)
	return false
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

	skeys, err := redis.Strings(cc.command("keys", serversPattern()))
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
		nids, err := redis.Strings(cc.command("keys", noticesPattern()))
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

func (cc *cacheClient) UpdateSingleItem(userid uint32, flag int, id uint32, count int) error {
	if flag == 2 {
		cc.command("del", useritems(userid, id))
	} else if flag == 3 || flag == 1 {
		citem := proto.CacheUserItem{
			UserId: int(userid),
			Id: int(id),
			Count: count,
		}
		if _, err := cc.command("hmset", redis.Args{useritems(userid, id)}.AddFlat(&citem)...); err != nil {
			fmt.Println("set notices error", err)
			return err
		}
	}
	return nil
}

func(cc *cacheClient) UpdateUserItems(userid uint32, items []proto.UserItem) error {
	for _, item := range items {
		citem := proto.CacheUserItem{
			UserId: int(userid),
			Id: int(item.ItemId),
			Count: item.Count,
		}
		if _, err := cc.command("hmset", redis.Args{useritems(userid, item.ItemId)}.AddFlat(&citem)...); err != nil {
			fmt.Println("set notices error", err)
		}
	}
	return nil
}

func(cc *cacheClient) GetUserItems(userid uint32) ([]*proto.UserItem, error) {

	skeys, err := redis.Strings(cc.command("keys", alluseritems(userid)))
	if err != nil {
		return nil, err
	}

	items := []*proto.UserItem{}

	for _, key := range skeys {
		values, err := redis.Values(cc.command("hgetall", key))
		if err != nil {
			fmt.Println("get user items err ", key, err)
			continue
		}

		var i proto.CacheUserItem
		if err := redis.ScanStruct(values, &i); err != nil {
			fmt.Println("get user items scan values error", err)
			continue
		}

		items = append(items, &proto.UserItem{
			ItemId: uint32(i.Id),
			Count: i.Count,
		})
	}

	return items, nil
}


func (cc *cacheClient) GetAllUsers() ([]*proto.CacheUser, error) {
	if cc.group != "saver" {
		return nil, nil
	}

	keys, err := redis.Strings(cc.command("keys", allUsers()))
	if err != nil {
		return nil, err
	}

	l := []*proto.CacheUser{}
	for _, key := range keys {
		values, err := redis.Values(cc.command("hgetall", key))
		if err != nil {
			fmt.Println("get user err ", key, err)
			continue
		}

		var i proto.CacheUser
		if err := redis.ScanStruct(values, &i); err != nil {
			fmt.Println("get user scan values error", err)
			continue
		}

		l = append(l, &i)
	}

	return l, nil
}

func (cc *cacheClient) GetAllUserItem() ([]*proto.CacheUserItem, error) {
	if cc.group != "saver" {
		return nil, nil
	}

	keys, err := redis.Strings(cc.command("keys", allItems()))
	if err != nil {
		return nil, err
	}

	l := []*proto.CacheUserItem{}
	for _, key := range keys {
		values, err := redis.Values(cc.command("hgetall", key))
		if err != nil {
			fmt.Println("get user items err ", key, err)
			continue
		}

		var i proto.CacheUserItem
		if err := redis.ScanStruct(values, &i); err != nil {
			fmt.Println("get user items scan values error", err)
			continue
		}

		l = append(l, &i)
	}

	return l, nil
}

func (cc *cacheClient) SaveGameRecord(head, content []byte) int {
	id, _ := redis.Int(cc.command("incr", recordId()))
	id += 100000
	if _, err := cc.command("set", recordHead(id), head, "ex", RecordTimeOut); err != nil {
		fmt.Println("save record head error ", err)
	}
	if _, err := cc.command("set", recordContent(id), content, "ex", RecordTimeOut); err != nil {
		fmt.Println("save record content error ", err)
	}
	return id
}

func (cc *cacheClient) SaveUserRecord(userId, recordId int) error {
	keys, _ := redis.Strings(cc.command("keys", userAllRecord(userId)))
	if len(keys) > MaxUserRecord {
		sort.Strings(keys)
		cc.command("del", keys[0])
	}
	if _, err := cc.command("set", userRecord(userId, recordId), recordId, "ex", RecordTimeOut); err != nil {
		fmt.Println("save user record error", err)
	}
	return nil
}

func(cc *cacheClient) GetGameRecordHead(userId int) (map[int][]byte, error) {
	keys, err := redis.Strings(cc.command("keys", userAllRecord(userId)))
	if err != nil {
		return nil, err
	}
	sort.Strings(keys)

	hrids := []string{}
	for _, key := range keys {
		x := strings.Split(key, ".")
		i, _ := strconv.Atoi(x[len(x)-1])
		hrids = append(hrids, recordHead(i))
	}

	m := make(map[int][]byte)
	heads, err := redis.ByteSlices(cc.command("mget", redis.Args{}.AddFlat(hrids)...))
	if err != nil {
		return nil, err
	}

	for i, key := range keys {
		x := strings.Split(key, ".")
		id, err := strconv.Atoi(x[len(x)-1])
		fmt.Println("inner ", i, id)
		if err == nil {
			m[id] = heads[i]
		}
	}

	return m, nil
}
func(cc *cacheClient) GetGameRecordContent(id int) ([]byte, error) {
	return redis.Bytes(cc.command("get", recordContent(id)))
}

func (cc *cacheClient) Scripts(args ...interface{}) {
	cc.ccConn.Do("eval", args...)
}
