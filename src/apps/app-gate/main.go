package main

import (
	"starter"
	"os"
	"strings"
	"fmt"
	"runtime"
	"runtime/pprof"
	"os/signal"
	"net/http"
	"strconv"
	_ "net/http/pprof"
)

var (
	pid      int
	progname string
)

func init() {
	pid = os.Getpid()
	paths := strings.Split(os.Args[0], "/")
	paths = strings.Split(paths[len(paths)-1], string(os.PathSeparator))
	progname = paths[len(paths)-1]

	runtime.MemProfileRate = 1

	fmt.Println("pid grogram ", pid, progname)
}

func saveHeapProfile() {
	/*
	fm, err := os.OpenFile("./gate_mem.out", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("create file error")
	}
	pprof.WriteHeapProfile(fm)
	fm.Close()
	*/
	//runtime.GC()

	{
		f, err := os.OpenFile("./gate_mem.prof", os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return
		}
		pprof.WriteHeapProfile(f)
		defer f.Close()
	}

	{
		f, err := os.OpenFile("./gate_cpu.prof", os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
}

func waitForSignal() os.Signal {
	signalChan := make(chan os.Signal, 1)
	defer close(signalChan)

	signal.Notify(signalChan, os.Kill, os.Interrupt)
	s := <-signalChan
	signal.Stop(signalChan)
	return s
}

func main() {
	go func() {
		http.ListenAndServe("localhost:9999", nil)
	}()

	go func() {
		http.HandleFunc("/go", func(w http.ResponseWriter, r *http.Request) {
			num := strconv.FormatInt(int64(runtime.NumGoroutine()), 10)
			w.Write([]byte(num))
		})
		http.HandleFunc("/mem", func(w http.ResponseWriter, r *http.Request) {
			f, err := os.OpenFile("./gate_mem.prof", os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				return
			}
			pprof.WriteHeapProfile(f)
			defer f.Close()
		})
		http.HandleFunc("/cpu", func(w http.ResponseWriter, r *http.Request) {
			f, err := os.OpenFile("./gate_cpu.prof", os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				return
			}
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		})
		http.ListenAndServe("localhost:6060", nil)
	}()

	defer func() {
		fmt.Println("hello defer")
		saveHeapProfile()
	}()
	starter.StartProgram("gate",nil)

	fmt.Println("wati for siginal.")
	s := waitForSignal()
	starter.StopGate()

	fmt.Printf("signal got: %v, all users: %d\n", s)
}
