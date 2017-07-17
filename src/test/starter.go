package main

import (
	"os"
	"fmt"
	"starter"
	"sync"
)

func start_test() {

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
