package gateway

import (
	"exportor/proto"
	"sync"
	"exportor/defines"
)

type serverInfo struct {
	typo 		string
	id 			uint32
	cli 		defines.ITcpClient
}

type serManager struct {
	sync.RWMutex
	idGen 		uint32
	sers 		map[uint32]*serverInfo
}

func newSerManager() *serManager {
	return &serManager{
		sers: make(map[uint32]*serverInfo),
	}
}

func (mgr *serManager) serConnected(client defines.ITcpClient) {

}

func (mgr *serManager) serDisconnected(client defines.ITcpClient) {
	mgr.Lock()
	mgr.Unlock()
	delete(mgr.sers, client.GetId())
}

func (mgr *serManager) serMessage(client defines.ITcpClient, m *proto.Message) {
	if m.Magic == proto.MagicDirectionGate {
		mgr.routeServer(client, m.Cmd, m.Msg)
	} else if m.Magic == proto.MagicDirectionClient {
		mgr.routeClient(client, m)
	}
}

func (mgr *serManager) addServer(client defines.ITcpClient, m *proto.RegisterServer) error {
	mgr.Lock()
	defer mgr.Unlock()

	mgr.idGen++
	mgr.sers[mgr.idGen] = &serverInfo{
		typo: m.Type,
		id:	mgr.idGen,
		cli: client,
	}
	client.Id(mgr.idGen)
	return nil
}

func (mgr *serManager) routeServer(client defines.ITcpClient, md int, m interface{}) {

}

func (mgr *serManager) routeClient(client defines.ITcpClient, m *proto.Message) {

}
