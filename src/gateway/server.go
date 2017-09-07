package gateway

import (
	"exportor/defines"
	net "network"
	"exportor/proto"
	"msgpacker"
	"fmt"
	"rpcd"
	"time"

	"runtime"
	"sync/atomic"
)

type gateway struct {
	option 		*defines.GatewayOption
	fserver 	defines.ITcpServer
	bserver 	defines.ITcpServer
	netOption 	*defines.NetServerOption
	idGen 		uint32
	cliManger 	*cliManager
	serManager 	*serManager
	msClient 	*rpcd.RpcdClient
}

func NewGateServer(opt *defines.GatewayOption) *gateway {
	gateway := &gateway{}
	gateway.option = opt
	gateway.cliManger = newCliManager(gateway)
	gateway.serManager = newSerManager(gateway)
	return gateway
}

func (gw *gateway) Start() error {

	go func() {
		for {
			t := time.NewTimer(time.Minute * 1)
			select {
			case <-t.C:
				runtime.GC()
			}
		}
	}()

	var xxcountr [60]int32
	fmtcounter := func() {
		//fmt.Println("time recv ......", xxcountr)
	}

	go func() {
		for {
			select {
			case <- time.After(time.Second * 3):
				fmtcounter()
			}
		}
	}()


	gw.fserver = net.NewTcpServer(&defines.NetServerOption{
		SendChSize: 256,
		Host: gw.option.FrontHost,
		RecvNum: 10,
		ConnectCb: func(client defines.ITcpClient) error {
			fmt.Println("client conect ", client.GetRemoteAddress())
			client.Set("deadline", time.Minute * 5)
			return nil
		},
		CloseCb: func(client defines.ITcpClient) {
			gw.serManager.clientDisconnected(client)
			gw.cliManger.cliDisconnect(client)
		},
		MsgCb: func(client defines.ITcpClient, m *proto.Message) {
			atomic.AddInt32(&xxcountr[time.Now().Second()], 1)
			gw.routeCliMessage(client, m)
		},
		AuthCb: func(client defines.ITcpClient) error {
			return gw.authClient(client)
		},
	})

	go func() {
		err := gw.fserver.Start()
		fmt.Println("fs server start ", err)
	}()

	worker := func() chan func() {
		ch := make(chan func(), 128)
		go func() {
			for {
				select {
				case f := <- ch:
					f()
				}
			}
		}()
		return ch
	}

	workerSize := 512
	workerSlot := make([]chan func(), workerSize)
	for i := 0; i < workerSize; i ++ {
		workerSlot[i] = worker()
	}
	dispatchSlot := 0


	gw.bserver = net.NewTcpServer(&defines.NetServerOption{
		SendChSize: 4096,
		Host: gw.option.BackHost,
		SendActor: 100,
		ConnectCb: func(client defines.ITcpClient) error {
			return nil
		},
		CloseCb: func(client defines.ITcpClient) {
			gw.serManager.serDisconnected(client)
			var res proto.MsServerDisReply
			err := gw.msClient.Call("ServerService.ServerDisconnected", &proto.MsServerDiscArg{Id: int(client.GetId())}, &res)
			fmt.Println("server disconnect ", err, res.ErrCode)
		},
		MsgCb: func(client defines.ITcpClient, m *proto.Message) {
			dispatchSlot++
			if m.Cmd == proto.LobbyRouteClient {

				header := &proto.LobbyGateHeader{}
				err := msgpacker.UnMarshal(m.Msg, &header)
				if err != nil {
					return
				}
				index := dispatchSlot % workerSize
				if len(workerSlot[index]) == 128 {
					fmt.Println("*************************")
				}
				workerSlot[index] <- func() {
					gw.lobbyRoute2client(header.Uids, header.Cmd, header.Msg)
				}

			} else if m.Cmd == proto.GameRouteClient {

				header := &proto.GameGateHeader{}
				err := msgpacker.UnMarshal(m.Msg, &header)
				if err != nil {
					return
				}
				index := int(header.Index)
				if index == 0 {
					index = dispatchSlot % workerSize
				}
				workerSlot[index%workerSize] <- func() {
					gw.gameRoute2client(header.Uids, header.Cmd, header.Msg)
				}

			} else if m.Cmd == proto.LobbyRouteGate {

			} else if m.Cmd == proto.GameRouteGate {

			}

			if m.Cmd == proto.GameRouteClient {
				gw.serManager.serMessage(client, m)
			} else {

			}
		},
		AuthCb: func(client defines.ITcpClient) error {
			return gw.authServer(client)
		},
	})

	gw.startRpc()

	go func() {
		err := gw.bserver.Start()
		fmt.Println("bs server start ", err)
	}()

	return nil
}

func (gw *gateway) startRpc() {
	gw.msClient = rpcd.StartClient(defines.MSServicePort)
}

func (gw *gateway) authClient(client defines.ITcpClient) error {
	gw.cliManger.cliConnect(client)
	return nil
}

func (gw *gateway) authServer(client defines.ITcpClient) error {
	m, err := client.Auth()
	if err != nil {
		return err
	}

	var register proto.RegisterServer
	if msgpacker.UnMarshal(m.Msg, &register) != nil {
		return err
	}

	return gw.serManager.addServer(client, &register)
}

func (gw *gateway) Stop() error {
	gw.fserver.Stop()
	gw.bserver.Stop()
	return nil
}

func (gw *gateway) routeCliMessage(client defines.ITcpClient, message *proto.Message) {
	fmt.Println("4c ", message.Cmd, client.GetId())

	cmd := message.Cmd
	if cmd >= proto.CmdRange_Base_S && cmd <= proto.CmdRange_Base_E {
		gw.handleClientMessage(client, message)
	} else if cmd >= proto.CmdRange_Gate_S && cmd <= proto.CmdRange_Gate_E {
		gw.handleClientMessage(client, message)
	} else if cmd >= proto.CmdRange_Lobby_S && cmd <= proto.CmdRange_Lobby_E {
		gw.serManager.client2Lobby(client, message)
	} else if cmd >= proto.CmdRange_Game_S && cmd <= proto.CmdRange_Game_E {
		gw.serManager.client2game(client, message)
	} else {
		fmt.Println("gateway unkown message cmd", cmd)
	}
}

func (gw *gateway) handleClientMessage(client defines.ITcpClient, message *proto.Message) {
	switch message.Cmd {
	default:
		fmt.Println("invalid client cmd ", message.Cmd, client.GetRemoteAddress())
	}
}

func (gw *gateway) lobbyRoute2client(uids []uint32, cmd uint32, data []byte) {
	gw.cliManger.route2client(uids, cmd, data)
}

func (gw *gateway) gameRoute2client(uids []uint32, cmd uint32, data []byte) {
	gw.cliManger.route2client(uids, cmd, data)
}

