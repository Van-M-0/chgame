package network

import (
	"exportor/defines"
	"net"
	"tools"
	"fmt"
)

type tcpServer struct {
	*netContext
	opt 			*defines.NetServerOption
	closeChn 		chan int
}

func newServer(opt *defines.NetServerOption) *tcpServer {
	server := &tcpServer{
		netContext: newNetContext(),
		opt: opt,
		closeChn: make(chan int),
	}
	return server
}

func (server *tcpServer) Start() error {

	l, err := net.Listen("tcp", server.opt.Host)
	if err != nil {
		return err
	}

	defer func() {
		l.Close()
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}

		tools.SafeGo(func() {
			server.handleClient(conn)
		})
	}
}

func (server *tcpServer) Stop() error {

}

func (server *tcpServer) handleClient(conn net.Conn) {

	client := newTcpClient(nil)
	client.configureConn(conn)

	defer func() {
		client.Close()
	}()

	if client.opt == nil {
		client.opt = &defines.NetClientOption{
			Codec: NewClientCodec(),
		}
	}

	if client.opt != nil && client.opt.ConnectCb != nil {
		client.opt.ConnectCb(client)
	}

	if client.opt != nil && client.opt.AuthCb != nil {
		if err := client.opt.AuthCb(client); err != nil {
			return
		}
	}

	go client.sendLoop()

	for {
		m, err := client.opt.Codec.Decode()
		if err != nil {
			fmt.Println("decode msg error")
			continue
		}

		client.opt.MsgCb(m)
	}

}
