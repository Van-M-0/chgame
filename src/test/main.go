package main

import (
	"github.com/golang/protobuf/proto"
	"fmt"
	"gameproto/clipb"
	myproto "exportor/proto"
	"gopkg.in/vmihailenco/msgpack.v2"
)

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

func main() {
	testmessage()
	//testmsg()
}

