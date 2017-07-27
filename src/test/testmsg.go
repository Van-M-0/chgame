package main

import (
	"network"
	"exportor/defines"
	"exportor/proto"
	"fmt"
	"msgpacker"
)

func testmsgpack() {
	fmt.Println("...............")
	server := network.NewTcpServer(&defines.NetServerOption{
		Host: ":9890",
		ConnectCb: func(client defines.ITcpClient) error {
			fmt.Println("client conect ", client.GetRemoteAddress())
			return nil
		},
		CloseCb: func(client defines.ITcpClient) {

		},
		MsgCb: func(client defines.ITcpClient, m *proto.Message) {

			type SubPacket struct {
				A 	uint32
				B 	uint32
			}

			type Packet struct {
				Cmd 	uint32
				Data 	[]byte
			}


			var p Packet
			msgpacker.UnMarshal(m.Msg, &p)

			fmt.Println(p)

			var s SubPacket
			msgpacker.UnMarshal(p.Data, &s)

			fmt.Println(s)

			fmt.Println(m, p, s)
		},
		AuthCb: func(client defines.ITcpClient) error {
			return nil
		},
	})
	err := server.Start()
	fmt.Println(err)
}