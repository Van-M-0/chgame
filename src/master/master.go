package master

import (
	"exportor/defines"
	"rpcd"
	"net/rpc"
	"tools"
	"statics"
	"mylog"
)

type Master struct {
	hp 		*http2Proxy
	sdk 	*SdkService
	wdClient *rpcd.RpcdClient
	ss 		*statics.StaticsServer
}

func NewMasterServer () defines.IServer {
	ms := &Master{}
	ms.hp = newHttpProxy(ms)
	ms.sdk = newSdkService(ms)
	ms.ss = statics.NewStaticsServer()
	return ms
}

func (ms *Master) Start() error {
	ms.StartRpc()
	ms.StartHttp()
	ms.loadData()
	ms.ss.Start()
	tools.WaitForSignal()
	ms.Stop()
	return nil
}

func (ms *Master) loadData() {
	GameModService.load()
}

func (ms *Master) Stop() error {
	mylog.Debug("master stop ....")
	ms.ss.Stop()
	return nil
}

func (ms *Master) StartRpc() {
	start := func() {
		rpc.Register(GameSerService)
		rpc.Register(newRoomService())
		rpc.Register(GameModService)
		rpc.Register(ms.sdk)
		rpc.Register(ms.ss)
		rpcd.StartServer(defines.MSServicePort)
	}
	ms.wdClient = rpcd.StartClient(tools.GetWorldServiceHost())
	go start()
}

func (ms *Master) StartHttp() {
	ms.hp.start()
}