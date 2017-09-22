package world

import (
	"exportor/defines"
	"rpcd"
	"sync/atomic"
	"fmt"
	"net/rpc"
)

type World struct {
	hp 		*http2Proxy
	db 		*dbClient
	ms 		*MasterService
	msterIds int32
}

func NewWorldServer () defines.IServer {
	ws := &World{}
	ws.hp = newHttpProxy(ws)
	ws.db = newDbClient(&defines.GlobalConfig)
	ws.ms = NewMasterService(ws)
	ws.msterIds = 1
	return ws
}

func (wd *World) Start() error {
	fmt.Println("world start ........")
	wd.StartRpc()
	wd.StartHttp()
	wd.loadData()
	return nil
}

func (wd *World) loadData() {
}

func (wd *World) Stop() error {
	return nil
}

func (wd *World) StartRpc() {
	start := func() {
		rpc.Register(wd.ms)
		rpcd.StartServer(defines.WDServicePort)
	}
	go start()
}

func (wd *World) StartHttp() {
	wd.hp.start()
}

func (wd *World) getMasterId() int {
	atomic.AddInt32(&wd.msterIds, 1)
	return int(wd.msterIds)
}