package main

import (
	"time"
	"fmt"
	"mylog"
)

func test_log() {
	t := time.Now().Unix()
	fmt.Println("time is ", t)
	a := time.Time{}

	mylog.Debug	(a.Unix())
}