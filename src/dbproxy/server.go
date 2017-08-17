package dbproxy

import (
	"exportor/defines"
	"fmt"
	"net/rpc"
	"rpcd"
	"cacher"
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
	dbservice 			*DBService
	dtSaver 			*dataSaver
}

func newDBProxyServer() *dbProxyServer {
	dbServer := &dbProxyServer{}
	dbServer.dbClient = newDbClient()
	dbServer.dbservice = newDbService(dbServer.dbClient)
	dbServer.dtSaver = newDataSaver(dbServer.dbClient)
	dbServer.cacheClient = cacher.NewCacheClient("dbproxy")
	return dbServer
}

func (ds *dbProxyServer) Start() error {
	ds.cacheClient.Start()
	ds.dbservice.start()
	ds.startRpc()
	ds.load2Cache()
	ds.dtSaver.start()
	//ds.con.Start()
	//ds.pub.Start()
	//ds.cacheClient.Start()
	//ds.getMessageFromBroker()
	//ds.handleNotify()
	//ds.handleNotifyRes()
	return nil
}

func (ds *dbProxyServer) startRpc() {

	rpc.Register(ds.dbservice)

	start := func() {
		rpcd.StartServer(defines.DBSerivcePort)
	}
	go start()
}

func (ds *dbProxyServer) Stop() error {
	return nil
}

func (ds *dbProxyServer) load2Cache() {

	delScript := `
			local keys = redis.call('keys', ARGV[1])

			local removeKeys = {}
			for i = 1, #keys do
				local s, e = string.find(keys[i], ARGV[2])
				if not s or not e then
					table.insert(removeKeys, keys[i])
				end
			end

			for i=1,#removeKeys,5000 do
				redis.call('del', unpack(removeKeys, i, math.min(i+4999, #removeKeys)))
			end

			return #removeKeys
	`
	ds.cacheClient.Scripts(delScript, 0, "*", "record..*")

	/*
	ft, err := time.Parse("2006-01-02 15:04:05", "2017-08-08 09:04:01")
	if err != nil {
		fmt.Println("fmt time error ", err)
		return
	}
	testCreateNotice := func() {
		ds.dbClient.db.Create(&table.T_Notice{
			Starttime: time.Now(),
			Finishtime: ft,
			Kind: "notice",
			Content: "你好，世界",
			Playtime: 10,
			Playcount: 1,
		})

		ds.dbClient.db.Create(&table.T_Notice{
			Starttime: time.Now(),
			Finishtime: ft,
			Kind: "notice",
			Content: "你好，世界 2",
			Playtime: 10,
			Playcount: 1,
		})
	}

	testCreateNotice()
	var notice []table.T_Notice
	ds.dbClient.db.Find(&notice)
	fmt.Println("load notice ", notice)
	var cnotice []*proto.CacheNotice
	for _, n := range notice {
		cnotice = append(cnotice, &proto.CacheNotice{
			Id: n.Index,
			StartTime: n.Starttime.Format("2006-01-02 15:04:05"),
			FinishTime: n.Finishtime.Format("2006-01-02 15:04:05"),
			Kind: n.Kind,
			Content: n.Content,
			PlayTime: n.Playtime,
			PlayCount: n.Playcount,
		})
	}
	ds.cacheClient.NoticeOperation(&cnotice, "update")

*/
	/*
	var n1 []*proto.CacheNotice
	ds.cacheClient.NoticeOperation(&n1, "getall")
	fmt.Println("n1 < ", n1, n1[0])
	*/
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

/*
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
		fmt.Println("bbbbbb", req, ret)
		if ret == true {
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
	fmt.Println("aaaaaaaaaaa", req, ret)
}
func (ds *dbProxyServer) handleCreateAccount(i interface{}) {
	req := i.(*proto.PMCreateAccount)
	fmt.Println("create account ", req.Name)
	var user table.T_Users
	var userSuccess *table.T_Users
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
			userSuccess = &table.T_Users{
				Account: acc,
				Name: req.Name,
				Sex: req.Sex,
				Level: 1,
				Exp: 0,
				Coins: 100,
				Gems: 1,
			}
			r = ds.dbClient.AddUserInfo(userSuccess)
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
		if res.Err == 0 {
			ds.cacheClient.SetUserInfo(userSuccess, true)
		}
		ds.pub.WaitPublish(defines.ChannelTypeDb, defines.ChannelCreateAccountFinish, res)
	}
}
*/
func (ds *dbProxyServer) handleNotifyRes() {
	for i :=1; i < 3; i++ {
		go func() {
			for {
				select {
				case fn := <- ds.chResNotify:
					fn()
				}
			}
		}()
	}
}
