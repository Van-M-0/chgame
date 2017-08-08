package main

import (
	"fmt"
	"runtime/debug"
	"time"
	"sync"
)

func sfroutine(fn func()) {
	safeCall := func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("------save data error------")
				fmt.Println(err)
				debug.PrintStack()
				fmt.Println("--------------------------")
			}
		}()
		fmt.Println("call function")
		fn()
	}

	go func() {
		fmt.Println("call go function")
		safeCall()
	}()
}

func sayhello() {
	fmt.Println("print hello")
}

func orm_test() {
	//dbproxy.Test()

	timeoutFn := func(tm time.Duration, fn func()) {
		for {
			fmt.Println("for{} .......")
			t := time.NewTimer(time.Second * tm)
			select {
			case <- t.C:
				fn()
			}
		}
	}

	sfroutine(func() {
		timeoutFn(1, sayhello)
	})

	fmt.Println("waiting.......")
	wg := new(sync.WaitGroup)
	wg.Add(1)
	wg.Wait()
}