package main

import (
	"time"
	"fmt"
)

func test_log() {
	t := time.Now().Unix()
	fmt.Println("time is ", t)
	a := time.Time{}

	fmt.Println	(a.Unix())
}