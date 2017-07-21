package game

import (
	"exportor/proto"
	"time"
	"math/rand"
	"exportor/defines"
)

type roomManager struct {
	sm 			*sceneManager
	rooms 		map[uint32]*room
}

func newRoomManager(manager *sceneManager) *roomManager {
	return &roomManager{
		sm: manager,
		rooms: make(map[uint32]*room),
	}
}

func (rm *roomManager) getRoomId() uint32 {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 50; i++ {
		id := uint32(rand.Intn(899999) + 100000)
		if _, ok := rm.rooms[id]; !ok {
			return id
		}
	}
	return 0
}

func (rm *roomManager) getRoom(id uint32) *room {
	if r, ok := rm.rooms[id]; ok {
		return r
	} else {
		return nil
	}
}

func (rm *roomManager) createRoom(info *playerInfo, message *proto.PlayerCreateRoom) {
	room := newRoom(rm)
	room.id = rm.getRoomId()
	if room.id != 0 {
		rm.rooms[room.id] = room
		room.createUserId = info.userId
		room.run()
		room.notify <- &roomNotify{
			cmd: proto.CmdGamePlayerCreateRoom,
			user: *info,
			data: message,
		}
		rm.sendMessage(info, proto.CmdGamePlayerCreateRoom, &proto.PlayerCreateRoomRet{ErrCode: defines.ErrCreateRoomSuccess})
	} else {
		rm.sendMessage(info, proto.CmdGamePlayerCreateRoom, &proto.PlayerCreateRoomRet{ErrCode: defines.ErrCreateRoomWait})
	}
}

func (rm *roomManager) enterRoom(info *playerInfo, roomId uint32) {
	room := rm.getRoom(roomId)
	if room == nil {
		rm.sendMessage(info, proto.CmdGamePlayerCreateRoom, &proto.PlayerEnterRoomRet{ErrCode: defines.ErrEnterRoomNotExists})
	}

	room.notify <- &roomNotify{
		cmd: proto.CmdGamePlayerEnterRoom,
		user: *info,
	}
}

func (rm *roomManager) leaveRoom(info *playerInfo, roomId uint32) {
	room := rm.getRoom(roomId)
	if room == nil {
		rm.sendMessage(info, proto.CmdGamePlayerLeaveRoom, &proto.PlayerLeaveRoomRet{ErrCode: defines.ErrLeaveRoomNotExists})
	}

	room.notify <- &roomNotify{
		cmd: proto.CmdGamePlayerEnterRoom,
		user: *info,
	}
}

func (rm *roomManager) offline(info *playerInfo) {

}

func (rm *roomManager) reEnter(info *playerInfo) {

}

func (rm *roomManager) gameMessage(info *playerInfo, cmd uint32, msg []byte) {
	room := rm.getRoom(info.roomid)
	if room == nil {
		return
	}

	room.notify <- &roomNotify{
		cmd: cmd,
		user: *info,
		data: msg,
	}
}

func (rm *roomManager) sendMessage(info *playerInfo, cmd uint32, data interface{}) {
	rm.sm.SendMessage(info.uid, cmd, data)
}

func (rm *roomManager) broadcastMessage(players []*playerInfo, cmd uint32, data interface{}) {

}

