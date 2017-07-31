package master

import (
	"exportor/defines"
	"rpcd"
	"net/rpc"
)

type Master struct {
}

func NewMasterServer () defines.IServer {
	return &Master{
	}
}

func (ms *Master) Start() error {
	ms.StartRpc()
	return nil
}

func (ms *Master) Stop() error {
	return nil
}

func (ms *Master) StartRpc() {
	start := func() {
		rpc.Register(newServerService())
		rpc.Register(newRoomService())
		rpcd.StartServer(defines.MSServicePort)
	}
	go start()
}