package gateway

import (
	"exportor/defines"
	net "network"
	"exportor/proto"
	"msgpacker"
	"fmt"
	"rpcd"
	"net/rpc"
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

	gw.fserver = net.NewTcpServer(&defines.NetServerOption{
		Host: gw.option.FrontHost,
		ConnectCb: func(client defines.ITcpClient) error {
			fmt.Println("client conect ", client.GetRemoteAddress())
			return nil
		},
		CloseCb: func(client defines.ITcpClient) {
			gw.cliManger.cliDisconnect(client)
			gw.serManager.clientDisconnected(client)
		},
		MsgCb: func(client defines.ITcpClient, m *proto.Message) {
			fmt.Println("handle client message ", m)
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

	gw.bserver = net.NewTcpServer(&defines.NetServerOption{
		Host: gw.option.BackHost,
		ConnectCb: func(client defines.ITcpClient) error {
			fmt.Println("server connection connected")
			return nil
		},
		CloseCb: func(client defines.ITcpClient) {
			gw.serManager.serDisconnected(client)
		},
		MsgCb: func(client defines.ITcpClient, m *proto.Message) {
			fmt.Println("handle bs server message ", m)
			gw.serManager.serMessage(client, m)
		},
		AuthCb: func(client defines.ITcpClient) error {
			return gw.authServer(client)
		},
	})

	func() {
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
	fmt.Println("auth server 1")
	m, err := client.Auth()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("auth server ", m)

	var register proto.RegisterServer
	if msgpacker.UnMarshal(m.Msg, &register) != nil {
		return err
	}

	return gw.serManager.addServer(client, &register)
}

func (gw *gateway) Stop() error {
	return nil
}

func (gw *gateway) routeCliMessage(client defines.ITcpClient, message *proto.Message) {
	fmt.Println(" ")
	fmt.Println("client cmd :", message.Cmd)
	fmt.Println(" ")
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

