package network

import (
	"exportor/defines"
	//"network/codec"
)

func NewTcpServer(opt *defines.NetServerOption) defines.ITcpServer {
	return newServer(opt)
}

func NewTcpClient(opt *defines.NetClientOption) defines.ITcpClient {
	return newTcpClient(opt)
}

/*
func NewClientCodec() defines.ICodec {
	return codec.NewClientCodec()
}

func NewServerCodec() defines.ICodec {
	return codec.NewServerCodec()
}
*/

func NewClientCodec() defines.ICodec {
	//return codec.NewMpCodec()
	return nil
}

