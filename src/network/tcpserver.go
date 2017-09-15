package network

import (
	"exportor/defines"
	"net"
	"time"
	"mylog"
)

type tcpServer struct {
	*netContext
	opt 			*defines.NetServerOption
	closeChn 		chan int
	sendChan 		chan *tcpClient
}

func newServer(opt *defines.NetServerOption) *tcpServer {
	server := &tcpServer{
		netContext: newNetContext(),
		opt: opt,
		closeChn: make(chan int, 1),
		sendChan: make(chan *tcpClient, 50000),
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

	//server.startSendLoop()

	acc := func() {
		for {
			select {
			case <- server.closeChn:
				mylog.Debug("tcp server stop")
			default:
				conn, err := l.Accept()
				if err != nil {
					continue
				}

				go func() {
					server.handleClient(conn)
				}()
			}
		}
	}

	for i := 0; i < 10 ; i ++ {
		go acc()
	}

	acc()


	return nil
}

func (server *tcpServer) Stop() error {
	server.closeChn <- 1
	return nil
}

func (server *tcpServer) startSendLoop() {

	//var xxcountr [60]int32
	fmtcounter := func() {
		//mylog.Debug("time sa", xxcountr)
	}

	go func() {
		for {
			select {
			case <- time.After(time.Second * 3):
				fmtcounter()
			}
		}
	}()

	for i := 0; i < 1256; i++ {
		go func() {
			mylog.Debug("send loop start ", i)
			for {
				select {
				case client := <- server.sendChan:
					if !client.IsClosed() {
						timediff := time.Now().Sub(client.lastInTime)
						if timediff > time.Second * 1 {
							///mylog.Debug("process queu time is ", timediff)
						}
						client.lastInTime = time.Now()
						if len(client.sendCh) > 0 {
							count := client.FlushSendBuffer()
							for cur := len(client.sendCh); cur > 0;  {
								count += client.FlushSendBuffer()
								//mylog.Debug("packet still more than > 50", client.GetId())
							}
							//atomic.AddInt32(&xxcountr[time.Now().Second()], 1)
						}
						/*
						client.notifyLock.Lock()
						client.notified = false
						client.notifyLock.Unlock()
						*/
					}
				}
			}
		}()
	}
}

func (server *tcpServer) handleClient(conn net.Conn) {
	client := newTcpClient(&defines.NetClientOption{
		SendChSize: server.opt.SendChSize,
		SendActor: server.opt.SendActor,
	})
	client.configureConn(conn)

	defer func() {
		mylog.Debug("server client close")
		server.opt.CloseCb(client)
		client.Close()
	}()

	server.opt.ConnectCb(client)
	if server.opt.AuthCb(client) != nil {
		return
	}
	go client.sendLoop()
	//server.sendChan <- client
	//client.notifySendChan = server.sendChan

	var t time.Duration
	deadTime := client.Get("deadline")
	if deadTime != nil {
		t = deadTime.(time.Duration)
	}

	/*
	if server.opt.RecvNum != 0 {
		read := func() {
			for {
				headerBuf := make([]byte, 8)
				if _, err := io.ReadFull(client.conn, headerBuf); err != nil {
					mylog.Debug(" srecv lopp err 1", err)
					return
				}
				header, err := client.packer.Unpack(headerBuf)
				if err != nil {
					mylog.Debug(" srecv lopp err 2", err)
					return
				}
				body := make([]byte, header.Len)
				if _, err := io.ReadFull(client.conn, body[:]); err != nil {
					mylog.Debug(" srecv lopp err 3", err)
					return
				}
				header.Msg = body
				mylog.Debug("recv header .... ", header)
				server.opt.MsgCb(client, header)
			}
		}

		for i := 0; i < server.opt.RecvNum; i++ {
			go read()
		}

	}
	*/
	for {
		if t != 0 {
			conn.SetDeadline(time.Now().Add(t))
		}
		m, err := client.readMessage()
		if err != nil {
			mylog.Debug("read smsg err : ", err)
			return
		}
		server.opt.MsgCb(client, m)
	}

}

