package proto


// gate way command 501 - 999

const (
	GateMsgTypePlayer			= 1
	GateMsgTypeServer 			= 2
)

type GateGameHeader struct {
	//gw header
	Uid 		uint32
	Type 		int
	//client message
	Cmd 		uint32
	Msg 		[]byte
}

type GateLobbyHeader struct {
	Uid			uint32
	Type 		int
	Cmd 		uint32
	Msg 		[]byte
}

type LobbyGateHeader struct {
	Uids 		[]uint32
	Cmd 		uint32
	Msg 		[]byte
}

type GameGateHeader struct {
	Uids 		[]uint32
	Cmd 		uint32
	Msg 		[]byte
}
