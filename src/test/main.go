package main

import (
	"github.com/golang/protobuf/proto"
	"fmt"
	"gameproto/clipb"
	myproto "exportor/proto"
)

func main() {

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

