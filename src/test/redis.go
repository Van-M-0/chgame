package main

import (
	"github.com/garyburd/redigo/redis"
	"log"
	"reflect"
	"sync"
	"fmt"
)

type Conf struct {
	Host 		string
}

type Stock struct {
	CompanyName string `redis:"company_name"`
	OpenPrice   string `redis:"open_price"`
	AskPrice    string `redis:"ask_price"`
	ClosePrice  string `redis:"close_price"`
	BidPrice    string `redis:"bid_price"`
	Id 			int    `redis:"id"`
}

func testredis() {
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
		publish("example", "hello")
		publish("example", "world")
		publish("pexample", "foo")
		publish("pexample", "bar")

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
