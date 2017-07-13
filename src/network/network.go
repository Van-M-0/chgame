package network

import (
	"exportor/defines"
	"exportor/network"
	"network/codec"
)

func NewTcpServer(opt *defines.NetServerOption) network.ITcpServer {
	return newServer(opt)
}

func NewTcpClient(opt *defines.NetClientOption) network.ITcpClient {
	return newTcpClient(opt)
}

func NewClientCodec() network.ICodec {
	return codec.NewClientCodec()
}

func NewServerCodec() network.ICodec {
	return codec.NewServerCodec()
}
