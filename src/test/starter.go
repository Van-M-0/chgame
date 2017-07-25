package main

import (
	"os"
	"fmt"
	"exportor/defines"
	"exportor/proto"
	"game"
	"time"
	"runtime/debug"
	"starter"
)

var register = make(map[string]interface{})

var _ = game.NewGameServer

func init() {
	register[defines.ChannelLoadUser] = proto.PMLoadUser{}
	register[defines.ChannelCreateAccountFinish] = proto.PMLoadUserFinish{}
	register[defines.ChannelCreateAccount] = proto.PMCreateAccount{}
	register[defines.ChannelCreateAccountFinish] = proto.PMCreateAccountFinish{}
}

func safacall(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recover enter")
			fmt.Println("recover err : " , r)
			debug.PrintStack()
		}
		fmt.Println("recover test goroutine quit")
	}()
	fn()
}

func safeGo(fn func()) {
	go func() {
		defer func() {
			fmt.Println("save go exit")
		}()
		time.Sleep(time.Duration(1) * time.Second)
		for {
			safacall(fn)
			time.Sleep(time.Duration(1) * time.Second)
		}
	}()
}

func badcall() {
	fmt.Println("bad call")
	var arr []byte
	arr[3] = 3
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

	/*
	safeGo(badcall)

	for {
		time.Sleep(time.Duration(1) * time.Second)
		fmt.Println("main programa run")
	}
	*/

	p := os.Args[1]
	fmt.Println("start args ", p)

	starter.StartProgram(p, nil)
}
