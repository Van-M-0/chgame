package proto


// gate way command 501 - 999

const (
	GateMsgTypePlayer			= 1
	GateMsgTypeServer 			= 2
)

type GateGameHeader struct {
	Uid 		uint32
	Type 		int
	Msg 		*Message
}
