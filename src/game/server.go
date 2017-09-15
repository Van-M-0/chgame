package game

import (
	"exportor/defines"
	"exportor/proto"
	"network"
	"msgpacker"
	"rpcd"
	"errors"
	"mylog"
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
		SendActor: 1,
		SendChSize: 10240,
		Host: gs.opt.GwHost,
		ConnectCb: func (client defines.ITcpClient) error {
			mylog.Debug("connect gate succcess, send auth info")
			var res proto.MsServerIdReply
			gs.msClient.Call("ServerService.GetServerId", &proto.MsServerIdArg{Type:"game"}, &res)
			gs.serverId = res.Id

			registerModules := func() string {
				mylog.Debug("register game conf ...", gs.opt.Moudles)
				var modReply proto.MsGameMoudleRegisterReply
				modList := make([]proto.MsModuleItem, 0)
				for _, m := range gs.opt.Moudles {
					mylog.Debug("register game conf", m.GameConf)
					modList = append(modList, proto.MsModuleItem{
						Kind: m.Type,
						GameConf: m.GameConf,
						GatewayHost: gs.opt.ClientHost,
					})
				}
				err := gs.msClient.Call("GameModuleService.RegisterModule", &proto.MsGameMoudleRegisterArg{
					ServerId: gs.serverId,
					ModList: modList,
				}, &modReply)

				mylog.Debug("registe server reply", modReply, err)

				return modReply.ErrCode
			}

			if err := registerModules(); err != "ok" {
				mylog.Debug("register game moudles error ", err)
				return errors.New("register server module error")
			}

			client.Send(proto.CmdRegisterServer, &proto.RegisterServer{
				Type: "game",
				ServerId: res.Id,
			})
			return nil
		},
		CloseCb: func (client defines.ITcpClient) {
			mylog.Debug("gameserver closed")
		},
		AuthCb: func (client defines.ITcpClient) error {
			return nil
		},
		MsgCb: func(client defines.ITcpClient, m *proto.Message) {
			var header proto.GateGameHeader
			if err := msgpacker.UnMarshal(m.Msg, &header); err != nil {
				mylog.Debug("unmarshal client route lobby header error")
				return
			}
			gs.scmgr.onGwMessage(m.Cmd, &header)
		},
	})
	gs.startRpc()
	gs.gwClient.Connect()

	gs.scmgr.start()

	/*
	go func() {
		signalChan := make(chan os.Signal, 1)
		<-signalChan
		var res proto.MsServerReleaseReply
		gs.msClient.Call("ServerService.GetServerId", &proto.MsServerReleaseArg{Id: gs.serverId}, &res)
		println("safe exit")
		os.Exit(0)
	}()
	*/

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

func (gs *gameServer) send2players(uids[] uint32, index uint32, cmd uint32, data interface{}) {
	if len(uids) == 0 {
		mylog.Debug("send player message empty uids")
		return
	}
	body, err := msgpacker.Marshal(data)
	if err != nil {
		mylog.Debug("msg 2 player error", uids, data)
		return
	}
	header := &proto.GameGateHeader{
		Uids: uids,
		Cmd: cmd,
		Index: index,
		Msg: body,
	}
	mylog.Debug("game send 2 player ", header.Uids, header.Cmd, header.Index)
	gs.gwClient.Send(proto.GameRouteClient, &header)
}

func (gs *gameServer) getSid() int {
	return gs.serverId
}


