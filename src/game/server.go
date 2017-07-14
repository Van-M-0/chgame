package game

import (
	"exportor/defines"
	"exportor/proto"
	"network"
	"communicator"
)

type gameServer struct {
	gwClient 		defines.ITcpClient
	cmClient 		defines.ICommunicatorClient
	scmgr 			*sceneManager
}

func newGameServer() *gameServer {
	return &gameServer{
		cmClient: communicator.NewCommunicator(nil),
		scmgr: newSceneManager(),
	}
}

func (gs *gameServer) Start(opt *defines.NetServerOption) {

	gs.gwClient = network.NewTcpClient(&defines.NetClientOption{
		Host: opt.GwHost,
		ConnectCb: func (client defines.ITcpClient) error {
			return nil
		},
		CloseCb: func (client defines.ITcpClient) {

		},
		AuthCb: func (client defines.ITcpClient) error {
			return nil
		},
		MsgCb: func(client defines.ITcpClient, m *proto.Message) {
			gs.onGwMessage(m)
		},
	})

}

func (gs *gameServer) Stop() {

}

func (gs *gameServer) onGwMessage(m *proto.Message) {

}



