package gateway

import (
	"exportor/network"
	"exportor/proto"
	"sync/atomic"
	"sync"
	"errors"
)

type serverInfo struct {
	typo 		string
	id 			uint32
	cli 		network.ITcpClient
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

func (mgr *serManager) serConnected(client network.ITcpClient) {

}

func (mgr *serManager) serDisconnected(client network.ITcpClient) {
	mgr.Lock()
	mgr.Unlock()
	delete(mgr.sers, client.GetId())
}

func (mgr *serManager) serMessage(client network.ITcpClient, m *proto.Message) {
	if m.Magic == proto.MagicDirectionGate {
		mgr.routeServer(client, m.Cmd, m.Msg)
	} else if m.Magic == proto.MagicDirectionClient {
		mgr.routeClient(client, m)
	}
}

func (mgr *serManager) addServer(client network.ITcpClient, m *proto.RegisterServer) error {
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

func (mgr *serManager) routeServer(client network.ITcpClient, md int, m interface{}) {

}

func (mgr *serManager) routeClient(client network.ITcpClient, m *proto.Message) {

}
