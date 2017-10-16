package game

import (
	"exportor/proto"
	"time"
	"math/rand"
	"exportor/defines"
	"mylog"
)

const (
	InnerCmdUserOffline = 512
	InnerCmdUserReEnter = 513
	InnerCmdUserLeaveCoinRoom = 514
)

const (
	EnterCoinShareRoomMaxCount = 10
)

const (
	RoomCategoryCard 	= 1
	RoomCategoryCoinBuild = 2
	RoomCategoryCoinShare = 3
)

type coinRoomInfo struct {
	id 			uint32
	count 		int
}

type coinRoomChangeContext struct {
	userid 		uint32
	roomid 		uint32
	kind 		int
}

type roomManager struct {
	sm 				*sceneManager
	rooms 			map[uint32]*room
	coinRoomList 	map[uint32]*coinRoomInfo
}

func newRoomManager(manager *sceneManager) *roomManager {
	return &roomManager{
		sm: manager,
		rooms: make(map[uint32]*room),
		coinRoomList: make(map[uint32]*coinRoomInfo),
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

func (rm *roomManager) getGameModule(kind int) *defines.GameModule {
	modules := rm.sm.gameServer.opt.Moudles
	for _, m := range modules {
		if m.Type == kind {
			return &m
		}
	}
	return nil
}

func (rm *roomManager) getRoom(id uint32) *room {
	if r, ok := rm.rooms[id]; ok {
		return r
	} else {
		return nil
	}
}

func (rm *roomManager) createRoom(info *defines.PlayerInfo, message *proto.PlayerCreateRoom) {

	module := rm.getGameModule(message.Kind)
	if module == nil {
		rm.sm.SendMessageRIndex(info.Uid, proto.CmdGameCreateRoom, &proto.PlayerCreateRoomRet{ErrCode: defines.ErrCreateRoomKind})
		return
	}

	var rep proto.MsCreateRoomIdReply
	rm.sm.msService.Call("RoomService.CreateRoomId", &proto.MsCreateoomIdArg{
		ServerId: rm.sm.gameServer.serverId,
		Conf: message.Conf,
	}, &rep)
	if rep.RoomId == 0 {
		rm.sm.SendMessageRIndex(info.Uid, proto.CmdGameCreateRoom, &proto.PlayerCreateRoomRet{ErrCode: defines.ErrCreateRoomRoomId})
		return
	}

	room := newRoom(rm)
	room.id = rep.RoomId
	mylog.Debug("room id ", room.id)
	rm.rooms[room.id] = room
	room.createUserId = info.UserId
	room.module = *module
	room.category = RoomCategoryCard
	ok := room.onCreate(&roomNotify{
		cmd: proto.CmdGameCreateRoom,
		user: *info,
		data: message,
	})
	if !ok {
		rm.sm.msService.Call("RoomService.ReleaseRoom", &proto.MsReleaseRoomArg{
			ServerId: rm.sm.gameServer.serverId,
			RoomId: room.id,
		}, &proto.MsReleaseReply{})
		delete(rm.rooms, room.id)
		return
	}
}

func (rm *roomManager) enterRoom(info *defines.PlayerInfo, roomId uint32) {
	room := rm.getRoom(roomId)
	if room == nil {
		rm.sm.SendMessageRIndex(info.Uid, proto.CmdGameCreateRoom, &proto.PlayerEnterRoomRet{ErrCode: defines.ErrEnterRoomNotExists})
		return
	}

	info.RoomId = roomId
	room.notify <- &roomNotify{
		cmd: proto.CmdGameEnterRoom,
		user: *info,
	}
}

func (rm *roomManager) destroyRoom(roomid uint32) {
	mylog.Debug("destroy room ", roomid)
	room := rm.getRoom(roomid)
	if room != nil {
		delete(rm.rooms, roomid)
		mylog.Debug("destroy room", room)
	}
	if _, ok := rm.coinRoomList[roomid]; ok {
		delete(rm.coinRoomList, roomid)
	}

	for _, user := range room.users {
		rm.sm.playerMgr.delPlayer(user)
	}

	room.OnStop()
}

func (rm *roomManager) offline(info *defines.PlayerInfo) {
	mylog.Debug("player off line", info)
	room := rm.getRoom(info.RoomId)
	if room == nil {
		mylog.Debug("offline ", info.RoomId)
		return
	}
	room.notify <- &roomNotify{
		cmd: InnerCmdUserOffline,
		user: *info,
	}
}

func (rm *roomManager) reEnter(info *defines.PlayerInfo) {
	mylog.Debug("player reenter", info)
	room := rm.getRoom(info.RoomId)
	if room == nil {
		mylog.Debug("reenter ", info.RoomId)
		return
	}
	room.notify <- &roomNotify{
		cmd: InnerCmdUserReEnter,
		user: *info,
	}
}

func (rm *roomManager) leaveRoom(info *defines.PlayerInfo, ret *proto.PlayerLeaveRoom) {
	room := rm.getRoom(info.RoomId)
	if room == nil {
		mylog.Debug("leave ", info.RoomId)
		return
	}
	room.notify <- &roomNotify{
		cmd: proto.CmdGamePlayerLeaveRoom,
		user: *info,
	}
}

func (rm *roomManager) gameMessage(info *defines.PlayerInfo, cmd uint32, msg []byte) {
	room := rm.getRoom(info.RoomId)
	if room == nil {
		mylog.Debug("room error ", info.RoomId)
		return
	}
	room.notify <- &roomNotify{
		cmd: cmd,
		user: *info,
		data: msg,
	}
}

func (rm *roomManager) chatMessage(info *defines.PlayerInfo, cmd uint32, msg []byte) {
	if info.RoomId == 0 {
		rm.sm.SendMessageRIndex(info.Uid, cmd, &proto.PlayerRoomChatRet{ErrCode: defines.ErrCommonInvalidReq})
		return
	}

	room := rm.getRoom(info.RoomId)
	if room == nil {
		rm.sm.SendMessageRIndex(info.Uid, cmd, &proto.PlayerRoomChatRet{ErrCode: defines.ErrCommonInvalidReq})
		return
	}

	room.notify <- &roomNotify{
		cmd: cmd,
		user: *info,
		data: msg,
	}
}

func (rm *roomManager) playerReleaseRoom(info *defines.PlayerInfo, cmd uint32, msg []byte) {
	if info.RoomId == 0 {
		rm.sm.SendMessageRIndex(info.Uid, cmd, &proto.PlayerGameReleaseRoomRet{ErrCode: defines.ErrCommonInvalidReq})
		return
	}

	room := rm.getRoom(info.RoomId)
	if room == nil {
		rm.sm.SendMessageRIndex(info.Uid, cmd, &proto.PlayerGameReleaseRoomRet{ErrCode: defines.ErrCommonInvalidReq})
		return
	}

	room.notify <- &roomNotify{
		cmd: cmd,
		user: *info,
		data: msg,
	}
}

func (rm *roomManager) playerReleaseRoomResponse(info *defines.PlayerInfo, cmd uint32, msg []byte) {
	if info.RoomId == 0 {
		rm.sm.SendMessageRIndex(info.Uid, cmd, &proto.PlayerGameReleaseRoomResponseRet{ErrCode: defines.ErrCommonInvalidReq})
		return
	}

	room := rm.getRoom(info.RoomId)
	if room == nil {
		rm.sm.SendMessageRIndex(info.Uid, cmd, &proto.PlayerGameReleaseRoomResponseRet{ErrCode: defines.ErrCommonInvalidReq})
		return
	}

	room.notify <- &roomNotify{
		cmd: cmd,
		user: *info,
		data: msg,
	}
}

func (rm *roomManager) sendMessage(info *defines.PlayerInfo, cmd uint32, data interface{}) {
	rm.sm.SendMessage(info, cmd, data)
}

func (rm *roomManager) broadcastMessage(players []*defines.PlayerInfo, cmd uint32, data interface{}) {
	mylog.Debug("broadcast message ", players, cmd, data)
	for _, user := range players {
		if user == nil {
			mylog.Debug("broad cast message find user nil")
			continue
		}
	}
	rm.sm.BroadcastMessage(players, cmd, data)
}


func (rm *roomManager) coinCreateRoom(info *defines.PlayerInfo, kind int, conf []byte) {

	module := rm.getGameModule(kind)
	if module == nil {
		rm.sm.SendMessageRIndex(info.Uid, proto.CmdGameEnterCoinRoom, &proto.PlayerCreateRoomRet{ErrCode: defines.ErrCreateRoomKind})
		return
	}

	if module.GameType != defines.GameTypeCoin {
		rm.sm.SendMessageRIndex(info.Uid, proto.CmdGameEnterCoinRoom, &proto.PlayerCreateRoomRet{ErrCode: defines.ErrCreateRoomType})
		return
	}

	var rep proto.MsCreateRoomIdReply
	rm.sm.msService.Call("RoomService.CreateRoomId", &proto.MsCreateoomIdArg{
		ServerId: rm.sm.gameServer.serverId,
	}, &rep)
	if rep.RoomId == 0 {
		rm.sm.SendMessageRIndex(info.Uid, proto.CmdGameCreateRoom, &proto.PlayerCreateRoomRet{ErrCode: defines.ErrCreateRoomRoomId})
		return
	}

	room := newRoom(rm)
	room.id = rep.RoomId
	mylog.Debug("room id ", room.id)
	rm.rooms[room.id] = room
	room.category = RoomCategoryCoinBuild
	room.createUserId = info.UserId
	room.module = *module

	rm.coinRoomList[room.id] = &coinRoomInfo{
		id:    room.id,
		count: 0,
	}

	ok := room.onCreate(&roomNotify{
		cmd: proto.CmdGameEnterCoinRoom,
		user: *info,
		data: &proto.PlayerCreateRoom{
			Kind: kind,
			Conf: conf,
		},
	})
	if !ok {
		rm.sm.msService.Call("RoomService.ReleaseRoom", &proto.MsReleaseRoomArg{
			ServerId: rm.sm.gameServer.serverId,
			RoomId: room.id,
		}, &proto.MsReleaseReply{})
		delete(rm.rooms, room.id)
	}
}

func (rm *roomManager) coinChangeRoom_1(info *defines.PlayerInfo, oldRoom uint32) {
	if _, ok := rm.coinRoomList[oldRoom]; !ok {
		mylog.Info("not exists change room old room id ", oldRoom, info)
		rm.sendMessage(info, proto.CmdGameEnterCoinRoom, &proto.PlayerCreateRoomRet{ErrCode: defines.ErrEnterCoinOldRoomNotExist})
		return
	}

	if r, ok := rm.rooms[oldRoom]; !ok {
		mylog.Info("not exists change room ", oldRoom, info)
		rm.sendMessage(info, proto.CmdGameEnterCoinRoom, &proto.PlayerCreateRoomRet{ErrCode: defines.ErrEnterCoinOldRoomNotExist})
		return
	} else {
		r.notify <- &roomNotify{
			cmd: InnerCmdUserLeaveCoinRoom,
			user: *info,
		}
	}
}

func (rm *roomManager) coinChangeRoom_2(context *coinRoomChangeContext) {
	info := rm.sm.playerMgr.getPlayerById(context.userid)
	if info == nil {
		mylog.Debug("coinchroom 2, user noti eists")
		return
	}

	if _, ok := rm.coinRoomList[context.roomid]; !ok {
		mylog.Info("not exists change room old room id ", context.roomid, info)
		return
	}
	if cr, ok := rm.coinRoomList[context.roomid]; ok {
		cr.count--
	}

	rm.coinEnterShareRoom(info, context.kind, EnterCoinShareRoomMaxCount)
}

func (rm *roomManager) coinEnterShareRoom(info *defines.PlayerInfo, kind int, tryCount int) {

	if tryCount == 0 {
		rm.sm.SendMessageRIndex(info.Uid, proto.CmdGameEnterCoinRoom, &proto.PlayerCreateRoomRet{ErrCode: defines.ErrCoinCreateRoomMaxCount})
		return
	}

	module := rm.getGameModule(kind)
	if module == nil {
		rm.sm.SendMessageRIndex(info.Uid, proto.CmdGameEnterCoinRoom, &proto.PlayerCreateRoomRet{ErrCode: defines.ErrCreateRoomKind})
		return
	}

	if module.GameType != defines.GameTypeCoin {
		rm.sm.SendMessageRIndex(info.Uid, proto.CmdGameEnterCoinRoom, &proto.PlayerCreateRoomRet{ErrCode: defines.ErrCreateRoomCmd})
		return
	}

	r := rm.selectShareRoom()
	if r == nil {
		var rep proto.MsCreateRoomIdReply
		rm.sm.msService.Call("RoomService.CreateRoomId", &proto.MsCreateoomIdArg{
			ServerId: rm.sm.gameServer.serverId,
		}, &rep)
		if rep.RoomId == 0 {
			rm.sm.SendMessageRIndex(info.Uid, proto.CmdGameCreateRoom, &proto.PlayerCreateRoomRet{ErrCode: defines.ErrCreateRoomRoomId})
			return
		}

		room := newRoom(rm)
		room.id = rep.RoomId
		mylog.Debug("room id ", room.id)
		rm.rooms[room.id] = room
		room.createUserId = info.UserId
		room.module = *module
		room.category = RoomCategoryCoinShare

		ok := room.onCreate(&roomNotify{
				cmd: proto.CmdGameEnterCoinRoom,
				user: *info,
				data: &proto.PlayerCreateRoom{
					Kind: kind,
				},
		})
		if !ok {
			rm.sm.msService.Call("RoomService.ReleaseRoom", &proto.MsReleaseRoomArg{
				ServerId: rm.sm.gameServer.serverId,
				RoomId: room.id,
			}, &proto.MsReleaseReply{})
			delete(rm.rooms, room.id)
		} else {
			r = room
		}
	}

	if r == nil {
		rm.sm.SendMessageRIndex(info.Uid, proto.CmdGameEnterCoinRoom, &proto.PlayerCreateRoomRet{ErrCode: defines.ErrCreateRoomCreate})
		return
	}

	ok := r.onUserEnter(&roomNotify{
		cmd: proto.CmdGameEnterCoinRoom,
		user: *info,
	})

	if ok {
		rm.coinRoomList[r.id].count++
	} else {
		rm.coinEnterShareRoom(info, kind, tryCount-1)
	}
}

func (rm *roomManager) coinRoomEnterWithRoomId(info *defines.PlayerInfo, roomId uint32) {
	if _, ok := rm.coinRoomList[roomId]; !ok {
		rm.sendMessage(info, proto.CmdGameEnterCoinRoom, &proto.PlayerCreateRoomRet{ErrCode: defines.ErrEnterCoinOldRoomNotExist})
		return
	}

	if r, ok := rm.rooms[roomId]; !ok {
		rm.sendMessage(info, proto.CmdGameEnterCoinRoom, &proto.PlayerCreateRoomRet{ErrCode: defines.ErrEnterCoinOldRoomNotExist})
		return
	} else if r.category != RoomCategoryCoinBuild {
		rm.sendMessage(info, proto.CmdGameEnterCoinRoom, &proto.PlayerCreateRoomRet{ErrCode: defines.ErrEnterCoinCategory})
		return
	} else {
		ok := r.onUserEnter(&roomNotify{
			cmd: proto.CmdGameEnterCoinRoom,
			user: *info,
		})

		if ok {
			rm.coinRoomList[r.id].count++
		}
	}
}

func (rm *roomManager) coinReEnterRoom(info *defines.PlayerInfo, room uint32) {
	r := rm.getRoom(room)
	if r == nil {
		rm.sm.SendMessageRIndex(info.Uid, proto.CmdGameEnterCoinRoom, &proto.PlayerEnterRoomRet{ErrCode: defines.ErrEnterRoomNotExists})
		return
	}
	r.onUserReenter(&roomNotify{
		cmd: proto.CmdGameEnterCoinRoom,
		user: *info,
	})
}

func (rm *roomManager) selectShareRoom() *room {
	for id, cr := range rm.coinRoomList {
		if r, ok := rm.rooms[id]; ok {
			if r.category != RoomCategoryCoinShare {
				continue
			}
			if cr.count > 0 && cr.count != 4 {
				return r
			}
		}
	}
	return nil
}
