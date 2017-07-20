package main

import (
	"os"
	"fmt"
	"starter"
	"exportor/defines"
	"exportor/proto"
)

var register = make(map[string]interface{})

func init() {
	register[defines.ChannelLoadUser] = proto.PMLoadUser{}
	register[defines.ChannelCreateAccountFinish] = proto.PMLoadUserFinish{}
	register[defines.ChannelCreateAccount] = proto.PMCreateAccount{}
	register[defines.ChannelCreateAccountFinish] = proto.PMCreateAccountFinish{}
}

func start_test() {
	/*
	type atest struct {
		d 	[]chan int
	}
	a := &atest{
		d: make([]chan int, 3),
	}
	for i := 0; i < 3; i++ {
		a.d[i] = make(chan int, 3)
		a.d[i] <- i
	}

	for i := 0; i < 3; i++ {
		fmt.Println("chan int ", <- a.d[i])
	}
	*/

	p := os.Args[1]
	fmt.Println("start args ", p)

	starter.StartProgram(p)
}
