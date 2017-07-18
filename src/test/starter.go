package main

import (
	"os"
	"fmt"
	"starter"
	"sync"
)

func start_test() {

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

	p := os.Args[1]
	fmt.Println("start args ", p)

	if p == "client" {
		starter.StartClient()
	} else if p == "lobby" {
		starter.StartLobby()
	} else if p == "gate" {
		starter.StartGate()
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)
	wg.Wait()
}
