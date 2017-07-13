package network

import (
	"exportor/defines"
	"exportor/proto"
	"net"
	"exportor/network"
)

type tcpClient struct {
	netContext
	id 				uint32
	conn 			net.Conn
	opt 			*defines.NetClientOption
	sendCh			chan *proto.Message
}

func newTcpClient(opt *defines.NetClientOption) *tcpClient {
	client := &tcpClient{
		opt: opt,
		sendCh: make(chan *proto.Message, opt.SendChSize),
	}
	return client
}

func (client *tcpClient) Id(id uint32) {
	client.id = id
}

func (client *tcpClient) GetId() uint32 {
	return client.id
}

func (client *tcpClient) Close() {
	client.opt.CloseCb(client)
}

func (client *tcpClient) Send(m *proto.Message) {
	client.sendCh <- m
}

func (client *tcpClient) sendLoop() {
	for {
		select {
		case m:= <- client.sendCh:
			client.write(m)
		}
	}
}

func (client *tcpClient) configureConn(conn net.Conn) {
	client.conn = conn
}


func (client *tcpClient) write(m *proto.Message) error {
	if err := client.opt.Codec.Encode(m); err != nil {
		return err
	}
	return nil
}

func (client *tcpClient) GetCodec() network.ICodec {
	return client.opt.Codec
}
