package master

import (
	"exportor/defines"
	"rpcd"
	"net/rpc"
)

type Master struct {
	hp 		*http2Proxy

}

func NewMasterServer (cfg *defines.StartConfigFile) defines.IServer {
	return &Master{
		hp: newHttpProxy(cfg.HttpHost),
	}
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
		rpc.Register(newServerService())
		rpc.Register(newRoomService())
		rpc.Register(GameModService)
		rpcd.StartServer(defines.MSServicePort)
	}
	go start()
}

func (ms *Master) StartHttp() {
	ms.hp.start()
}