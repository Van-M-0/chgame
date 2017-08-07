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
			var itemlist []table.T_UserItem
			service.db.db.Find(&itemlist).Where("userid = ?",  userInfo.Userid)
			for _, item := range itemlist {
				res.UserItemList = append(res.UserItemList, proto.UserItem {
					ItemId: item.Itemid,
					Count: item.Count,
				})
			}
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
	ret := service.db.GetUserInfo(req.Acc, &user)
	service.lock.Unlock()
	if !ret {
		name := "name_" + strconv.Itoa(int(time.Now().Unix()))
		pwd := "123456"
		r := service.db.AddAccountInfo(&table.T_Accounts{
			Account: req.Acc,
			Password: pwd,
		})
		if r {
			userSuccess = &table.T_Users{
				Account: req.Acc,
				Name: name,
				Level: 1,
				Exp: 0,
				Gold: 100,
				Diamond: 10,
				RoomCard: 1,
				Score: 0,
			}
			r = service.db.AddUserInfo(userSuccess)
			res.Err = "ok"
			res.Acc = req.Acc
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
	/*
	var items []*table.T_ItemConfig
	service.db.db.Find(&items)
	var itemarea []*table.T_ItemArea
	service.db.db.Find(&itemarea)

	for _, n := range items {
		if n.Sell == 1 {
			res.Malls = append(res.Malls, &proto.MallItem{
				Id:       int(n.Itemid),
				Name:     n.Itemname,
				Category: n.Category,
				BuyValue: n.Buyvalue,
				Nums:     n.Nums,
			})
		}
	}
	*/
	return nil
}

func (service *DBService) LoadItemConfig(req *proto.MsLoadItemConfigArg, res *proto.MsLoadItemConfigReply) error {
	var items []*table.T_ItemConfig
	service.db.db.Find(&items)
	for _, item := range items {
		res.ItemConfigList = append(res.ItemConfigList, proto.ItemConfig{
			Itemid: item.Itemid,
			Itemname: item.Itemname,
			Category: item.Category,
			Nums: item.Nums,
			Sell: item.Sell,
			Buyvalue: item.Buyvalue,
			GameKind: item.GameKind,
			Description: item.Description,
		})
	}
	return nil
}

func (service *DBService) LoadUserRank(req *proto.MsLoadUserRankArg, res *proto.MsLoadUserRankReply) error {
	var users []*table.T_Users
	res.ErrCode = "error"
	if req.RankType == defines.RankTypeDiamond {
		service.db.db.Order("diamond desc").Limit(req.Count).Find(&users)
		for i, u := range users {
			res.Users = append(res.Users, &proto.UserRankItem{
				Rank: i,
				Name: u.Name,
				UserId: int(u.Userid),
				HeadImg: u.Headimg,
				Value: int64(u.Diamond),
			})
		}
		res.ErrCode = "ok"
	}
	fmt.Println("serveice.loaduserrank ", users)
	return nil
}

func (service *DBService) LoadGameLibs(req *proto.MsLoadGameLibsArg, res *proto.MsLoadGameLibsReply) error {
	var l []table.T_Gamelib
	service.db.db.Find(&l)
	res.ErrCode = "ok"
	for _, lib := range l {
		res.Libs = append(res.Libs, proto.GameLibItem{
			Id: lib.Id,
			Name: lib.Name,
			Area: lib.Area,
			City: lib.City,
			Province: lib.Province,
		})
	}
	return nil
}