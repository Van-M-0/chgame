package dbproxy

import (
	"exportor/defines"
	"cacher"
	"communicator"
	"exportor/proto"
	"gopkg.in/vmihailenco/msgpack.v2"
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
	chDbRet     		chan *dbret
	chSaveCache 		chan *saveCache
	chNotifyChannel 	chan *notifyChannel
}

func newDbProxyServer() *dbProxyServer {
	return &dbProxyServer{
		cacheClient: cacher.NewCacheClient("dbproxy"),
		commClient:  communicator.NewCommunicator(nil),
		chNotify:    make(chan *notifier),
		chDbRet:     make(chan *dbret),
		chSaveCache: make(chan *saveCache),
		chNotifyChannel: make(chan *notifyChannel),
	}
}

func (ds *dbProxyServer) Start() {
	ds.handleNotify()
	ds.handleDbRet()
	ds.save2Cache()
	ds.notifyChannel()
}

func (ds *dbProxyServer) Stop() {

}

func (ds *dbProxyServer) joinChannel() {
	ds.commClient.JoinChanel("loadUser", false, func(d []byte) {
		m := ds.deserilize(castProxyLoadUser, d)
		if m != nil {
			m = m.(*proto.ProxyLoadUserInfo)
		}
		ds.chNotify <- &notifier{cmd: castProxyLoadUser, i: m}
	})

	ds.commClient.JoinChanel("createAccount", false, func(d []byte) {
	})

}

func (ds *dbProxyServer) handleNotify() {
	go func() {
		select {
		case n := <- ds.chNotify:
			switch n.cmd {
			case castProxyLoadUser:
				var userInfo proto.T_Users
				ds.chDbRet <- &dbret{
					cmd: n.cmd,
					ret: ds.dbClient.GetUserInfo(n.i.(*proto.ProxyLoadUserInfo).Name, &userInfo),
					i: &userInfo,
				}
			default:
			}
		}
	}()
}

func (ds *dbProxyServer) handleDbRet() {
	go func() {
		select {
		case r := <- ds.chDbRet:
			switch r.cmd {
			case castProxyLoadUser:
			default:

			}
		}
	}()
}

func (ds *dbProxyServer) save2Cache() {
	go func() {

	}()
}

func (ds *dbProxyServer) notifyChannel() {
	go func() {

	}()
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
