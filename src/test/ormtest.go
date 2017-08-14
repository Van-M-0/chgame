package main

import (
	"fmt"
	"runtime/debug"
	"dbproxy"
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
	dbproxy.Test()


}