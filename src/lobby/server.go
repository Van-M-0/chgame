//go:binary-only-package-my
package lobby

import (
	"network"
	"exportor/defines"
	"exportor/proto"
	"msgpacker"
	"fmt"
	"rpcd"
	"net/rpc"
)

type lobby struct {
	gwClient 		defines.ITcpClient
	userMgr 		*userManager
	opt 			*defines.LobbyOption
	processor 		*userProcessorManager
	bpro 			*brokerProcessor
	mall 			*mallService
	hp 				*http2Proxy
	ns 				*noticeService
	dbClient 		*rpcd.RpcdClient
	msClient 		*rpcd.RpcdClient
	serverId 		int
	gameService 	*GameService
}

func newLobby(option *defines.LobbyOption) *lobby {
	lb := &lobby{}
	lb.opt = option
	lb.userMgr = newUserManager()
	lb.processor = newUserProcessorMgr()
	lb.bpro = newBrokerProcessor()
	lb.mall = newMallService(lb)
	lb.hp = newHttpProxy()
	lb.ns = newNoticeService(lb)
	lb.gameService = newGameService(lb)
	return lb
}

func (lb *lobby) Start() error {

	lb.gwClient = network.NewTcpClient(&defines.NetClientOption{
		Host: lb.opt.GwHost,
		ConnectCb: func (client defines.ITcpClient) error {
			fmt.Println("connect gate succcess, send auth info")
			var res defines.MsServerIdReply
			lb.msClient.Call("ServerService.GetServerId", &defines.MsServerIdArg{Type:"lobby"}, &res)
			lb.serverId = res.Id
			client.Send(proto.CmdRegisterServer, &proto.RegisterServer{
				Type: "lobby",
				ServerId: res.Id,
			})

			return nil
		},
		CloseCb: func (client defines.ITcpClient) {
			fmt.Println("closed gate success")
		},
		AuthCb: func (client defines.ITcpClient) error {
			return nil
		},
		MsgCb: func(client defines.ITcpClient, m *proto.Message) {
			lb.onGwMessage(m)
		},
	})

	lb.StartRpc()
	lb.gwClient.Connect()

	lb.userMgr.setLobby(lb)
	lb.userMgr.start()
	lb.processor.Start()
	lb.bpro.Start()
	lb.hp.start()
	lb.ns.start()


	return nil
}

func (lb *lobby) StartRpc() {
	lb.msClient = rpcd.StartClient(defines.MSServicePort)
	lb.dbClient = rpcd.StartClient(defines.DBSerivcePort)
	start := func() {
		rpc.Register(lb.gameService)
		rpcd.StartServer(defines.LbServicePort)
	}
	go start()
}

func (lb *lobby) Stop() error {
	return nil
}

func (lb *lobby) onGwMessage(message *proto.Message) {
	fmt.Println("lobby on gw message ", message)
	if message.Cmd == proto.ClientRouteLobby {
		var header proto.GateLobbyHeader
		if err := msgpacker.UnMarshal(message.Msg, &header); err != nil {
			fmt.Println("unmarshal client route lobby header error")
			return
		}
		fmt.Println("lobby on gw message 1", header.Cmd, header)
		lb.processor.process(header.Uid, func() {
			lb.handleClientMessage(header.Uid, header.Cmd, header.Msg)
		})
	} else if message.Cmd == proto.GateRouteLobby {

	}
}

func (lb *lobby) handleClientMessage(uid uint32, cmd uint32, data []byte) {
	switch cmd {
	case proto.CmdClientLogin:
		var login proto.ClientLogin
		if err := msgpacker.UnMarshal(data, &login); err != nil {
			fmt.Println("unmarshal client login errr", err)
			return
		}
		fmt.Println("unmarshal client login", login)
		lb.userMgr.handleUserLogin(uid, &login)
	case proto.CmdGuestLogin:
		var guest proto.GuestLogin
		if err := msgpacker.UnMarshal(data, &guest); err != nil {
			fmt.Println("unmarshal client login errr", err)
			return
		}
		fmt.Println("unmarshal guest login", guest)
	case proto.CmdCreateAccount:
		var acc proto.CreateAccount
		if err := msgpacker.UnMarshal(data, &acc); err != nil {
			fmt.Println("unmarshal client account errr", err)
			return
		}
		fmt.Println("unmarshal create account", acc)
		lb.userMgr.handleCreateAccount(uid, &acc)
	default:
		fmt.Println("lobby handle invalid client cmd ", cmd)
	}
}

func (lb *lobby) send2player(uid uint32, cmd uint32, data interface{}) {
	body, err := msgpacker.Marshal(data)
	if err != nil {
		return
	}
	header := &proto.LobbyGateHeader{
		Uids: []uint32{uid},
		Cmd: cmd,
		Msg: body,
	}
	fmt.Println("lobby send 2 player ", header)
	lb.gwClient.Send(proto.LobbyRouteClient, &header)
}

func (lb *lobby) broadcastMessage(cmd uint32, data interface{}) {
	uids := lb.userMgr.getAllUsers()
	body, err := msgpacker.Marshal(data)
	if err != nil {
		return
	}
	header := &proto.LobbyGateHeader{
		Uids: uids,
		Cmd: cmd,
		Msg: body,
	}
	fmt.Println("lobby send 2 player ", header)
	lb.gwClient.Send(proto.LobbyRouteClient, &header)
}
