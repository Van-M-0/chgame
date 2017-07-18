package dbproxy

import (
	"exportor/defines"
	"cacher"
	"communicator"
	"exportor/proto"
	"gopkg.in/vmihailenco/msgpack.v2"
	"fmt"
	"dbproxy/table"
)


const (
	castProxyLoadUser = 1
)

type notifier struct {
	cmd 	int
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
	commClient  		defines.ICommunicatorClient
	dbClient    		*dbClient
	chNotify    		chan *notifier
	chResNotify 		chan func()
}

func newDBProxyServer() *dbProxyServer {
	return &dbProxyServer{
		cacheClient: cacher.NewCacheClient("dbproxy"),
		commClient:  communicator.NewCommunicator(nil),
		chNotify:    make(chan *notifier, 4096),
		chResNotify: make(chan func(), 4096),
	}
}

func (ds *dbProxyServer) Start() {
	ds.handleNotify()
	ds.handleNotifyRes()
}

func (ds *dbProxyServer) Stop() {

}

func (ds *dbProxyServer) joinChannel() {
	ds.commClient.JoinChanel(defines.ChannelLoadUser, false, defines.WaitChannelInfinite, func(d []byte) {
		m := ds.deserilize(castProxyLoadUser, d)
		if m != nil {
			ds.chNotify <- &notifier{cmd: castProxyLoadUser, i: m.(*proto.NotifyRequest).Req}
		} else {
			fmt.Println("notify : cast loaduser err")
		}
	})

	ds.commClient.JoinChanel(defines.ChannelCreateAccount, false, defines.WaitChannelInfinite, func(d []byte) {

	})
}

func (ds *dbProxyServer) handleNotify() {
	go func() {
		select {
		case n := <- ds.chNotify:
			switch n.cmd {
			case castProxyLoadUser:
				var userInfo table.T_Users
				ret := ds.dbClient.GetUserInfo(n.i.(*proto.ProxyLoadUserInfo).Name, &userInfo)
				ds.chResNotify <- func() {
					err := ds.cacheClient.SetUserInfo(&userInfo, ret)
					d, _ := msgpack.Marshal(&proto.NotifyResponse{
						Err: err,
						Res: &userInfo,
					})
					ds.commClient.Notify(defines.ChannelLoadUserFinish, d)
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

func (ds *dbProxyServer) deserilize(cmd int, d []byte) interface{} {
	var i interface{}
	if cmd == castProxyLoadUser {
		i = &proto.ProxyLoadUserInfo{}
	}
	err := msgpack.Unmarshal(d, i)
	if err != nil {
		return i
	}
	return nil
}
