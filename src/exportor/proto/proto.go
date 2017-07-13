package proto

import "gameproto/clipb"

const (
	MagicDirectionGate 			= 1
	MagicDirectionClient		= 2
)

const (
	BaseCmdHeader 				= 1
	BaseCmdRegister2gate 		= 2
)

type Message struct {
	Cmd   int
	Magic int
	Msg   interface{}
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