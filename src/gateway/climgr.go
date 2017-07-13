package gateway

import (
	"exportor/network"
	"exportor/proto"
)

type cliManager struct {
	clis 		map[uint32]network.ITcpClient
}

func newCliManager() *cliManager {
	return &cliManager{
		clis: make(map[uint32]network.ITcpClient),
	}
}

func (mgr *cliManager) cliConnect(cli network.ITcpClient) error {
	return nil
}

func (mgr *cliManager) cliDisconnect(cli network.ITcpClient) {

}

func (mgr *cliManager) cliMsg(cli network.ITcpClient, m *proto.Message) {

}
