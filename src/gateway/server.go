package gateway

import (
	"exportor/defines"
	net "network"
	"exportor/proto"
	"msgpacker"
	"fmt"
)

type gateway struct {
	option 		*defines.GatewayOption
	fserver 	defines.ITcpServer
	bserver 	defines.ITcpServer
	netOption 	*defines.NetServerOption
	idGen 		uint32
	cliManger 	*cliManager
	serManager 	*serManager
}

func NewGateServer(opt *defines.GatewayOption) *gateway {
	return &gateway{
		option: opt,
		cliManger: newCliManager(),
		serManager: newSerManager(),
	}
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

func (gw *gateway) authClient(client defines.ITcpClient) error {

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
	return nil
}

func (gw *gateway) routeCliMessage(client defines.ITcpClient, message *proto.Message) {
	cmd := message.Cmd
	if cmd >= proto.CmdRange_Base_S && cmd <= proto.CmdRange_Base_E {
		gw.handleClientMessage(client, message)
	} else if cmd >= proto.CmdRange_Gate_S && cmd <= proto.CmdRange_Gate_E {
		gw.handleClientMessage(client, message)
	} else if cmd >= proto.CmdRange_Lobby_S && cmd <= proto.CmdRange_Lobby_E {
		gw.serManager.client2Lobby(client, message)
	} else if cmd >= proto.CmdRange_Game_S && cmd <= proto.CmdRange_Game_E {
		gw.serManager.client2game(client, message)
	}
}

func (gw *gateway) handleClientMessage(client defines.ITcpClient, message *proto.Message) {
	switch message.Cmd {
	default:
		fmt.Println("invalid client cmd ", message.Cmd, client.GetRemoteAddress())
	}
}
