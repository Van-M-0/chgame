package main

import (
	"exportor/proto"
	"gopkg.in/vmihailenco/msgpack.v2"
	"fmt"
	"reflect"
)

func prototest() {

	type hello struct {
		Name 	string
		Id 		int
	}

	m := &proto.Message{
		Cmd: 1010,
		Msg: &hello{
			Name: "hello",
			Id: 999,
		},
	}

	b, err := msgpack.Marshal(m)
	if err != nil {
		fmt.Println("marshal struct err ", err)
	}

	fmt.Println("marshal result : ",err, b)

	m1 := &proto.Message{
		Msg: &hello{

		},
	}
	err = msgpack.Unmarshal(b, m1)
	if err != nil {
		fmt.Println("...... marshal struct err ", err)
	}

	m2, _ := m1.Msg.(*hello)

	type a1 struct  {
		int32
	}

	fmt.Println("marshal result .....: ",m2, b)

	mm, _ := proto.NewRawMessage(proto.GameCmdPlayerLogin)
	fmt.Println(reflect.TypeOf(mm))
}
