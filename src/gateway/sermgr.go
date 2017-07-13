package gateway

import (
	"exportor/network"
	"exportor/proto"
	"sync/atomic"
)

type serManager struct {
	idGen 		uint32
	sers 		map[uint32]network.ITcpClient
}

func newSerManager() *serManager {
	return &serManager{
		sers: make(map[uint32]network.ITcpClient),
	}
}

func (mgr *serManager) serConnected(client network.ITcpClient) {

}

func (mgr *serManager) serDisconnected(client network.ITcpClient) {

}

func (mgr *serManager) serMessage(client network.ITcpClient, m *proto.Message) {

}

func (mgr *serManager) addServer(client network.ITcpClient, m *proto.Message) error {
	client.Id(atomic.AddUint32(&mgr.idGen, 1))

	return nil
}
