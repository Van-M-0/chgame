package main

import (
	"fmt"
	"net"
	"encoding/binary"
	"io"
	"msgpacker"
	"time"
	"os"
	"strings"
	"strconv"
	"net/http"
	"mylog"
	"gopkg.in/vmihailenco/msgpack.v2"
	"sync"
	"sync/atomic"
)
/*
func testpb() {
	p := &gameproto.Person{}

	p.Name = "hello"
	p.Phones = []*gameproto.Person_PhoneNumber {
		&gameproto.Person_PhoneNumber{
			Number: "123123",
			Type: gameproto.Person_PhoneType(123),
		},
		&gameproto.Person_PhoneNumber{
			Number: "4444",
			Type: gameproto.Person_PhoneType(333),
		},
	}

	myproto.Register(101, p)

	b, err := proto.Marshal(p)
	if err != nil {
		fmt.Println("proto marshal message error")
		return
	}

	fmt.Println(string(b))
	fmt.Println(p)

	i, _ := myproto.NewPbMessage(101)
	err = proto.Unmarshal(b, i.Msg.(proto.Message))
	if err != nil {
		fmt.Println("unmarsha error", err)
		return
	}
	fmt.Println("new message ", i)
}
*/
/*
func testmsg() {

	gs := &myproto.GatewayMessage{
		Direction: 1,
	}

	b, err := msgpack.Marshal(gs)
	if err != nil {
		fmt.Println("marshal msg error ", err)
		return
	}

	fmt.Println(b)
	fmt.Println(gs)

	gh, err := myproto.NewRawMessage(myproto.GatewayCmdHeader)
	if err != nil {
		fmt.Println("marshal msg error", err)
		return
	}

	err = msgpack.Unmarshal(b, gh)
	if err != nil {
		fmt.Println("unmarshal gateway heade error ", err)
	}

	msg, err := myproto.NewRawMessage(gh.(*myproto.GatewayMessage).Cmd)
	if err != nil {
		fmt.Println("new raw message rror")
	}

	newmsg := &myproto.GatewayMessage{
		Msg: msg,
	}
	err = msgpack.Unmarshal(b, newmsg)

	fmt.Println(newmsg.Msg)

}
*/

/*
func testmessage() {

	type hello struct {
		a int
		str string
	}

	b, err := msgpack.Marshal(&hello{a:1010, str:"wewewe"})
	if err != nil {
		return
	}

	m := myproto.Message{
		Cmd: 1012,
		Magic: 12321,
		Msg: b,
	}

	b1, _ := msgpack.Marshal(&m)

	var m1 myproto.Message
	msgpack.Unmarshal(b1, &m1)
	fmt.Println("afdsfsdfsf ", m1)

	fmt.Println("fsdfsd ------", m1.Msg)

}
*/

func test1() {
	type B struct {
		A 	int
	}

	data, err := msgpacker.Marshal(&B{
		A: 123,
	})

	fmt.Println(data, err)
}

func timetest() {
	TimeKeyFormat := "2006-01-02"
	t := time.Now().Format(TimeKeyFormat)
	fmt.Println("format time is ", t)

	t2, e:= time.Parse(TimeKeyFormat, "2017-10-3")
	ts := t2.Format(TimeKeyFormat)
	fmt.Println("parse time is ", t2, ts, e)

	fmt.Println("time comparation ", t > ts)


}

func closechantest() {
	ch := make(chan bool)
	close(ch)
}

func timeaddtest() {
	i := "2006-01-02"
	TimeKeyFormat := "2006-01-02"
	t, _ := time.Parse(TimeKeyFormat, i)
	t = t.Add(time.Duration(time.Hour) * 24)
	i = t.Format(TimeKeyFormat)

	fmt.Println("time i ", i)
}

func ticktest() {
	t := time.NewTicker(time.Duration(time.Second * 3))

	go func() {
		for {
			select {
			case <-t.C:
				fmt.Println("ticker comming")
			}
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

type roomTimerHandle struct {
	id 		int
	t 		*time.Timer
	kill 	chan bool
	stop	int32
}

func (h *roomTimerHandle) GetId() int {
	return h.id
}

func timerTest() {

	bcquit := make(chan bool)

	fmt.Println("time set ", time.Now())
	tm := time.NewTimer(time.Duration(300) * time.Millisecond * 10)
	handle := &roomTimerHandle{
		id:1,
		t: tm,
		kill: make(chan bool),
	}

	go func() {
		select {
		case <- bcquit:
			fmt.Println("time quit")
			return
		case <- tm.C:
			atomic.AddInt32(&handle.stop, 1)
			fmt.Println("time coming ", time.Now())
		case <- handle.kill:
			fmt.Println("time kill")
			atomic.AddInt32(&handle.stop, 1)
			close(handle.kill)
			return
		}
	}()

	handle.kill <- true

	fmt.Println("waiting")

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}


func main() {

	fmt.Println("xxxxxxxxxxxxxxxxx")

	timerTest()

	ticktest()

	if true {
		return
	}

	timeaddtest()

	closechantest()
	//timetest()



	if true {
		return
	}

	type AAATest struct {
		a 		int
	}

	mmaps := make(map[int]map[int]*AAATest)

	if _, ok := mmaps[1]; !ok {
		mmaps[1] = make(map[int]*AAATest)
	}
	mmaps[1][1] = &AAATest{
		a: 1111,
	}
	fmt.Println("mmaps ", mmaps[1][1])

	if true {
		return
	}


	logger := mylog.New()
	logger.Formatter = new(mylog.GameFormatter)
	logger.Out = os.Stdout
	logger.Warn("hello .........", []int{1, 2, 3})
	logger.Info("hello .........", []int{1, 2, 3})


	mylog.Error()

	downlaod := func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Disposition", "attachment; filename=WHATEVER_YOU_WANT")
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))

		file := "./file/a.txt"
		http.ServeFile(w, r, file)
	}

	os.Mkdir("file", 0777)
	//http.Handle("/pollux/", http.StripPrefix("/pollux/", http.FileServer(http.Dir("file"))))
	http.HandleFunc("/pollux/", downlaod)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("listen and serve error ", err)
	}

	tptr := time.NewTimer(time.Duration(3 * time.Second))
	t := *tptr

	select {
	case  <- t.C:
		fmt.Println("hello timer")
	}

	type AAA struct {
		M 		map[int]int
	}
	a := &AAA{
		M: make(map[int]int),
	}
	a.M[1] = 100

	data, err := msgpacker.Marshal(&a)
	fmt.Println(data, err)

	var a1 AAA
	msgpacker.UnMarshal(data, &a1)

	fmt.Println(a1, a1.M[1])

	str := strings.Split("1", ",")
	fmt.Println(str)
	r := make([]int, 0)
	for _, s := range str {
		if i, err := strconv.Atoi(s); err != nil {
			r = append(r, i)
		} else {
			fmt.Println(err)
		}
	}
	fmt.Println(r)

	/*
	test1()

	data := []byte {130, 164 ,100, 97, 116, 97, 129 ,163, 102, 115 ,100 ,205 ,1, 209, 163, 99, 109, 100, 205, 4, 106}

	type clisub struct{
		Fsd float64
	}
	type cli struct {
		Cmd 	float64
		Data 	clisub
	}

	dad, err := msgpacker.Marshal(&cli{
		Cmd:1130,
		Data: clisub{
			Fsd: 465,
		},
	})
	fmt.Println(dad, err)

	type d1 struct {
		fsd 	int
	}
	type md struct {
		Cmd	int
		Data d1
	}
	var m md
	err = msgpacker.UnMarshal(data, &m)
	fmt.Println(data, err, m)
	*/
	//testmsgpack()
	//testRpc()
	//testmessage()
	//testmsg()

	//tcpServer()

	//send1(nil)
	//testredis()
	//testps()

	//testCommunicator()
	//prototest()

	//testproto2()

	//start_test()
	//testMq()

	//genMysqlTables()

	//test_log()
	orm_test()

	signalChan := make(chan os.Signal, 1)
	go func() {
		//阻塞程序运行，直到收到终止的信号
		<-signalChan
		println(".........")
		os.Exit(0)
	}()

	time.Sleep(time.Duration(10) * time.Second)
	println("exit")

}


func tcpServer() {
	fmt.Println("tcp server port ", 9895)
	l, err := net.Listen("tcp", ":9895")
	if err != nil {
		fmt.Println("listen error ", err)
		return
	}

	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("accept client error ", err)
			return
		}

		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer func() {
		conn.Close()
	}()

	byHeader := make([]byte, 2)
	byBody := make([]byte, 65534)

	for {
		if _, err := io.ReadFull(conn, byHeader); err != nil {
			fmt.Println("client read header error")
			return
		}
		fmt.Println("recv header len ", byHeader)
		msgLen := binary.BigEndian.Uint16(byHeader)
		fmt.Println("recv msg len ", msgLen)
		if _, err := io.ReadFull(conn, byBody[:msgLen]); err != nil {
			fmt.Println("client read body error")
			return
		}

		codecClient(byBody[:msgLen])

		send1(conn)
	}
}

func send1(conn net.Conn) {

	type conf struct {
		A 		string
		I 		int
	}

	type simple2 struct {
		Conf 		conf
		Arr 		[3]int
		Slice 		[]string
	}

	//myproto.Register(1222, (*simple2)(nil))

	s2 := &simple2{
		Arr: [3]int{123, 456, 789},
		Slice: []string{"hello", "你好"},
		Conf: conf{
			A: "conf-string",
			I: 89,
		},
	}

	b, err := msgpack.Marshal(s2)
	if err != nil {
		fmt.Println("marsha eror ", err)
		return
	}


	byHeader := make([]byte, 2)
	binary.BigEndian.PutUint16(byHeader, uint16(len(b)))

	body := make([]byte, 2 + len(b))
	copy(body[:2], byHeader)
	copy(body[2:], b)

	fmt.Println("header " , byHeader)
	fmt.Println("body ", b)
	//body = append(body, byHeader...)
	//body = append(body, b...)

	n, err := conn.Write(body)
	if err != nil {
		fmt.Println("write bytes ", b)
	}
	fmt.Println("send msg ", s2)
	fmt.Println("send header data ", byHeader)
	fmt.Println("send data ", n, body)

	/*
	ns2, _:= myproto.NewRawMessage(1222)

	err = msgpack.Unmarshal(b, ns2)

	cns2 := ns2.(*simple2)

	fmt.Println("unmarsha error ", err, cns2.Slice, cns2.Arr, cns2.Conf.A, cns2.Conf.I)
	*/
}

/*
func codec1(msg []byte) {
	type simple struct {
		I 		int
		Ui 		uint
		A 		string
	}

	myproto.Register(1010, (*simple)(nil))

	s1, _:= myproto.NewRawMessage(1010)

	msgpack.Unmarshal(msg, s1)

	nmsg, ok := s1.(*simple)
	if !ok {
		fmt.Println("cast simple error")
	}

	fmt.Println("unmarsha simple message s1 ", nmsg.I, nmsg.A, nmsg.Ui)
}

func codec2(msg []byte) {

}
*/

func codecClient(msg []byte) {
	//codec1(msg)
}
