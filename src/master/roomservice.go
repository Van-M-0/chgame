package master

import (
	"exportor/proto"
	"sync"
	"time"
	"math/rand"
)

type room struct {
	ServerId 	int
	Conf 		[]byte
}

type RoomService struct {
	rmLock 		sync.RWMutex
	rooms 		map[uint32]*room
}

func newRoomService() *RoomService {
	rs := &RoomService{}
	rs.rooms = make(map[uint32]*room)
	return rs
}

var lastRoomId uint32
func (rs *RoomService) CreateRoomId(req *proto.MsCreateoomIdArg, res *proto.MsCreateRoomIdReply) error {
	rs.rmLock.Lock()
	res.RoomId = 0
	rand.Seed(time.Now().UnixNano() + int64(lastRoomId))
	for i := 0; i < 50; i++ {
		id := uint32(rand.Intn(899999) + 100000)
		if _, ok := rs.rooms[id]; !ok {
			res.RoomId = id
			break
		}
	}
	lastRoomId = res.RoomId
	rs.rooms[res.RoomId] = &room{ServerId: req.ServerId, Conf: req.Conf}
	rs.rmLock.Unlock()
	return nil
}

func (rs *RoomService) ReleaseRoom(req *proto.MsReleaseRoomArg, res *proto.MsReleaseReply) error {
	rs.rmLock.Lock()
	res.ErrCode = "error"
	if room, ok := rs.rooms[req.RoomId]; ok {
		if room.ServerId == req.ServerId {
			res.ErrCode = "ok"
		}
	}
	rs.rmLock.Unlock()
	return nil
}

func (rs *RoomService) GetRoomServerId(req *proto.MsGetRoomServerIdArg, res *proto.MsGetRoomServerIdReply) error {
	rs.rmLock.Lock()
	res.ServerId = -1
	if ser, ok := rs.rooms[req.RoomId]; ok {
		res.ServerId = ser.ServerId
		res.Conf = ser.Conf
	}
	rs.rmLock.Unlock()
	return nil
}
