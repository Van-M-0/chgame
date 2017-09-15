package master

import (
	"exportor/defines"
	"rpcd"
	"net/rpc"
)

type Master struct {
	hp 		*http2Proxy
	sdk 	*SdkService
	wdClient *rpcd.RpcdClient
}

func NewMasterServer (cfg *defines.StartConfigFile) defines.IServer {
	ms := &Master{}
	ms.hp = newHttpProxy(cfg.HttpHost)
	ms.sdk = newSdkService(ms)
	return ms
}

func (ms *Master) Start() error {
	ms.StartRpc()
	ms.StartHttp()
	ms.loadData()
	return nil
}

func (ms *Master) loadData() {
	GameModService.load()
}

func (ms *Master) Stop() error {
	return nil
}

func (ms *Master) StartRpc() {
	start := func() {
		rpc.Register(GameSerService)
		rpc.Register(newRoomService())
		rpc.Register(GameModService)
		rpc.Register(ms.sdk)
		rpcd.StartServer(defines.MSServicePort)
	}
	//ms.wdClient = rpcd.StartClient(defines.WDServicePort)
	go start()
}

func (ms *Master) StartHttp() {
	ms.hp.start()
}