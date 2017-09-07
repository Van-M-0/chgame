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

	//service.lock.Lock()
	ret := service.db.GetUserInfo(req.Acc, &userInfo)
	//service.lock.Unlock()
	fmt.Println("luser login ", req, res)
	if req.LoginType == defines.LoginTypeWechat {
		if !ret {
			name := req.Name
			pwd := "11111"
			r := service.db.AddAccountInfo(&table.T_Accounts{
				Account: req.Acc,
				Password: pwd,
			})
			var userSuccess *table.T_Users
			if r {
				userSuccess = &table.T_Users{
					Account:  req.Acc,
					Name:     name,
					Headimg:  req.Headimg,
					Sex: 	  uint8(req.Sex),
					OpenId:   req.Acc,
					Level:    1,
					Exp:      0,
					Gold:     100,
					Diamond:  10,
					RoomCard: 1,
					Score:    0,
					Regtime:  time.Now(),
					Ip: 	  "",
				}
				r = service.db.AddUserInfo(userSuccess)
			}
		}
		//service.lock.Lock()
		ret = service.db.GetUserInfo(req.Acc, &userInfo)
		//service.lock.Unlock()

		userInfo.Headimg = req.Headimg
	} else {
		if !ret {
			name := "name_" + req.Acc
			pwd := "11111"
			r := service.db.AddAccountInfo(&table.T_Accounts{
				Account: req.Acc,
				Password: pwd,
			})
			var userSuccess *table.T_Users
			if r {
				userSuccess = &table.T_Users{
					Account:  req.Acc,
					Name:     name,
					Headimg:  req.Headimg,
					Sex: 	  uint8(req.Sex),
					OpenId:   "",
					Level:    1,
					Exp:      0,
					Gold:     100,
					Diamond:  10,
					RoomCard: 1,
					Score:    0,
					Regtime:  time.Now(),
					Ip: 	  "",
				}
				r = service.db.AddUserInfo(userSuccess)
			}
		}
		//service.lock.Lock()
		ret = service.db.GetUserInfo(req.Acc, &userInfo)
		//service.lock.Unlock()
	}

	fmt.Println("user login ", req)
	if ret == true {

		// restore cache data
		var cacheUser proto.CacheUser
		if err := service.cc.GetUserInfoById(userInfo.Userid, &cacheUser); err == nil {
			userInfo.Roomid = uint32(cacheUser.RoomId)
			fmt.Println("restore roomid is ", userInfo.Roomid)
		}

		err := service.cc.SetUserInfo(&userInfo, ret)
		if err != nil {
			fmt.Println("set cache user error ", err)
			res.Err = "cache"
		} else {
			res.Err = "ok"
			var itemlist []table.T_UserItem
			service.db.db.Where("userid = ?",  userInfo.Userid).Find(&itemlist)

			var items []proto.UserItem
			for _, item := range itemlist {
				items = append(items, proto.UserItem {
					ItemId: item.Itemid,
					Count: item.Count,
				})
			}
			service.cc.UpdateUserItems(userInfo.Userid, items)

			var ud table.T_Userdata
			service.db.db.Where("userid = ?", userInfo.Userid).Find(&ud)
			res.Ud = ud.Data

			var ii table.T_AuthInfo
			service.db.db.Where("userid = ?", userInfo.Userid).Find(&ii)
			if ii.Userid != 0 {
				res.Identify = proto.SynceIdentifyInfo{
					Name: ii.Name,
					Phone: ii.Phone,
					Card: ii.Idcard,
				}
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
	//service.lock.Lock()
	ret := service.db.GetUserInfo(req.Acc, &user)
	//service.lock.Unlock()
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
				Regtime: time.Now(),
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

func (service *DBService) LoadActivity(req *proto.MsLoadActivitysArg, res *proto.MsLoadActivitysReply) error {
	var la []*table.T_Activity
	var lr []*table.T_ActivityReward
	service.db.db.Find(&la)
	service.db.db.Find(&lr)

	res.ErrCode = "ok"
	for _, a := range la {
		res.Activitys = append(res.Activitys, &proto.ActivityItem{
			Id: a.Id,
			Desc: a.Desc,
			Actype: a.Actype,
			Starttime: a.Starttime,
			Finishtime: a.Finishtime,
			Rewardids: a.Rewardids,
		})
	}

	for _, r := range lr {
		res.ActivityRewards = append(res.ActivityRewards, &proto.ActivityRewardItem{
			Id: r.Id,
			RewardType: r.RewardType,
			ItemId: r.ItemId,
			Num: r.Num,
		})
	}

	return nil
}

func (service *DBService) LoadQuests(req *proto.MsLoadQuestArg, res *proto.MsLoadQuestReply) error {
	var ql []table.T_Quest
	var qr []table.T_QuestReward
	service.db.db.Find(&ql)
	service.db.db.Find(&qr)

	res.ErrCode = "ok"
	for _, q := range ql {
		res.Quests = append(res.Quests, proto.QuestItem{
			Id: q.Id,
			Title: q.Title,
			Content: q.Content,
			Type: q.Type,
			MaxCount: q.MaxCount,
			RewardIds: q.RewardIds,
		})
	}

	for _, r := range qr {
		res.Rewards = append(res.Rewards, proto.QuestRewardItem{
			Id: r.Id,
			ItemId: r.ItemId,
			Num: r.Num,
		})
	}

	return nil
}

func (service *DBService) SaveUserData(req *proto.MsSaveUserDataArg, rep *proto.MsSaveUserDataReply) error {
	service.db.db.Where("userid = ?", req.UserId).Save(&table.T_Userdata{
		Userid: req.UserId,
		Data: req.UserData,
	})
	rep.ErrCode = "ok"
	return nil
}

func (service *DBService) SaveIdentifyInfo(req *proto.MsSaveIdentifyInfoArg, rep *proto.MsSaveIdentifyInfoReply) error {
	service.db.db.Where("userid = ?", req.Userid).Save(&table.T_AuthInfo{
		Userid: req.Userid,
		Phone: req.Phone,
		Name: req.Name,
		Idcard: req.Idcard,
	})
	rep.ErrCode = "ok"
	return nil
}

func (service *DBService) LoadClubInfo(req *proto.MsLoadClubInfoReq, rep *proto.MsLoadClubInfoReply) error {
	var cc []table.T_Club
	var ce []table.T_ClubMember
	service.db.db.Find(&cc)
	service.db.db.Find(&ce)

	rep.ErrCode = "ok"
	for _, club := range cc {
		rep.Clubs = append(rep.Clubs, &proto.ClubItem {
			Id: club.Id,
			CreatorName: club.Creatorname,
			CreatorId: club.Creatorid,
		})
	}

	for _, member := range ce {
		rep.ClubMembers = append(rep.ClubMembers, &proto.ClubMemberItem {
			UserId: member.Userid,
			ClubId: member.Clubid,
		})
	}

	return nil
}

func (service *DBService) ClubOperation(req *proto.MsClubOperationReq, rep *proto.MsClubOperationReply) error {
	rep.ErrCode = "ok"
	if req.Op == "create" {
		e1 := service.db.db.Create(&table.T_Club{
			Id: req.Club.Id,
			Creatorid: req.Club.CreatorId,
			Creatorname: req.Club.CreatorName,
		}).RowsAffected != 0
		e2 := service.db.db.Create(&table.T_ClubMember{
			Userid: req.UserId,
			Clubid: req.Club.Id,
		}).RowsAffected != 0
		if !e1 || !e2 {
			fmt.Println(e1, e2)
			rep.ErrCode = "err"
		}
	} else if req.Op == "join" {
		e2 := service.db.db.Create(&table.T_ClubMember{
			Userid: req.UserId,
			Clubid: req.Club.Id,
		}).RowsAffected != 0
		if !e2  {
			fmt.Println(e2)
			rep.ErrCode = "err"
		}
	} else if req.Op == "leave" {
		var club table.T_Club
		service.db.db.Where("id = ?", req.Club.Id).Find(&club)
		var member table.T_ClubMember
		service.db.db.Where("userid = ?", req.UserId).Find(&member)
		if club.Id == 0 || member.Userid == 0 {
			rep.ErrCode = "err"
		} else {
			del := service.db.db.Where("clubid = ? and userid = ?", member.Clubid, member.Userid).Delete(&member).RowsAffected
			fmt.Println("dele rowo ", del)
			if del == 0 {
				rep.ErrCode = "err"
			}
		}
	}

	rep.Club = req.Club
	rep.UserId = req.UserId

	return nil
}
