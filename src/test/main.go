package main

import (
	"fmt"
	"gopkg.in/vmihailenco/msgpack.v2"
	"net"
	"encoding/binary"
	"io"
	"msgpacker"
)
/*
func testpb() {
	p := &gameproto.Person{}

	p.Name = "hello"
	p.Phones = []*gameproto.Person_PhoneNumber {
		&gameproto.Person_PhoneNumber{
			Number: "123123",
			Type: gameproto.Person_PhoneType(123),
		},
		&gameproto.Person_PhoneNumber{
			Number: "4444",
			Type: gameproto.Person_PhoneType(333),
		},
	}

	myproto.Register(101, p)

	b, err := proto.Marshal(p)
	if err != nil {
		fmt.Println("proto marshal message error")
		return
	}

	fmt.Println(string(b))
	fmt.Println(p)

	i, _ := myproto.NewPbMessage(101)
	err = proto.Unmarshal(b, i.Msg.(proto.Message))
	if err != nil {
		fmt.Println("unmarsha error", err)
		return
	}
	fmt.Println("new message ", i)
}
*/
/*
func testmsg() {

	gs := &myproto.GatewayMessage{
		Direction: 1,
	}

	b, err := msgpack.Marshal(gs)
	if err != nil {
		fmt.Println("marshal msg error ", err)
		return
	}

	fmt.Println(b)
	fmt.Println(gs)

	gh, err := myproto.NewRawMessage(myproto.GatewayCmdHeader)
	if err != nil {
		fmt.Println("marshal msg error", err)
		return
	}

	err = msgpack.Unmarshal(b, gh)
	if err != nil {
		fmt.Println("unmarshal gateway heade error ", err)
	}

	msg, err := myproto.NewRawMessage(gh.(*myproto.GatewayMessage).Cmd)
	if err != nil {
		fmt.Println("new raw message rror")
	}

	newmsg := &myproto.GatewayMessage{
		Msg: msg,
	}
	err = msgpack.Unmarshal(b, newmsg)

	fmt.Println(newmsg.Msg)

}
*/

/*
func testmessage() {

	type hello struct {
		a int
		str string
	}

	b, err := msgpack.Marshal(&hello{a:1010, str:"wewewe"})
	if err != nil {
		return
	}

	m := myproto.Message{
		Cmd: 1012,
		Magic: 12321,
		Msg: b,
	}

	b1, _ := msgpack.Marshal(&m)

	var m1 myproto.Message
	msgpack.Unmarshal(b1, &m1)
	fmt.Println("afdsfsdfsf ", m1)

	fmt.Println("fsdfsd ------", m1.Msg)

}
*/

func test1() {
	type B struct {
		A 	int
	}

	data, err := msgpacker.Marshal(&B{
		A: 123,
	})

	fmt.Println(data, err)
}

func main() {

	/*
	test1()

	data := []byte {130, 164 ,100, 97, 116, 97, 129 ,163, 102, 115 ,100 ,205 ,1, 209, 163, 99, 109, 100, 205, 4, 106}

	type clisub struct{
		Fsd float64
	}
	type cli struct {
		Cmd 	float64
		Data 	clisub
	}

	dad, err := msgpacker.Marshal(&cli{
		Cmd:1130,
		Data: clisub{
			Fsd: 465,
		},
	})
	fmt.Println(dad, err)

	type d1 struct {
		fsd 	int
	}
	type md struct {
		Cmd	int
		Data d1
	}
	var m md
	err = msgpacker.UnMarshal(data, &m)
	fmt.Println(data, err, m)
	*/
	//testmsgpack()
	//testRpc()
	//testmessage()
	//testmsg()

	//tcpServer()

	//send1(nil)
	//testredis()
	//testps()

	//testCommunicator()
	//prototest()

	//testproto2()

	start_test()
	//testMq()

	//genMysqlTables()

	//test_log()
}


func tcpServer() {
	fmt.Println("tcp server port ", 9895)
	l, err := net.Listen("tcp", ":9895")
	if err != nil {
		fmt.Println("listen error ", err)
		return
	}

	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("accept client error ", err)
			return
		}

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer func() {
		conn.Close()
	}()

	byHeader := make([]byte, 2)
	byBody := make([]byte, 65534)

	for {
		if _, err := io.ReadFull(conn, byHeader); err != nil {
			fmt.Println("client read header error")
			return
		}
		fmt.Println("recv header len ", byHeader)
		msgLen := binary.BigEndian.Uint16(byHeader)
		fmt.Println("recv msg len ", msgLen)
		if _, err := io.ReadFull(conn, byBody[:msgLen]); err != nil {
			fmt.Println("client read body error")
			return
		}

		codecClient(byBody[:msgLen])

		send1(conn)
	}
}

func send1(conn net.Conn) {

	type conf struct {
		A 		string
		I 		int
	}

	type simple2 struct {
		Conf 		conf
		Arr 		[3]int
		Slice 		[]string
	}

	//myproto.Register(1222, (*simple2)(nil))

	s2 := &simple2{
		Arr: [3]int{123, 456, 789},
		Slice: []string{"hello", "你好"},
		Conf: conf{
			A: "conf-string",
			I: 89,
		},
	}

	b, err := msgpack.Marshal(s2)
	if err != nil {
		fmt.Println("marsha eror ", err)
		return
	}


	byHeader := make([]byte, 2)
	binary.BigEndian.PutUint16(byHeader, uint16(len(b)))

	body := make([]byte, 2 + len(b))
	copy(body[:2], byHeader)
	copy(body[2:], b)

	fmt.Println("header " , byHeader)
	fmt.Println("body ", b)
	//body = append(body, byHeader...)
	//body = append(body, b...)

	n, err := conn.Write(body)
	if err != nil {
		fmt.Println("write bytes ", b)
	}
	fmt.Println("send msg ", s2)
	fmt.Println("send header data ", byHeader)
	fmt.Println("send data ", n, body)

	/*
	ns2, _:= myproto.NewRawMessage(1222)

	err = msgpack.Unmarshal(b, ns2)

	cns2 := ns2.(*simple2)

	fmt.Println("unmarsha error ", err, cns2.Slice, cns2.Arr, cns2.Conf.A, cns2.Conf.I)
	*/
}

/*
func codec1(msg []byte) {
	type simple struct {
		I 		int
		Ui 		uint
		A 		string
	}

	myproto.Register(1010, (*simple)(nil))

	s1, _:= myproto.NewRawMessage(1010)

	msgpack.Unmarshal(msg, s1)

	nmsg, ok := s1.(*simple)
	if !ok {
		fmt.Println("cast simple error")
	}

	fmt.Println("unmarsha simple message s1 ", nmsg.I, nmsg.A, nmsg.Ui)
}

func codec2(msg []byte) {

}
*/

func codecClient(msg []byte) {
	//codec1(msg)
}
