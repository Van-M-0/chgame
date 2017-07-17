package gateway

import (
	"exportor/network"
	"exportor/proto"
	"exportor/defines"
)

type cliManager struct {
	clis 		map[uint32]defines.ITcpClient
}

func newCliManager() *cliManager {
	return &cliManager{
		clis: make(map[uint32]defines.ITcpClient),
	}
}

func (mgr *cliManager) cliConnect(cli defines.ITcpClient) error {
	return nil
}

func (mgr *cliManager) cliDisconnect(cli defines.ITcpClient) {

}

func (mgr *cliManager) cliMsg(cli defines.ITcpClient, m *proto.Message) {

}

func (mgr *cliManager) routeToClient(uid uint32, data []byte) error {
	return nil
}

func (mgr *cliManager) bcToClient(uid []uint32, data []byte) error {
	return nil
}
