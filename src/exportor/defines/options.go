package defines

import (
	"exportor/network"
	"net"
)

type NetClientOption struct {
	Host       string
	Port       int
	SendChSize int
	Codec      network.ICodec
	ConnectCb  network.ClientConnectCb
	CloseCb    network.ClientCloseCb
	MsgCb      network.ClientMessageCb
	AuthCb     network.AuthCb
}

type NetServerOption struct {
	Host 		string
	MaxRecvSize int
	EncryptCode string
	Codec       network.ICodec
	ConnectCb   network.ClientConnectCb
	CloseCb     network.ClientCloseCb
	MsgCb       network.ClientMessageCb
	AuthCb      network.AuthCb
	listenConn  net.Conn
}

type GatewayOption struct {
	FrontHost 	string
	MaxClient   int

	BackHost 	string
}