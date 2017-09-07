//go:binary-only-package-my
package lobby

import (
	"network"
	"exportor/defines"
	"exportor/proto"
	"msgpacker"
	"fmt"
	"rpcd"
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
	rs 				*rankService
	cs 				*recordService
	as 				*Activities
	qs 				*QuestService
	is 				*IdentifyService
	clubs 			*AgentClub
	dbClient 		*rpcd.RpcdClient
	msClient 		*rpcd.RpcdClient
	serverId 		int
}

func newLobby(option *defines.LobbyOption) *lobby {
	lb := &lobby{}
	lb.opt = option
	lb.userMgr = newUserManager()
	lb.processor = newUserProcessorMgr()
	lb.bpro = newBrokerProcessor()
	lb.mall = newMallService(lb)
	lb.ns = newNoticeService(lb)
	lb.rs = newRankService(lb)
	lb.cs = newRecordService(lb)
	lb.as = newActivities(lb)
	lb.qs = newQuestService(lb)
	lb.is = newIdentifyService(lb)
	lb.clubs = newAgentClub(lb)
	return lb
}

func (lb *lobby) Start() error {

	lb.gwClient = network.NewTcpClient(&defines.NetClientOption{
		SendChSize: 10240,
		Host: lb.opt.GwHost,
		SendActor: 100,
		ConnectCb: func (client defines.ITcpClient) error {
			//fmt.Println("connect gate succcess, send auth info")
			var res proto.MsServerIdReply
			lb.msClient.Call("ServerService.GetServerId", &proto.MsServerIdArg{Type:"lobby"}, &res)
			lb.serverId = res.Id
			client.Send(proto.CmdRegisterServer, &proto.RegisterServer{
				Type: "lobby",
				ServerId: res.Id,
			})
			//fmt.Println("lobby auth ", res.Id)
			return nil
		},
		CloseCb: func (client defines.ITcpClient) {
			//fmt.Println("closed gate success")
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
	lb.ns.start()
	lb.rs.start()
	lb.mall.start()
	lb.cs.start()
	lb.as.start()
	lb.qs.start()
	lb.is.start()
	lb.clubs.start()
	return nil
}

func (lb *lobby) StartRpc() {
	lb.msClient = rpcd.StartClient(defines.MSServicePort)
	lb.dbClient = rpcd.StartClient(defines.DBSerivcePort)
	start := func() {
		rpcd.StartServer(defines.LbServicePort)
	}
	go start()
}

func (lb *lobby) Stop() error {
	return nil
}

func (lb *lobby) onGwMessage(message *proto.Message) {
	if message.Cmd == proto.ClientRouteLobby {
		var header proto.GateLobbyHeader
		if err := msgpacker.UnMarshal(message.Msg, &header); err != nil {
			fmt.Println("unmarshal client route lobby header error", err, header, message.Msg)
			return
		}
		lb.processor.process(header.Uid, func() {
			lb.handleClientMessage(header.Uid, header.Cmd, header.Msg)
		})
	} else if message.Cmd == proto.GateRouteLobby {
		var header proto.GateLobbyHeader
		if err := msgpacker.UnMarshal(message.Msg, &header); err != nil {
			fmt.Println("unmarshal client route lobby header error")
			return
		}
		//fmt.Println("gm", header.Cmd, header)
		lb.handleGateMessage(header.Uid, header.Cmd, header.Msg)
	} else {
		fmt.Println("lobby on gw message router error ", message)
	}
}

func (lb *lobby) handleClientMessage(uid uint32, cmd uint32, data []byte) {
	switch cmd {
	case proto.CmdLobbyPerformance:
		var pf proto.LobbyPerformance
		if err := msgpacker.UnMarshal(data, &pf); err != nil {
			fmt.Println("unmarshal performance packet error")
			return
		}
		//fmt.Println("cmsg ", uid, pf.SubCmd)
		lb.userMgr.handleUserPerformanceMessage(uid, &pf)
	case proto.CmdClientLogin:
		var login proto.ClientLogin
		if err := msgpacker.UnMarshal(data, &login); err != nil {
			fmt.Println("unmarshal client login errr", err)
			return
		}
		lb.userMgr.handleUserLogin(uid, &login)
	case proto.CmdGuestLogin:
		var guest proto.GuestLogin
		if err := msgpacker.UnMarshal(data, &guest); err != nil {
			fmt.Println("unmarshal client login errr", err)
			return
		}
	case proto.CmdWechatLogin:
		var login proto.WechatLoginReq
		if err := msgpacker.UnMarshal(data, &login); err != nil {
			fmt.Println("unmarshal client login errr", err)
			return
		}
		go func() {
			var res proto.MsSdkWechatLoginReply
			err := lb.msClient.Call("SdkService.WechatLogin", &proto.MsSdkWechatLoginArg{
				Code: login.Code,
				Device: login.Device,
			},&res)
			fmt.Println("wechat login call ", err)
			loginRet := &proto.WechatLoginRet{}
			loginRet.ErrCode = "ok"
			loginRet.Code = login.Code
			if res.ErrCode == "ok" {
				loginRet.OpenId = res.OpenId
				loginRet.Token = res.Token
			}
			fmt.Println("wechat login ret ", loginRet)
			lb.send2player(uid, proto.CmdWechatLogin, loginRet)
		}()
	case proto.CmdCreateAccount:
		var acc proto.CreateAccount
		if err := msgpacker.UnMarshal(data, &acc); err != nil {
			fmt.Println("unmarshal client account errr", err)
			return
		}
		lb.userMgr.handleCreateAccount(uid, &acc)
	case proto.CmdUserLoadNotice:
		var req proto.LoadNoticeListReq
		if err := msgpacker.UnMarshal(data, &req); err != nil {
			fmt.Println("unmarshal client account errr", err)
			return
		}
		lb.ns.handleLoadNotices(uid, &req)
	case proto.CmdHornMessage:
		var req proto.UserHornMessageReq
		if err := msgpacker.UnMarshal(data, &req); err != nil {
			fmt.Println("unmarshal client account errr", err)
			return
		}
		lb.userMgr.handleUserHornMessage(uid, &req)
	case proto.CmdClientLoadMallList:
		var req proto.ClientLoadMallList
		if err := msgpacker.UnMarshal(data, &req); err != nil {
			fmt.Println("unmarshal client account errr", err)
			return
		}
		lb.mall.onUserLoadMalls(uid, &req)
	case proto.CmdClientBuyItem:
		var req proto.ClientBuyReq
		if err := msgpacker.UnMarshal(data, &req); err != nil {
			fmt.Println("unmarshal client account errr", err)
			return
		}
		lb.mall.OnUserBy(uid, &req)
	case proto.CmdUserLoadRank:
		var req proto.ClientLoadUserRank
		if err := msgpacker.UnMarshal(data, &req); err != nil {
			fmt.Println("unmarshal client account errr", err)
			return
		}
		lb.rs.onUserGetRanks(uid, &req)
	case proto.CmdUserGetRecordList:
		var req proto.ClientGetRecordList
		if err := msgpacker.UnMarshal(data, &req); err != nil {
			fmt.Println("unmarshal record list errr", err)
			return
		}
		lb.cs.OnUserGetRecordList(uid, &req)
	case proto.CmdUserGetRecord:
		var req proto.ClientGetRecord
		if err := msgpacker.UnMarshal(data, &req); err != nil {
			fmt.Println("unmarshal record list errr", err)
			return
		}
		lb.cs.OnUserGetRecord(uid, &req)
	case proto.CmdUserLoadActivityList:
		var req proto.ClientLoadActitity
		if err := msgpacker.UnMarshal(data, &req); err != nil {
			fmt.Println("unmarshal record list errr", err)
			return
		}
		lb.as.OnUserLoadActivities(uid, &req)
	case proto.CmdUserLoadQuest:
		var req proto.ClientLoadQuest
		if err := msgpacker.UnMarshal(data, &req); err != nil {
			fmt.Println("unmarshal quest  errr", err)
			return
		}
		lb.qs.OnUserLoadQuest(uid, &req)
	case proto.CmdUserProcessQuest:
		var req proto.ClientProcessQuest
		if err := msgpacker.UnMarshal(data, &req); err != nil {
			fmt.Println("unmarshal quest process error")
			return
		}
		lb.qs.OnUserProcessQuest(uid, &req)
	case proto.CmdUserCompleteQuest:
		var req proto.ClientCompleteQuest
		if err := msgpacker.UnMarshal(data, &req); err != nil {
			fmt.Println("unmarshal quest complete error")
			return
		}
		lb.qs.OnUserCompletionQuest(uid, &req)
	case proto.CmdUserIdentify:
		var req proto.ClientIdentify
		if err := msgpacker.UnMarshal(data, &req); err != nil {
			fmt.Println("unmarshal quest process error")
			return
		}
		lb.is.OnUserCheckUserIdentifier(uid, &req)
	case proto.CmdUserCreatClub:
		var req proto.ClientCreateClub
		if err := msgpacker.UnMarshal(data, &req); err != nil {
			fmt.Println("unmarshal createclub error")
			return
		}
		lb.clubs.OnUserCreateClub(uid, &req)
	case proto.CmdUserJoinClub:
		var req proto.ClientJoinClub
		if err := msgpacker.UnMarshal(data, &req); err != nil {
			fmt.Println("unmarshal joinclub error")
			return
		}
		lb.clubs.OnUserJoinClub(uid, &req)
	case proto.CmdUserLeaveClub:
		var req proto.ClientLeaveClub
		if err := msgpacker.UnMarshal(data, &req); err != nil {
			fmt.Println("unmarshal joinclub error")
			return
		}
		lb.clubs.OnUserLeaveClub(uid, &req)
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
	//fmt.Println("ls2p ", header.Cmd, header.Uids)
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
	fmt.Println("lobby bc 2 player ", header.Cmd, header.Uids)
	lb.gwClient.Send(proto.LobbyRouteClient, &header)
}

func (lb *lobby) broadcastWorldMessage(cmd uint32, data interface{}) {
	body, err := msgpacker.Marshal(data)
	if err != nil {
		return
	}
	header := &proto.LobbyGateHeader{
		Uids: nil,
		Cmd: cmd,
		Msg: body,
	}
	fmt.Println("bc world message ", header)
	lb.gwClient.Send(proto.LobbyRouteClient, &header)
}

func (lb *lobby) handleGateMessage(uid, cmd uint32, data []byte) {
	if cmd == proto.CmdClientDisconnected {
		lb.userMgr.handleClientDisconnect(uid)
	}
}
