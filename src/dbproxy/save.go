package dbproxy

import (
	"time"
	"exportor/defines"
	"cacher"
	"fmt"
	"runtime/debug"
	"dbproxy/table"
)

type dataSaver struct {
	t 			map[string]time.Duration
	cc 			defines.ICacheClient
	ch 			chan interface{}
	dc 			*dbClient
}

func newDataSaver(dc *dbClient) *dataSaver {
	ds := &dataSaver{}
	ds.cc = cacher.NewCacheClient("saver")
	ds.t = make(map[string]time.Duration)
	ds.dc = dc
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

	timeoutFn := func(tm time.Duration, fn func()) {
		t := time.NewTimer(time.Second * tm)
		for {
			select {
			case <- t.C:
				fn()
			}
		}
	}
	ds.safeRoutine(func() {
		timeoutFn(2, ds.collect)
	})
	ds.safeRoutine(func() {
		timeoutFn(2, ds.save)
	})
}

func (ds *dataSaver) stop() {
	ds.cc.Stop()
}

func (ds *dataSaver) collect() {
	getUsers := func() {
		users, _:= ds.cc.GetAllUsers()
		if users != nil {
			l := []*table.T_Users{}
			for _, user := range users {
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
					Userid: uint32(item.UserId),
					Count: item.Count,
					Itemid: uint32(item.Id),
				})
			}
			ds.ch <- l
		}
	}

	getUsers()
	getUserItems()
}

func (ds *dataSaver) save() {
	fmt.Println("saver start")
	for {
		select {
		case d := <- ds.ch:
			switch data := d.(type) {
			case []*table.T_Users:
				for _, u := range data {
					ds.dc.db.Save(u)
				}
			case []*table.T_UserItem:
				for _, u := range data {
					ds.dc.db.Save(u)
				}
			}
		}
	}
}