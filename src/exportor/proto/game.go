package proto

// game proto 5000 - 6000
const (
	CmdGamePlayerLogin = 5000
)

type PlayerLogin struct {
	Uid 		uint32
}

type PlayerLoginRet struct {
	ErrCode 	int
}

type XzRoomConf struct {

}

type PlayerCreateRoomMsg struct {
	Uid 		uint32
}

type PlayerEnterMsg struct {

}

type PlayerLeaveMsg struct {

}

type PlayerOfflineMsg struct {

}

type RoomCreateMsg struct {

}

type RoomDestroyMsg struct {

}



