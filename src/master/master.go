package master

import (
	"exportor/defines"
	"rpcd"
	"net/rpc"
	"exportor/proto"
	"tools"
)

type Master struct {
	hp 		*http2Proxy
	sdk 	*SdkService
	wdClient *rpcd.RpcdClient
}

var _masterId int

func NewMasterServer () defines.IServer {
	ms := &Master{}
	ms.hp = newHttpProxy()
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
	ms.wdClient = rpcd.StartClient(tools.GetWorldServiceHost())
	var rep proto.WsGetMasterIdReply
	if err := ms.wdClient.Call("MasterService.GetMasterId", &proto.WsGetMasterIdArg{}, &rep); err != nil {
		panic("get master id error" + err.Error())
		return
	}
	_masterId = rep.Id
	go start()
}

func (ms *Master) StartHttp() {
	ms.hp.start()
}