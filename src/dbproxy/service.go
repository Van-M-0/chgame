package dbproxy

import (
	"sync"
	"dbproxy/table"
	"fmt"
	"exportor/defines"
	"cacher"
	"strconv"
	"time"
	"exportor/proto"
)

type DBService struct {
	lock 		sync.RWMutex
	db 			*dbClient
	cc 			defines.ICacheClient
}

func newDbService(db *dbClient) *DBService {
	service := &DBService{}
	service.db = db
	service.cc = cacher.NewCacheClient("DBService")
	return service
}

func (service *DBService) start() {
	service.cc.Start()
}

func (service *DBService) UserLogin(req *proto.DbUserLoginArg, res *proto.DbUserLoginReply) error {
	var userInfo table.T_Users
	service.lock.Lock()
	ret := service.db.GetUserInfo(req.Acc, &userInfo)
	service.lock.Unlock()

	fmt.Println("user login ", req)
	if ret == true {
		err := service.cc.SetUserInfo(&userInfo, ret)
		if err != nil {
			fmt.Println("set cache user error ", err)
			res.Err = "cache"
		} else {
			res.Err = "ok"
		}
	} else {
		res.Err = "notexists"
	}
	fmt.Println("user login ", res)
	return nil
}

func (service *DBService) CreateAccount(req *proto.DbCreateAccountArg, res *proto.DbCreateAccountReply) error {
	var user table.T_Users
	var userSuccess *table.T_Users
	fmt.Println("create account ", req)
	service.lock.Lock()
	ret := service.db.GetUserInfo(req.UserName, &user)
	service.lock.Unlock()
	if !ret {
		acc := "acc_" + strconv.Itoa(int(time.Now().Unix()))
		pwd := "123456"
		r := service.db.AddAccountInfo(&table.T_Accounts{
			Account: acc,
			Password: pwd,
		})
		if r {
			userSuccess = &table.T_Users{
				Account: acc,
				Name: req.UserName,
				Level: 1,
				Exp: 0,
				Gold: 100,
				RoomCard: 1,
			}
			r = service.db.AddUserInfo(userSuccess)
			res.Err = "ok"
			res.Acc = acc
		} else {
			res.Err = "cache"
		}
	} else {
		res.Err = "exists"
	}
	fmt.Println("create account ", res)
	return nil
}

func (service *DBService) LoadNotice(req *proto.MsLoadNoticeArg, res *proto.MsLoadNoticeReply) error {
	var notice []*table.T_Notice
	service.db.db.Find(&notice)
	for _, n := range notice {
		res.Notices = append(res.Notices, &proto.NoticeItem{
			Id: n.Index,
			Kind: n.Kind,
			Content: n.Content,
			StartTime: n.Starttime,
			FinishTime: n.Finishtime,
			Counts: n.Playcount,
			PlayTime: n.Playtime,
		})
	}
	return nil
}

func (service *DBService) LoadMallItem(req *proto.MsLoadNoticeArg, res *proto.MsLoadMallItemListReply) error {
	var malls []*table.T_MallItem
	service.db.db.Find(&malls)
	for _, n := range malls {
		res.Malls = append(res.Malls, &proto.MallItem{
			Id: n.Itemid,
			Name: n.Itemname,
			Category: n.Category,
			BuyValue: n.Buyvalue,
			Nums: n.Nums,
			BuyLimt: n.Limit,
		})
	}
	return nil
}