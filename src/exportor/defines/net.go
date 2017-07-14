package defines

import (
	"exportor/proto"
)

type ClientConnectCb func(ITcpClient) error
type ClientCloseCb func(ITcpClient)
type ClientMessageCb func(ITcpClient, *proto.Message)
type AuthCb func(ITcpClient) error

type INetContext interface {
	Set(key string, val interface{})
	Get(key string) interface{}
}

type ITcpClient interface {
	INetContext
	Id(uint32)
	GetId() uint32
	GetRemoteAddress() string
	Connect() error
	Close()	error
	OnMessage(*proto.Message)
	Send(*proto.Message) error
	ActiveRead([]byte, int) error
	GetCodec() ICodec
}

type ITcpServer interface {
	INetContext
	Start()	error
	Stop() error
}

type INet interface {
	NewClient(opt *NetClientOption) ITcpClient
	NewServer(opt *NetServerOption) ITcpServer
}

type ICodec interface {
	Encode(m *proto.Message) error
	Decode()(*proto.Message, error)
	DecodeRaw(size int) ([]byte, error)
}