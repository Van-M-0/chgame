package main

import (
	"communicator"
	"exportor/defines"
	"fmt"
	"bytes"
)

func testCommunicator() {

	type Hello struct {
		A 		int
		Str 	string
		Slic 	[]int
	}

	var buf bytes.Buffer
	fmt.Fprint(&buf, &Hello{
		A:100,
		Slic: []int{123, 3},
		Str: "hello world",
	})

	fmt.Println(buf.Bytes(), string(buf.Bytes()))

	fmt.Println("test comunicaor")
	c := communicator.NewCommunicator(&defines.CommunicatorOption{

	})
	c.JoinChanel("hello", false, func(data []byte) {
		fmt.Println("hello channel ", string(data))
	})
	c.JoinChanel("h*", true, func(data []byte) {
		fmt.Println("hello channel ", string(data))
	})

	c.Notify("hello", "world")
}