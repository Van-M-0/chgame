package defines


type IServer interface {
	Start() error
	Stop() error
}

type IGateway interface {
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
	SetCacheNotify(ICacheNotify)
	GetUserInfo(name string)
	SetUserInfo(d interface{}, dbRet bool) error
}

type ICacheLoader interface {
	LoadUsers(name string)
}
