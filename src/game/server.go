package game

import (
	"exportor/defines"
	"exportor/proto"
	"network"
	"fmt"
	"msgpacker"
	"rpcd"
)

type gameServer struct {
	opt 			*defines.GameOption
	gwClient 		defines.ITcpClient
	msClient 		*rpcd.RpcdClient
	scmgr 			*sceneManager
	serverId 		int
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
			var res defines.MsServerIdReply
			gs.msClient.Call("ServerService.GetServerId", &defines.MsServerIdArg{Type:"game"}, &res)
			gs.serverId = res.Id
			client.Send(proto.CmdRegisterServer, &proto.RegisterServer{
				Type: "game",
				ServerId: res.Id,
			})
			return nil
		},
		CloseCb: func (client defines.ITcpClient) {

		},
		AuthCb: func (client defines.ITcpClient) error {
			/*
			m, err := client.Auth()
			if err != nil {
				return err
			}
			if m.Cmd != proto.GateRouteGame {
				err := fmt.Errorf("server auth error ")
				fmt.Println(err)
				return err
			}

			var header proto.GateGameHeader
			if err := msgpacker.UnMarshal(m.Msg, &header); err != nil {
				fmt.Println("auth game server ",  err)
				return err
			}


			if header.Type != proto.GateMsgTypeServer {
				err := fmt.Errorf("server auth type error ")
				fmt.Println(err)
				return err
			}

			var r proto.RegisterServerRet
			if err := msgpacker.UnMarshal(header.Msg, &r); err != nil {
				fmt.Println("auth game server ",  err)
				return err
			}

			gs.serverId = r.ServerId

			fmt.Println("game server id ", gs.serverId)
			*/
			return nil
		},
		MsgCb: func(client defines.ITcpClient, m *proto.Message) {
			var header proto.GateGameHeader
			if err := msgpacker.UnMarshal(m.Msg, &header); err != nil {
				fmt.Println("unmarshal client route lobby header error")
				return
			}
			gs.scmgr.onGwMessage(m.Cmd, &header)
		},
	})
	gs.startRpc()
	gs.gwClient.Connect()

	gs.scmgr.start()

	return nil
}

func (gs *gameServer) startRpc() {
	gs.msClient = rpcd.StartClient(defines.MSServicePort)
}

func (gs *gameServer) Stop() error {
	return nil
}

func (gs *gameServer) authServer(message *proto.Message) {

}

func (gs *gameServer) send2players(uids[] uint32, cmd uint32, data interface{}) {
	if len(uids) == 0 {
		fmt.Println("send player message empty uids")
		return
	}
	body, err := msgpacker.Marshal(data)
	if err != nil {
		fmt.Println("msg 2 player error", uids, data)
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

func (gs *gameServer) getSid() int {
	return gs.serverId
}


