package proto

// game proto 5000 - 6000
const (
	CmdGamePlayerLogin 			= 5000
	CmdGamePlayerCreateRoom		= 5001
	CmdGamePlayerEnterRoom		= 5002
	CmdGamePlayerLeaveRoom		= 5003

	CmdGamePlayerMessage 		= 5020
)

type PlayerLogin struct {
	Uid 		uint32
}

type PlayerLoginRet struct {
	ErrCode 	int
}

type PlayerCreateRoom struct {
	//common
	Type 		int
	Enter 		bool
	//special
	conf 		[]byte
}

type PlayerCreateRoomRet struct {
	ErrCode 	int
}

type PlayerEnterRoom struct {
	RoomId		uint32
}

type PlayerEnterRoomRet struct {
	ErrCode 	int
}

type PlayerLeaveRoom struct {
	RoomId 		uint32
}

type PlayerLeaveRoomRet struct {
	ErrCode 	uint32
}

type PlayerGameMessage struct {
	Cmd 		uint32
	Msg 		[]byte
}