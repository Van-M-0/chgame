package dbproxy

import (
	"exportor/defines"
	"cacher"
	"communicator"
	"fmt"
	"exportor/proto"
	"dbproxy/table"
	"time"
	"strconv"
)

type request struct {
	cmd 	string
	i 		interface{}
}

type dbProxyServer struct {
	cacheClient 		defines.ICacheClient
	com 				defines.ICommunicator
	pub 				defines.IMsgPublisher
	con 				defines.IMsgConsumer
	dbClient    		*dbClient
	chNotify    		chan *request
	chResNotify 		chan func()
}

func newDBProxyServer() *dbProxyServer {
	return &dbProxyServer{
		dbClient: newDbClient(),
		cacheClient: cacher.NewCacheClient("dbproxy"),
		//com:  communicator.NewCommunicator(),
		pub: communicator.NewMessagePulisher(),
		con: communicator.NewMessageConsumer(),
		chNotify:    make(chan *request, 4096),
		chResNotify: make(chan func(), 4096),
	}
}

func (ds *dbProxyServer) Start() error {
	ds.con.Start()
	ds.pub.Start()
	ds.cacheClient.Start()
	ds.getMessageFromBroker()
	ds.handleNotify()
	ds.handleNotifyRes()
	return nil
}

func (ds *dbProxyServer) Stop() error {
	return nil
}


func (ds *dbProxyServer) getMessageFromBroker () {
	fmt.Println("get message from broker")
	getChannelMessage := func(key string) {
		for {
			data := ds.con.GetMessage(defines.ChannelTypeDb, key)
			fmt.Println("get message ", key, data)
			ds.chNotify <- &request{cmd: key, i: data}
		}
	}

	go getChannelMessage(defines.ChannelLoadUser)
	go getChannelMessage(defines.ChannelCreateAccount)
}

func (ds *dbProxyServer) handleNotify() {
	go func() {
		for {
			select {
			case n := <-ds.chNotify:
				switch n.cmd {
				case defines.ChannelLoadUser:
					ds.handleLogin(n.i)
				case defines.ChannelCreateAccount:
					ds.handleCreateAccount(n.i)
				}
			}
		}
	}()
}

func (ds *dbProxyServer) handleLogin(i interface{}) {
	req := i.(*proto.PMLoadUser)
	var userInfo table.T_Users
	ret := ds.dbClient.GetUserInfo(req.Acc, &userInfo)
	fmt.Println(defines.ChannelLoadUser, req, ret)
	ds.chResNotify <- func() {
		if ret {
			err := ds.cacheClient.SetUserInfo(&userInfo, ret)
			ret := &proto.PMLoadUserFinish{Err: err, Code: 0}
			fmt.Println("load user finish ", ret)
			ds.pub.WaitPublish(defines.ChannelTypeDb, defines.ChannelLoadUserFinish, ret)
		} else {
			ret := &proto.PMLoadUserFinish{Code: 1}
			fmt.Println("load user finish ", ret)
			ds.pub.WaitPublish(defines.ChannelTypeDb, defines.ChannelLoadUserFinish, ret)
		}
	}
}

func (ds *dbProxyServer) handleCreateAccount(i interface{}) {
	req := i.(*proto.PMCreateAccount)
	fmt.Println("create account ", req.Name)
	var user table.T_Users
	ret := ds.dbClient.GetUserInfoByName(req.Name, &user)
	var res proto.PMCreateAccountFinish
	if !ret {
		acc := "acc_" + strconv.Itoa(int(time.Now().Unix()))
		pwd := "123456"
		r := ds.dbClient.AddAccountInfo(&table.T_Accounts{
			Account: acc,
			Password: pwd,
		})
		if r {
			r = ds.dbClient.AddUserInfo(&table.T_Users{
				Account: acc,
				Name: req.Name,
				Sex: req.Sex,
				Level: 1,
				Exp: 0,
				Coins: 100,
				Gems: 1,
			})
			res.Err = 0
			res.Account = acc
			res.Pwd = pwd
		} else {
			res.Err = 2
		}
	} else {
		res.Err = 1
	}

	ds.chResNotify <- func() {
		fmt.Println("create account ret ", ret, res)
		ds.pub.WaitPublish(defines.ChannelTypeDb, defines.ChannelCreateAccountFinish, res)
	}
}

func (ds *dbProxyServer) handleNotifyRes() {
	for i :=1; i < 3; i++ {
		go func() {
			select {
			case fn := <- ds.chResNotify:
				fn()
			}
		}()
	}
}
