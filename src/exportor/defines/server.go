package defines

import "exportor/proto"

type IServer interface {
	Start() error
	Stop() error
}

type IGateway interface {
	IServer
}

type ILobby interface {
	IServer
}

type IGame interface {
	IServer
}

type CommunicatorCb func([]byte)
type ICommunicatorClient interface {
	Notify(chanel string, v ...interface{})	error
	JoinChanel(chanel string, reg bool, cb CommunicatorCb) error
}

type ICacheNotify interface {
	OnProps(category string, data []byte)
}

type ICacheClient interface {
	IServer
	GetUserInfo(name string, user *proto.CacheUser) error
	GetUserInfoById(uid uint32, user *proto.CacheUser) error
	SetUserInfo(d interface{}, dbRet bool) error
}

type ICacheLoader interface {
	LoadUsers(name string)
}
