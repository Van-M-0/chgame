package network

import (
	"exportor/defines"
	"net"
	"errors"
	"fmt"
	"io"
	"msgpacker"
	"exportor/proto"
	"sync/atomic"
	"time"
	"sync"
	"mylog"
)


type message struct {
	cmd 	uint32
	data 	interface{}
}

type tcpClient struct {
	*netContext
	id 				uint32
	conn 			net.Conn
	opt 			*defines.NetClientOption
	sendCh			chan *message
	headerBuf 		[8]byte
	packer 			*msgpacker.MsgPacker
	authed 			bool
	quit 			chan bool
	stoped 			int32
	lastInTime 		time.Time
	notifySendChan 	chan *tcpClient
	notified 		bool
	notifyLock		sync.Mutex
}

func newTcpClient(opt *defines.NetClientOption) *tcpClient {
	mylog.Debug("new client send chan size ", opt.SendChSize)
	client := &tcpClient{
		opt: opt,
		sendCh: make(chan *message, opt.SendChSize),
		quit: make(chan bool),
		packer: msgpacker.NewMsgPacker(),
		netContext: newNetContext(),
	}
	return client
}

func (client *tcpClient) Id(id uint32) {
	client.id = id
}

func (client *tcpClient) GetId() uint32 {
	return client.id
}

func (client *tcpClient) GetRemoteAddress() string {
	return client.conn.RemoteAddr().String()
}

func (client *tcpClient) Connect() error {
	conn, err := net.Dial("tcp", client.opt.Host)
	if err != nil {
		return err
	}
	client.conn = conn
	client.opt.ConnectCb(client)
	go client.sendLoop()
	if client.opt.AuthCb(client) != nil {
		client.Close()
		return errors.New("connect auth error")
	}
	client.recvLoop()

	return nil
}

func (client *tcpClient) Close() error {
	if client.opt.CloseCb != nil {
		client.opt.CloseCb(client)
	}
	if err := client.conn.Close(); err != nil {
		mylog.Debug("client.close error ", err)
	}
	atomic.AddInt32(&client.stoped, 1)
	//for i := 0; i < client.opt.SendActor + 1; i++ {
		client.quit <- true
	//}
	mylog.Debug("close client ...")
	close(client.sendCh)
	for {
		if value, ok := <-client.sendCh; ok {
			mylog.Debug("send chan value ", value)
		} else {
			break
		}
	}
	mylog.Debug("client.closecb ")
	return nil
}

func (client *tcpClient) IsClosed() bool { return atomic.LoadInt32(&client.stoped) != 0 }

func (client *tcpClient) Send(cmd uint32, data interface{}) error {
	if atomic.LoadInt32(&client.stoped) != 0 {
		mylog.Debug("cs .", client.GetId())
		return nil
	}
	if client.notifySendChan != nil {

		if len(client.sendCh) > client.opt.SendChSize - 3 {
			mylog.Debug("send chan size ", client.GetId(), len(client.sendCh))
		}

		client.sendCh <- &message{cmd: cmd, data: data}

		/*
		client.notifyLock.Lock()
		if client.notified && len(client.sendCh) == 0 {
			client.notifyLock.Unlock()
			return nil
		}
		client.notifyLock.Unlock()
		*/
		client.notified = true
		client.notifySendChan <- client

		return nil
	}
	if len(client.sendCh) > client.opt.SendChSize - 3 {
		mylog.Debug("==================== ", client.GetId(), len(client.sendCh))
		mylog.Debug("send chan size ", client.GetId(), len(client.sendCh))
		mylog.Debug("==================== ", client.GetId(), len(client.sendCh))
	}
	client.sendCh <- &message{cmd: cmd, data: data}
	return nil
}

func (client *tcpClient) FlushSendBuffer() int {
	count := len(client.sendCh)
	data := make([]byte, 0)
	for k := 0; k < count; k++ {
		m := <- client.sendCh
		//mylog.Debug("flush client packet ", m.cmd, m.data)
		if raw, err :=client.packer.Pack(m.cmd, m.data); err == nil {
			data = append(data, raw...)
		} else {
			mylog.Debug("send pack data err ", err)
		}
	}
	if len(data) > 0 {
		//mylog.Debug("ssend actor send data ------: ", client.GetId(), count, len(data), time.Now())
		if _, err := client.conn.Write(data); err != nil {
			mylog.Debug("write data error ", err)
			return 0
		}
		//mylog.Debug("swf ", time.Now())
		//} else {
		//	mylog.Debug("not data send", len(client.sendCh))
	}
	return count
}

var clientSendCounter [300]int32
var staticLastTime time.Time
var startTime time.Time
func (client *tcpClient) sendLoop() {
	//mylog.Debug("send looop start")

	send := func() {
		//t := time.NewTimer(time.Second * 30)
		gt, ge := time.Now(), time.Now()
		now := time.Now()
		p1, p2 := now,now
		p3, p4 := now,now
		t := time.Tick(time.Millisecond * 80)
		for {
			if ge.Sub(gt) >= time.Second * 1 {
				//mylog.Debug("client ---xxxxxxxxxxxxxxxxxxxxxxx", ge.Sub(gt), ge, gt, time.Now(), client.GetId(), len(client.sendCh))
				(fmt.Sprintf("xxxxxxxxx id %v, p:(%v %v %v) a(%v %v %v %v) l(%v)", client.GetId(), ge.Sub(gt), gt, ge, p1, p2, p3, p4, len(client.sendCh)))
			}
			//mylog.Debug("send loop scheldu ", ge.Sub(gt))
			gt = time.Now()
			select {
			case <- client.quit:
				mylog.Debug("c ",client.GetId(), len(client.sendCh))
				return
				/*
			case <-t.C:
				mylog.Debug("prof tout " ,time.Now().Sub(wst))
				if time.Now().Sub(wst) > time.Second * 4 {
					mylog.Debug("write not return ", wst, wet, time.Now())
					return
				}
				*/
			//case m:= <- client.sendCh:
			case <-t :

				p1 = time.Now()
				count := len(client.sendCh)
				data := make([]byte, 0)

				/*
				if raw, err :=client.packer.Pack(m.cmd, m.data); err == nil {
					data = append(data, raw...)
				} else {
					mylog.Debug("send pack data err ", err)
				}
				*/
				p2 = time.Now()

				for k := 0; k < count; k++ {
					m := <- client.sendCh
					if raw, err :=client.packer.Pack(m.cmd, m.data); err == nil {
						mylog.Debug("packet ", m.cmd, client.GetId())
						data = append(data, raw...)
					} else {
						mylog.Debug("send pack data err ", err)
					}
				}
				p3 = time.Now()
				if startTime.Second() == 0 {
					startTime = p3
				}
				if len(data) > 0 {
					//atomic.AddInt32(&clientSendCounter[int(time.Now().Sub(startTime).Seconds())], int32(count))
					//mylog.Debug("client send actor send data ------: ", client.GetId(), count, len(data), time.Now())
					if _, err := client.conn.Write(data); err != nil {
						mylog.Debug("write data error ", err)
						return
					}
					//mylog.Debug("wf ", client.GetId(), time.Now())
				//} else {
				//	mylog.Debug("client not data send", len(client.sendCh))
				}
				if staticLastTime.Second() == 0 {
					staticLastTime = time.Now()
				}
				p4 = time.Now()
				if p4.Sub(staticLastTime).Seconds() >= 1 {
					//mylog.Debug("client send counter ", clientSendCounter)
					staticLastTime = p4
				}
				/*
				if raw, err :=client.packer.Pack(m.cmd, m.data); err != nil {
				} else {
					if _, err = client.conn.Write(raw); err != nil {
						return
					}
				}
				count := len(client.sendCh)
				for i := 0; i < count; i++ {
					m := <- client.sendCh
					if raw, err :=client.packer.Pack(m.cmd, m.data); err != nil {
					} else {
						if _, err = client.conn.Write(raw); err != nil {
							return
						}
					}
				}
				*/
			}
			ge = time.Now()
		}
	}
	if client.opt.SendActor > 0 {
		//var xxxcounter [60]int32

		go func() {
			for {
				select {
				case <- time.After(time.Second * 2):
					//mylog.Debug("send ...... ", xxxcounter)
				}
			}
		}()

		sf := func () {

			defer func() {
				mylog.Debug("send loop error ? ")
			}()
			t := time.Tick(time.Millisecond * 30)
			for {
				select {
				case <- client.quit:
					mylog.Debug("server send actor quit")
					return
				case <- t:
					if !client.IsClosed() {
						if len(client.sendCh) > 0 {
							//mylog.Debug("f 1")
							count := client.FlushSendBuffer()
							//mylog.Debug("f 2")
							//atomic.AddInt32(&xxxcounter[time.Now().Second()], int32(count))
							tick := 0
							for cur := len(client.sendCh); cur > 30;  {
								//mylog.Debug("f 3")
								count += client.FlushSendBuffer()
								//mylog.Debug("f 4")
								//atomic.AddInt32(&xxxcounter[time.Now().Second()], int32(count))
								//mylog.Debug("packet still more than > 50", client.GetId())
								tick++
								if tick % 30 == 0 { break }
							}
							if len(client.sendCh) > 10 {
								//mylog.Debug("ser chan is ", len(client.sendCh))
							}
						}
					} else {
						mylog.Debug("client is closed ?")
					}
					/*
					count := len(client.sendCh)
					data := make([]byte, 0)
					for k := 0; k < count; k++ {
						m := <- client.sendCh
						if raw, err :=client.packer.Pack(m.cmd, m.data); err == nil {
							data = append(data, raw...)
						} else {
							mylog.Debug("send pack data err ", err)
						}
					}
					atomic.AddInt32(&xxxcounter[time.Now().Second()], int32(count))
					tick++
					if tick % 30 == 0 {
						mylog.Debug("ssend conn ", xxxcounter)
					}
					if len(data) > 0 {
						//mylog.Debug("ssend actor send data ------: ", client.GetId(), count, len(data), time.Now())
						if _, err := client.conn.Write(data); err != nil {
							mylog.Debug("write data error ", err)
							return
						}
						//mylog.Debug("swf ", time.Now())
					//} else {
					//	mylog.Debug("not data send", len(client.sendCh))
					}
					*/
				}
			}
		}
		for i := 0; i < 0; i++ {
			go sf()
		}
		sf()
	} else {
		send()
	}
}

func (client *tcpClient) recvLoop() {
	client.authed = true
	/*
	read := func() {
		for {
			headerBuf := make([]byte, 8)
			if _, err := io.ReadFull(client.conn, headerBuf); err != nil {
				mylog.Debug("recv lopp err", err)
				return
			}
			header, err := client.packer.Unpack(headerBuf)
			if err != nil {
				mylog.Debug("recv lopp err 1", err)
				return
			}
			body := make([]byte, header.Len)
			if _, err := io.ReadFull(client.conn, body[:]); err != nil {
				mylog.Debug("recv lopp err 2", err)
				return
			}
			header.Msg = body
			mylog.Debug("recv loop ", header)
			client.opt.MsgCb(client, header)
		}
	}
	for i := 0; i < 10; i ++ {
		go read()
	}
	*/

	go func() {
		defer func() {
			client.Close()
		}()

		for {
			m, err := client.readMessage()
			if err != nil {
				mylog.Debug("get msg err : ", err)
				return
			}
			client.opt.MsgCb(client, m)
		}
	}()
}

func (client *tcpClient) Auth() (*proto.Message, error) {
	return client.readMessage()
}

func (client *tcpClient) readMessage() (*proto.Message, error) {
	if _, err := io.ReadFull(client.conn, client.headerBuf[:]); err != nil {
		return nil, err
	}
	header, err := client.packer.Unpack(client.headerBuf[:])
	if err != nil {
		return nil, err
	}

	body := make([]byte, header.Len)
	if _, err := io.ReadFull(client.conn, body[:]); err != nil {
		return nil, err
	}
	header.Msg = body
	return header, nil
}

func (client *tcpClient) configureConn(conn net.Conn) {
	client.conn = conn
}
