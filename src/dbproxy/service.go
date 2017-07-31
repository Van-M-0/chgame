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

func (service *DBService) UserLogin(req *defines.DbUserLoginArg, res *defines.DbUserLoginReply) error {

	var cacheUser proto.CacheUser
	if err := service.cc.GetUserInfo(req.Acc, &cacheUser); err == nil {
		res.Err = "ok"
		return nil
	}

	var userInfo table.T_Users
	ret := service.db.GetUserInfo(req.Acc, &userInfo)
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

func (service *DBService) CreateAccount(req *defines.DbCreateAccountArg, res *defines.DbCreateAccountReply) error {
	var user table.T_Users
	var userSuccess *table.T_Users
	fmt.Println("create account ", req)
	ret := service.db.GetUserInfoByName(req.UserName, &user)
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
				Coins: 100,
				Gems: 1,
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

