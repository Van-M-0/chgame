package proto

// game proto 5000 - 6000
const (
	GameCmdPlayerLogin		= 5000
)

type PlayerLogin struct {
	Uid 		uint32
	Name 		string
	headImg 	string
	Ip 			string
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



