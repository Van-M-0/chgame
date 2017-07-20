package dbproxy

import (
	"exportor/defines"
	"cacher"
	"communicator"
	"exportor/proto"
	"gopkg.in/vmihailenco/msgpack.v2"
	"dbproxy/table"
	"fmt"
)

type notifier struct {
	cmd 	string
	i 		interface{}
}

type dbret struct {
	cmd 	int
	ret 	bool
	i 		interface{}
}

type saveCache struct {

}

type notifyChannel struct {

}

type dbProxyServer struct {
	cacheClient 		defines.ICacheClient
	com 				defines.ICommunicator
	pub 				defines.IMsgPublisher
	con 				defines.IMsgConsumer
	dbClient    		*dbClient
	chNotify    		chan *notifier
	chResNotify 		chan func()
}

func newDBProxyServer() *dbProxyServer {
	return &dbProxyServer{
		dbClient: newDbClient(),
		cacheClient: cacher.NewCacheClient("dbproxy"),
		//com:  communicator.NewCommunicator(),
		pub: communicator.NewMessagePulisher(),
		con: communicator.NewMessageConsumer(),
		chNotify:    make(chan *notifier, 4096),
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
		data := ds.con.GetMessage(defines.ChannelTypeDb, key)
		ds.chNotify <- &notifier{cmd: key, i: data}
	}

	go getChannelMessage(defines.ChannelLoadUser)
	go getChannelMessage(defines.ChannelCreateAccount)
}

func (ds *dbProxyServer) handleNotify() {
	go func() {
		select {
		case n := <- ds.chNotify:
			switch n.cmd {
			case defines.ChannelLoadUser:
				req := n.i.(proto.PMLoadUser)
				var userInfo table.T_Users
				ret := ds.dbClient.GetUserInfo(req.Acc, &userInfo)
				fmt.Println(defines.ChannelLoadUser, req, ret)
				ds.chResNotify <- func() {
					err := ds.cacheClient.SetUserInfo(&userInfo, ret)
					d, _ := msgpack.Marshal(&proto.PMLoadUserFinish{
						Acc: req.Acc,
						Err: err,
					})
					ds.pub.SendPublish(defines.ChannelLoadUserFinish, d)
				}
			}
		}
	}()
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

