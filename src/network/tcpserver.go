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
	return nil
}

func (server *tcpServer) handleClient(conn net.Conn) {
	client := newTcpClient(&defines.NetClientOption{
	})
	client.configureConn(conn)

	defer func() {
		client.Close()
		server.opt.CloseCb(client)
	}()

	server.opt.ConnectCb(client)
	if server.opt.AuthCb(client) != nil {
		return
	}
	go client.sendLoop()

	for {
		m, err := client.readMessage()
		if err != nil {
			fmt.Println("decode msg error")
			continue
		}
		server.opt.MsgCb(client, m)
	}
}

