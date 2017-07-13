package proto

import "gameproto/clipb"

const (
	BaseCmdHeader 				= 1

	BaseCmdRegister2gate 		= 2
)

type Message struct {
	Cmd 		int
	Type 		string
	Msg 		interface{}
}


type RegisterServer struct {
	Type 		string
}

func registerComm() {
	Register(BaseCmdHeader, (*gameproto.CliMsgHeader)(nil))
	Register(BaseCmdRegister2gate, (*RegisterServer)(nil))
}

func init () {
	registerComm()
}