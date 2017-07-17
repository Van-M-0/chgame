package main

/*
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

func testproto2() {
	type a1 struct {
		A 	string
		B 	int
	}

	a := &a1 {
		A: "hello",
		B: 123,
	}

	b, err := msgpack.Marshal(a)
	fmt.Println("fsdfsd", err, b)

	var i interface{}
	i = &a1 {}

	err = msgpack.Unmarshal(b, i)
	fmt.Println("111", err, i)

	a2 := i.(*a1)
	fmt.Println(a2)

}
*/