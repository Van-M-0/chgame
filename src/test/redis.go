package main

import (
	"github.com/garyburd/redigo/redis"
	"log"
	"reflect"
	"sync"
	"fmt"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type Conf struct {
	Host 		string
}

type Stock struct {
	CompanyName string `redis:"company_name"`
	OpenPrice   string `redis:"open_price"`
	AskPrice    string `redis:"ask_price"`
	ClosePrice  string `redis:"close_price"`
	BidPrice    string
	Id 			int    `redis:"id"`
}

func testscan() {
	conn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		log.Fatalf("Couldn't connect to Redis: %v\n", err)
	}
	defer conn.Close()

	stockData := map[string]*Stock{
		"GOOG": &Stock{CompanyName: "Google Inc.", OpenPrice: "803.99", AskPrice: "795.50", ClosePrice: "802.66", BidPrice: "793.36", Id : 100,

		},
		"MSFT": &Stock{AskPrice: "N/A", OpenPrice: "28.30", CompanyName: "Microsoft Corpora", BidPrice: "28.50", ClosePrice: "28.37", Id : 200},
	}

	for sym, row := range stockData {
		if _, err := conn.Do("HMSET", redis.Args{sym}.AddFlat(row)...); err != nil {
			log.Fatal(err)
		}
	}

	for sym := range stockData {
		values, err := redis.Values(conn.Do("HGETALL", sym))
		if err != nil {
			log.Fatal(err)
		}
		var stock Stock
		if err := redis.ScanStruct(values, &stock); err != nil {
			log.Fatal(err)
		}
		log.Printf("%s: %+v", sym, &stock, reflect.TypeOf(stock.Id))
	}
}

func testredis() {
	fmt.Println("test redis ..........")
	conn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		log.Fatalf("Couldn't connect to Redis: %v\n", err)
	}
	defer conn.Close()

	conn.Do("set", "a13213", 123123, "ex", 60)

	/*
	type HelloWorld struct {
		A 		int
		Str 	string
	}

	data, err := msgpack.Marshal(&HelloWorld{A:97, Str:"hello world"})
	conn.Do("set", "record", data)

	br, e := redis.Bytes(conn.Do("get", "record"))
	fmt.Println("get record reply ", br, e)

	var h HelloWorld
	err = msgpacker.UnMarshal(br, &h)
	fmt.Println(h,  err)

	keys, err := redis.Strings(conn.Do("keys", "ak.*"))
	fmt.Println("keys ..... ", keys)

	sort.Strings(keys)

	fmt.Println("keys ..... ", keys)
	*/

	/*
	cc := cacher.NewCacheClient("a")
	cc.Start()

	rid := cc.SaveGameRecord([]byte{12,12}, []byte{34, 34})
	cc.SaveUserRecord(100, rid)
	cc.SaveUserRecord(100, rid+1)
	cc.SaveUserRecord(101, rid)

	m, err := cc.GetGameRecordHead(100)
	fmt.Println("user record ", m, err)
	c, err := cc.GetGameRecordContent(rid)
	fmt.Println("content ", c, err)
	*/

	/*
	servers := map[string]*proto.CacheServer {
		"_SER_server1": &proto.CacheServer{
			Type: "server1",
			Id: 1,
			OnlineCount: 100,
		},
		"_SER_server2": &proto.CacheServer{
			Type: "server2",
			Id: 2,
			OnlineCount: 3,
		},
	}

	for key, val := range servers {
		if _, err := conn.Do("HMSET", redis.Args{key}.AddFlat(val)...); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("set beign ............")
	var si proto.CacheServer
	si.Id = 1
	arg := redis.Args{"_SER_server1"}.AddFlat(si)
	fmt.Println("arg ..", arg)
	if _, err := conn.Do("HMSET", arg...); err != nil {
		log.Fatal(err)
	}

	conn.Do("hset", "_SER_server2", "id", 100)
	*/
}

func publishm(channel string, args ...interface{}) {
	c, err := redis.Dial("tcp", ":6379")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()
	c.Do("PUBLISH", channel, args)
}

func publish(channel, value interface{}) {
	c, err := redis.Dial("tcp", ":6379")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()
	c.Do("PUBLISH", channel, value)
}

func testps() {
	c, err := redis.Dial("tcp", ":6379")
	if err != nil {
		log.Fatalf("Couldn't connect to Redis: %v\n", err)
	}
	defer c.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	psc := redis.PubSubConn{Conn: c}


	// This goroutine receives and prints pushed notifications from the server.
	// The goroutine exits when the connection is unsubscribed from all
	// channels or there is an error.
	go func() {
		defer wg.Done()
		for {
			switch n := psc.Receive().(type) {
			case redis.Message:
				fmt.Printf("Message: %s %s\n", n.Channel, n.Data)

				var str error
				msgpack.Unmarshal(n.Data, &str)
				fmt.Println(str)
			case redis.PMessage:
				fmt.Printf("PMessage: %s %s %s\n", n.Pattern, n.Channel, n.Data)
			case redis.Subscription:
				fmt.Printf("Subscription: %s %s %d\n", n.Kind, n.Channel, n.Count)
				if n.Count == 0 {
					return
				}
			case error:
				fmt.Printf("error: %v\n", n)
				return
			}
		}
	}()

	// This goroutine manages subscriptions for the connection.
	go func() {
		defer wg.Done()

		psc.Subscribe("example")
		psc.PSubscribe("p*")

		// The following function calls publish a message using another
		// connection to the Redis server.
		type A struct {
			A1 int
			Str string
		}
		dat, _ := msgpack.Marshal(nil)
		publish("example",  dat)


		// Unsubscribe from all connections. This will cause the receiving
		// goroutine to exit.
		psc.Unsubscribe()
		psc.PUnsubscribe()
	}()

	wg.Wait()

	// Output:
	// Subscription: subscribe example 1
	// Subscription: psubscribe p* 2
	// Message: example hello
	// Message: example world
	// PMessage: p* pexample foo
	// PMessage: p* pexample bar
	// Subscription: unsubscribe example 1
	// Subscription: punsubscribe p* 0
}
