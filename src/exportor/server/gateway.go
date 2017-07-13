package server

type IServer interface {
	Start() error
	Stop() error
}

type IGateway interface {
	IServer
}
