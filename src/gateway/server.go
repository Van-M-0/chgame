package gateway

import (
	"exportor/defines"
	"exportor/network"
	net "network"
	"exportor/proto"
	"errors"
)

type gateway struct {
	option 		*defines.GatewayOption
	fserver 	network.ITcpServer
	bserver 	network.ITcpServer
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
		ConnectCb: func(client network.ITcpClient) error {
			client.GetRemoteAddress()
			return nil
		},
		CloseCb: func(client network.ITcpClient) {
			gw.cliManger.cliDisconnect(client)
		},
		MsgCb: func(client network.ITcpClient, m *proto.Message) {
			gw.cliManger.cliMsg(client, m)
		},
		AuthCb: func(client network.ITcpClient) error {
			gw.cliManger.cliConnect(client)
			return nil
		},
	})

	if err := gw.fserver.Start(); err != nil {
		return err
	}

	gw.bserver = net.NewTcpServer(&defines.NetServerOption{
		Host: gw.option.FrontHost,
		ConnectCb: func(client network.ITcpClient) error {
			return nil
		},
		CloseCb: func(client network.ITcpClient) {
			gw.serManager.serDisconnected(client)
		},
		MsgCb: func(client network.ITcpClient, m *proto.Message) {
			gw.serManager.serMessage(client, m)
		},
		AuthCb: func(client network.ITcpClient) error {
			gw.authServer(client)
			return nil
		},
	})

	if err := gw.bserver.Start(); err != nil {
		return err
	}


	return nil
}

func (gw *gateway) authClient(client network.ITcpClient) error {
	encrypt := make([]byte, 32)
	if err := client.ActiveRead(encrypt, 32); err != nil {
		return err
	}

	return nil
}

func (gw *gateway) authServer(client network.ITcpClient) error {
	codec := client.GetCodec()
	m, err := codec.Decode()
	if err != nil {
		return err
	}

	if m.Magic != proto.MagicDirectionGate {
		return errors.New("not gate direction")
	}

	serInfo, ok := m.Msg.(*proto.RegisterServer)
	if !ok {
		return errors.New("cast register server info err")
	}

	return gw.serManager.addServer(client, serInfo)
}


func (gw *gateway) Stop() error {
	return nil
}
