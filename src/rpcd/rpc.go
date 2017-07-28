package rpcd

import (
	"net/rpc"
	"net"
	"fmt"
	"net/http"
)

type RpcdClient struct {
	*rpc.Client
}

func (c *RpcdClient) Call(method string, arg interface{}, res interface{}) error {
	return c.Client.Call(method, arg, res)
}

func StartServer(port string) error {
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", port)
	if e != nil {
		fmt.Println("listen err", e)
		panic("")
	}
	http.Serve(l, nil)
	return nil
}

func StartClient(port string) *RpcdClient {
	c, err := rpc.DialHTTP("tcp", port)
	if err != nil {
		fmt.Println("dail rpc server error ", err)
		panic("")
	}
	return &RpcdClient{
		Client:	c,
	}
}
