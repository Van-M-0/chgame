package proto

// base proto 1 - 500
// gate way command 501 - 999
// lobby proto 1000 - 2000
// game proto 5000 - 6000

const (
	CmdRange_Base_S 		= 1
	CmdRange_Base_E 		= 500
	CmdRange_Gate_S 		= 501
	CmdRange_Gate_E 		= 1000
	CmdRange_Lobby_S 		= 1001
	CmdRange_Lobby_E 		= 2000
	CmdRange_Game_S 		= 5001
	CmdRange_Game_E 		= 6000
)

const (
	ClientRouteLobby			= 1
	GateRouteLobby 				= 2

	ClientRouteGame	 			= 3
	GateRouteGame			 	= 4

	LobbyRouteGate				= 5
	GameRouteGate				= 6

	LobbyRouteClient			= 7
	GameRouteClient 			= 8
)

const (
	CmdRegisterServer 			= 100
)

type Message struct {
	Len 	uint32
	Cmd 	uint32
	Msg 	[]byte
}

type RegisterServer struct {
	Type 		string
}