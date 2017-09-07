package dbproxy

import (
	"time"
	"exportor/defines"
	"cacher"
	"fmt"
	"runtime/debug"
	"dbproxy/table"
)

type user_item_key struct {
	uid 		uint32
	iid 		uint32
}

type dataSaver struct {
	t 			map[string]time.Duration
	cc 			defines.ICacheClient
	ch 			chan interface{}
	dc 			*dbClient

	lsUsers 		map[uint32]*table.T_Users
	lsUserItems		map[user_item_key]*table.T_UserItem
}

func newDataSaver(dc *dbClient) *dataSaver {
	ds := &dataSaver{}
	ds.cc = cacher.NewCacheClient("saver")
	ds.t = make(map[string]time.Duration)
	ds.ch = make(chan interface{})
	ds.dc = dc
	ds.lsUsers = make(map[uint32]*table.T_Users)
	ds.lsUserItems = make(map[user_item_key]*table.T_UserItem)
	return ds
}

func (ds *dataSaver) safeRoutine(fn func()) {
	safeCall := func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("------save data error------")
				fmt.Println(err)
				debug.PrintStack()
				fmt.Println("--------------------------")
			}
		}()
		fn()
	}

	go func() {
		safeCall()
	}()
}

func (ds *dataSaver) start() {
	ds.cc.Start()

	preList := []*table.T_UserItem{}
	ds.dc.db.Find(&preList)
	for _, i := range preList {
		ds.lsUserItems[user_item_key{
			uid: i.Userid,
			iid: i.Itemid,
		}] = &table.T_UserItem{
			Userid: i.Userid,
			Itemid: i.Itemid,
			Count: i.Count,
		}
	}

	timeoutFn := func(tm time.Duration, fn func()) {
		for {
			t := time.NewTimer(time.Minute * tm)
			select {
			case <- t.C:
				fn()
			}
		}
	}
	ds.safeRoutine(func() {
		timeoutFn(2, ds.collect)
	})
	ds.safeRoutine(ds.save)
}

func (ds *dataSaver) stop() {
	ds.cc.Stop()
}

func (ds *dataSaver) collect() {
	getUsers := func() {
		users, _ := ds.cc.GetAllUsers()
		if users != nil {
			l := []*table.T_Users{}
			for _, user := range users {
				//fmt.Println("collect users", *user)
				l = append(l, &table.T_Users{
					Userid: uint32(user.Uid),
					Account: user.Account,
					OpenId: user.Openid,
					Name: user.Name,
					Sex: user.Sex,
					Headimg: user.HeadImg,
					Diamond: uint32(user.Diamond),
					RoomCard: uint32(user.RoomCard),
					Gold: user.Gold,
					Score: uint32(user.Score),
					Roomid: uint32(user.RoomId),
				})
			}
			ds.ch <- l
		}
	}

	getUserItems := func() {
		userItems, _ := ds.cc.GetAllUserItem()
		if userItems != nil {
			l := []*table.T_UserItem{}
			for _, item := range userItems {
				l = append(l, &table.T_UserItem{
					Itemid: uint32(item.Id),
					Userid: uint32(item.UserId),
					Count: item.Count,
				})
			}
			ds.ch <- l
		}
	}

	getUsers()
	getUserItems()
}

func (ds *dataSaver) save() {
	for {
		select {
		case d := <- ds.ch:
			switch data := d.(type) {
			case []*table.T_Users:
				for _, u := range data {
					if ou, ok := ds.lsUsers[u.Userid]; ok {
						if *ou != *u {
							ds.dc.db.Table("t_users").Where("userid = ?", u.Userid).Updates(map[string]interface{}{
								"headimg": 	u.Headimg,
								"level": 	u.Level,
								"exp":		u.Exp,
								"diamond":	u.Diamond,
								"gold":		u.Gold,
								"score":	u.Score,
							})
							ds.lsUsers[u.Userid] = u
						}
					} else {
						ds.dc.db.Table("t_users").Where("userid = ?", u.Userid).Updates(map[string]interface{}{
							"headimg": 	u.Headimg,
							"level": 	u.Level,
							"exp":		u.Exp,
							"diamond":	u.Diamond,
							"gold":		u.Gold,
							"score":	u.Score,
							"roomid":	u.Roomid,
						})
						ds.lsUsers[u.Userid] = u
					}
				}
			case []*table.T_UserItem:
				for _, item := range data {
					key := user_item_key{uid: item.Userid, iid: item.Itemid}
					if oi, ok := ds.lsUserItems[key]; ok {
						if *oi != *item {
							if item.Count != 0 {
								ds.lsUserItems[key] = item
								ds.dc.db.Model(item).Where("userid = ? and itemid = ?", item.Userid, item.Itemid).Update("count", item.Count)
							} else {
								ds.dc.db.Where("userid = ? and itemid = ?", item.Userid, item.Itemid).Delete(item)
								delete(ds.lsUserItems, key)
							}
						}
					} else {
						if item.Count != 0 {
							ds.lsUserItems[key] = item
							ds.dc.db.Save(item)
						}
					}
				}
			default:
				fmt.Println("no handler data ", d)
			}
		}
	}
}