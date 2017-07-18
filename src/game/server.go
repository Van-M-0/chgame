package game

import (
	"exportor/defines"
	"exportor/proto"
	"network"
	"fmt"
	"msgpacker"
)

type gameServer struct {
	opt 			*defines.GameOption
	gwClient 		defines.ITcpClient
	scmgr 			*sceneManager
}

func newGameServer(option *defines.GameOption) *gameServer {
	gs := &gameServer{}
	gs.opt = option
	gs.scmgr = newSceneManager(gs)
	return gs
}

func (gs *gameServer) Start() error {

	gs.gwClient = network.NewTcpClient(&defines.NetClientOption{
		Host: gs.opt.GwHost,
		ConnectCb: func (client defines.ITcpClient) error {
			fmt.Println("connect gate succcess, send auth info")
			client.Send(proto.CmdRegisterServer, &proto.RegisterServer{
				Type: "game",
			})
			return nil
		},
		CloseCb: func (client defines.ITcpClient) {

		},
		AuthCb: func (client defines.ITcpClient) error {
			return nil
		},
		MsgCb: func(client defines.ITcpClient, m *proto.Message) {
			var header proto.GateGameHeader
			if err := msgpacker.UnMarshal(m.Msg, &header); err != nil {
				fmt.Println("unmarshal client route lobby header error")
				return
			}
			gs.scmgr.onGwMessage(&header)
		},
	})
	gs.gwClient.Connect()

	return nil
}

func (gs *gameServer) Stop() error {
	return nil
}

func (gs *gameServer) authServer(message *proto.Message) {

}

func (gs *gameServer) send2players(uids[] uint32, cmd uint32, data interface{}) {
	body, err := msgpacker.Marshal(data)
	if err != nil {
		return
	}
	header := &proto.GameGateHeader{
		Uids: uids,
		Cmd: cmd,
		Msg: body,
	}
	fmt.Println("game send 2 player ", header)
	gs.gwClient.Send(proto.GameRouteClient, &header)
}



