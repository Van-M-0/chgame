package defines

import (
	"net"
)

type CodecCreator func() ICodec
type NetClientOption struct {
	Host       	string
	SendChSize 	int
	//Codec      	ICodec
	ConnectCb  	ClientConnectCb
	CloseCb    	ClientCloseCb
	MsgCb      	ClientMessageCb
	AuthCb     	AuthCb
}

type NetServerOption struct {
	GwHost		string
	CmHost 		string
	Host 		string

	MaxRecvSize int
	EncryptCode string
	//Codec       CodecCreator
	ConnectCb   ClientConnectCb
	CloseCb     ClientCloseCb
	MsgCb      	ClientMessageCb
	AuthCb      AuthCb
	listenConn  net.Conn
}

type GatewayOption struct {
	FrontHost 	string
	MaxClient   int

	BackHost 	string
}

type CommunicatorOption struct {
	Host 			string
	ReadTimeout 	int
	WriteTimeout	int
}