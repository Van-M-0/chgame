package game

import (
	"fmt"
	"exportor/proto"
	"exportor/defines"
	"msgpacker"
	"runtime/debug"
	"sync/atomic"
	"time"
)


type voter struct {
	user 	*defines.PlayerInfo
	agreed 	int	//0, 1 a 2 disa
}

type roomNotify struct {
	cmd 		uint32
	user 		defines.PlayerInfo
	data 		interface{}
}

type room struct {
	module 			defines.GameModule
	game 			defines.IGame
	id 				uint32
	createUserId 	uint32
	manager 		*roomManager
	notify 			chan *roomNotify
	quit 			chan bool
	users 			map[uint32]*defines.PlayerInfo
	closed 			int32

	releaseVoter 	map[uint32]*voter
	timeoutCheck 	time.Timer
	isReleased 		bool
}

func newRoom(manager *roomManager) *room {
	return &room {
		manager: manager,
		notify: make(chan *roomNotify, 1024),
		quit: make(chan bool),
		users: make(map[uint32]*defines.PlayerInfo),
		releaseVoter: make(map[uint32]*voter),
	}
}

func (rm *room) safeCall() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("room call error", r)
			debug.PrintStack()
		}
	}()

	select {
	case n := <- rm.notify:
		fmt.Println("room process message ", n.cmd, n)
		if rm.closed == 1 {
			fmt.Println("room closed")
			return
		}

		if n.cmd == proto.CmdEnterRoom {
			rm.onUserEnter(n)
		} else if n.cmd == proto.CmdGamePlayerLeaveRoom {
			rm.onUserLeave(n)
		} else if n.cmd == proto.CmdGamePlayerMessage {
			rm.onUserMessage(n)
		} else if n.cmd == proto.CmdGamePlayerRoomChat {
			rm.onUserChatMessage(n)
		} else if n.cmd == proto.CmdGamePlayerReleaseRoom {
			rm.onUserReleaseRoom(n)
		} else if n.cmd == proto.CmdGamePlayerReleaseRoomResponse {
			rm.onUserReleaseRoomResponse(n)
		} else if n.cmd == InnerCmdUserReEnter {
			rm.onUserReenter(n)
		} else if n.cmd == InnerCmdUserOffline {
			rm.onUserOffline(n)
		}
	case <- rm.quit:
		fmt.Println("room destroy", rm.id)
	case <- rm.timeoutCheck.C:
		rm.releaseTimeoutCheck()
	}
}

func (rm *room) run() {
	fmt.Println("room run")
	go func () {
		for {
			rm.safeCall()
		}
	}()
}

func (rm *room) destroy() {

}

func (rm *room) onCreate(notify *roomNotify) bool {

	replyErr := func(err int) {
		rm.SendUserMessage(&notify.user, proto.CmdGameCreateRoom, &proto.PlayerCreateRoomRet{ErrCode: err})
	}

	rm.game = rm.module.Creator()
	if rm.game == nil {
		replyErr(defines.ErrCreateRoomGameMoudele)
		return false
	}

	if rm.GetUserInfoFromCache(&notify.user) != nil {
		replyErr(defines.ErrCreateRoomSystme)
		return false
	}

	if notify.user.RoomId != 0 {
		replyErr(defines.ErrCreateRoomHaveRoom)
		return false
	}

	if err := rm.game.OnInit(rm, rm.module.GameData); err != nil {
		replyErr(defines.ErrCreateRoomGameMoudele)
		return false
	}

	if msg, ok := notify.data.(*proto.PlayerCreateRoom); !ok {
		replyErr(defines.ErrCreateRoomSystme)
		return false
	} else {
		if err := rm.game.OnGameCreate(&notify.user, &defines.CreateRoomConf{
			RoomId: rm.id,
			Conf: msg.Conf,
		}); err != nil {
			replyErr(defines.ErrCreateRoomSystme)
			return false
		}
	}

	rm.run()
	rm.SendUserMessage(&notify.user, proto.CmdGameCreateRoom, &proto.PlayerCreateRoomRet{
		ErrCode: defines.ErrCommonSuccess,
		RoomId: rm.id,
		ServerId: rm.manager.sm.gameServer.serverId,
	})

	return true
}

func (rm *room) onUserEnter(notify *roomNotify) {

	if rm.GetUserInfoFromCache(&notify.user) != nil {
		rm.SendUserMessage(&notify.user, proto.CmdGameEnterRoom, &proto.PlayerEnterRoomRet{ErrCode: defines.ErrCommonCache})
		return
	}

	notify.user.RoomId = rm.id
	if err := rm.game.OnUserEnter(&notify.user); err != nil {
		rm.SendUserMessage(&notify.user, proto.CmdGameEnterRoom, &proto.PlayerEnterRoomRet{ErrCode: defines.ErrEnterRoomMoudle})
		notify.user.RoomId = 0
	} else {
		rm.users[notify.user.UserId] = &notify.user
		rm.updateProp(notify.user.UserId, defines.PpRoomId, rm.id)
		rm.SendUserMessage(&notify.user, proto.CmdGameEnterRoom, &proto.PlayerEnterRoomRet{
			ErrCode: defines.ErrCommonSuccess,
			ServerId: rm.manager.sm.gameServer.serverId,
		})
	}
	fmt.Println("onuser enter ", rm.users)
}

func (rm *room) onUserLeave(notify *roomNotify) {
	fmt.Println("user leave room")
	rm.game.OnUserLeave(&notify.user)
	rm.updateProp(notify.user.UserId, defines.PpRoomId, uint32(0))
	delete(rm.users, notify.user.UserId)

	rm.SendUserMessage(&notify.user, proto.CmdGamePlayerReturn2lobby, &proto.PlayerReturn2Lobby{ErrCode: defines.ErrCommonSuccess})
	//rm.SendUserMessage(&notify.user, proto.CmdGamePlayerLeaveRoom, &proto.PlayerLeaveRoomRet{ErrCode: defines.ErrCommonSuccess})

	/*
	if len(rm.users) == 0 {
		rm.ReleaseRoom()
	}

	*/
}

func (rm *room) onUserReenter(notify *roomNotify) {
	rm.game.OnUserReEnter(&notify.user)
	rm.SendUserMessage(&notify.user, proto.CmdGameEnterRoom, &proto.PlayerEnterRoomRet{
		ErrCode: defines.ErrCommonSuccess,
		ServerId: rm.manager.sm.gameServer.serverId,
	})
}

func (rm *room) onUserOffline(notify *roomNotify) {
	rm.game.OnUserOffline(&notify.user)
}

func (rm *room) onUserMessage(notify *roomNotify) {
	var message proto.PlayerGameMessage
	if err := msgpacker.UnMarshal(notify.data.([]byte), &message); err != nil {
		fmt.Println("unmarsh client message error", notify.data)
		return
	}
	fmt.Println("notify ",notify, message.B)
	if err := rm.game.OnUserMessage(&notify.user, message.A, message.B); err != nil {

	} else {

	}
}

func (rm *room) onUserChatMessage(notify *roomNotify) {
	var message proto.PlayerRoomChat
	if err := msgpacker.UnMarshal(notify.data.([]byte), &message); err != nil {
		return
	}

	users := []*defines.PlayerInfo{}
	for _, user := range rm.users {
		users = append(users, user)
	}

	rm.manager.broadcastMessage(users, proto.CmdGamePlayerRoomChat, &proto.PlayerRoomChatRet{
		ErrCode: defines.ErrCommonSuccess,
		Userid: notify.user.UserId,
		Content: message.Content,
	})
}

func (rm *room) onUserReleaseRoom(notify *roomNotify) {

	if rm.isReleased {
		rm.SendUserMessage(&notify.user, proto.CmdGamePlayerReleaseRoom, &proto.PlayerGameReleaseRoomRet{
			ErrCode: defines.ErrCommonInvalidReq,
		})
		return
	}

	users := []*defines.PlayerInfo{}
	for _, user := range rm.users {
		users = append(users, user)

		rm.releaseVoter[user.UserId] = &voter{
			user: user,
			agreed: 0,
		}
		if user.UserId == notify.user.UserId {
			rm.releaseVoter[user.UserId].agreed = 1
		}
	}

	released := rm.checkReleaseRoomCondition()

	rm.manager.broadcastMessage(users, proto.CmdGamePlayerReleaseRoom, &proto.PlayerGameReleaseRoomRet{
		ErrCode: defines.ErrCommonSuccess,
		Sponser: notify.user.UserId,
		Released: released,
	})

	if released {
		rm.ReleaseRoom()
		return
	}

	rm.isReleased = true
	rm.timeoutCheck = *time.NewTimer(time.Duration(60 * time.Second))
}

func (rm *room) onUserReleaseRoomResponse(notify *roomNotify) {

	var message proto.PlayerGameReleaseRoomResponse
	if err := msgpacker.UnMarshal(notify.data.([]byte), &message); err != nil {
		rm.SendUserMessage(&notify.user, proto.CmdGamePlayerReleaseRoomResponse, &proto.PlayerGameReleaseRoomResponseRet{
			ErrCode: defines.ErrCoomonSystem,
		})
		return
	}

	if rm.isReleased == false {
		rm.SendUserMessage(&notify.user, proto.CmdGamePlayerReleaseRoomResponse, &proto.PlayerGameReleaseRoomResponseRet{
			ErrCode: defines.ErrCommonInvalidReq,
		})
		return
	}

	voter, ok := rm.releaseVoter[notify.user.UserId]
	if !ok || voter.agreed != 0 {
		rm.SendUserMessage(&notify.user, proto.CmdGamePlayerReleaseRoomResponse, &proto.PlayerGameReleaseRoomResponseRet{
			ErrCode: defines.ErrCommonInvalidReq,
		})
		return
	} else {
		if message.Agree {
			voter.agreed = 1
		} else {
			voter.agreed = 2
		}
	}

	users := []*defines.PlayerInfo{}
	for _, user := range rm.users {
		users = append(users, user)
	}

	released := rm.checkReleaseRoomCondition()

	rm.manager.broadcastMessage(users, proto.CmdGamePlayerReleaseRoomResponse, &proto.PlayerGameReleaseRoomResponseRet{
		ErrCode: defines.ErrCommonSuccess,
		Released: released,
		Agree: message.Agree,
		Voter: notify.user.UserId,
	})

	if released {
		rm.ReleaseRoom()
	}
}

func (rm *room) releaseTimeoutCheck() {
	for _, u := range rm.releaseVoter {
		if u.agreed == 0 {
			u.agreed = 1
		}
	}
	released := rm.checkReleaseRoomCondition()
	fmt.Println("release time check ", released)
	if released {
		rm.ReleaseRoom()
	} else {
		rm.isReleased = false
	}
}

func (rm *room) checkReleaseRoomCondition() bool {
	agreeCount := 0
	totalCount := 0
	for _, u := range rm.releaseVoter {
		totalCount++
		if u.agreed == 1 {
			agreeCount++
		}
	}
	if agreeCount == totalCount {
		return true
	}
	/*
	if agreeCount == rm.game.GetPlayerCount() {
		return true
	}
	*/
	return false
}

func (rm *room) GetRoomId() uint32 {
	return rm.id
}

func (rm *room) ReleaseRoom() {
	if rm.closed == 1 {
		return
	}
	atomic.AddInt32(&rm.closed, 1)
	rm.manager.sm.sceneNotify <- &request{kind: requestRoom, cmd :"closeroom", data: rm.id}
}

func (rm *room) OnStop() {
	if rm.id != 0 {
		rm.manager.sm.msService.Call("RoomService.ReleaseRoom", &proto.MsReleaseRoomArg{
			ServerId: rm.manager.sm.gameServer.serverId,
			RoomId: rm.id,
		}, &proto.MsReleaseReply{})
	}

	rm.game.OnRelease()

	for _, user := range rm.users {
		rm.updateProp(user.UserId, defines.PpRoomId, uint32(0))
	}

	for _, user := range rm.users {
		fmt.Println("player return to lobby", user)
		rm.SendUserMessage(user, proto.CmdGamePlayerReturn2lobby, &proto.PlayerReturn2Lobby{ErrCode: defines.ErrCommonSuccess})
	}
}

func (rm *room) SendGameMessage(info *defines.PlayerInfo, cmd uint32, data interface{}) {
	fmt.Println("send user message ", info, cmd, data)
	rm.manager.sendMessage(info, proto.CmdGamePlayerMessage, &proto.PlayerGameMessageRet{
		Cmd: cmd,
		Msg: data,
	})
}

func (rm *room) SendUserMessage(info *defines.PlayerInfo, cmd uint32, data interface{}) {
	rm.manager.sendMessage(info, cmd, data)
}

func (rm *room) BroadcastMessage(info []*defines.PlayerInfo, cmd uint32, data interface{}) {
	fmt.Println("bc user message ", cmd, data)
	if info == nil || len(info) == 0 {
		fmt.Println("broad cast message error ", info)
		return
	}
	rm.manager.broadcastMessage(info, proto.CmdGamePlayerMessage, &proto.PlayerGameMessageRet{
		Cmd: cmd,
		Msg: data,
	})
}

func (rm *room) SetTimer(id uint32, data interface{}) error {
	fmt.Println("SetTimer not implement")
	return nil
}

func (rm *room) KillTimer(id uint32) error {
	fmt.Println("KillTimer not implement")
	return nil
}

func (rm *room) updateProp(userId uint32, prop int, value interface{}) {
	if user, ok := rm.users[userId]; ok {
		if rm.manager.sm.cc.UpdateUserInfo(userId, prop, value) {
			update := true
			if prop == defines.PpRoomId {
				user.RoomId = value.(uint32)
			} else if prop == defines.PpGold {
				user.Gold = value.(int64)
			} else if prop == defines.PpDiamond {
				user.Diamond = value.(int)
			} else if prop == defines.PpRoomCard {
				user.RoomCard = value.(int)
			} else if prop == defines.PpScore {
				user.Score = value.(int)
			} else {
				update = false
				fmt.Println("update user prop not exists ", userId, prop)
			}
			if update {
				rm.manager.sendMessage(user, proto.CmdBaseUpsePropUpdate, &proto.SyncUserProps {
					Props: &proto.UserProp{
						Ppkey: prop,
						PpVal: value,
					},
				})
			}
		}
	} else {
		fmt.Println("update prop user not in")
	}
}

func (rm *room) updateUserItem(user *defines.PlayerInfo, itemId uint32, count int) bool {
	updateFlag := 0
	defer func() {
		p := proto.ItemProp{
			Flag: updateFlag,
			ItemId: itemId,
			Count: count,
		}
		if updateFlag != 0 {
			rm.manager.sendMessage(user, proto.CmdBaseUpsePropUpdate, &proto.SyncUserProps {
				Items: &p,
			})
		}
	}()
	for index, item := range user.Items {
		if item.ItemId == itemId {
			item.Count += count
			if item.Count <= 0 {
				user.Items = append(user.Items[:index], user.Items[index+1:]...)
				updateFlag = 2
			} else {
				updateFlag = 1
			}
			rm.manager.sm.cc.UpdateSingleItem(user.UserId, updateFlag, item.ItemId, item.Count)
			return true
		}
	}
	if count > 0 {
		updateFlag = 3
		user.Items = append(user.Items, &proto.UserItem{
			ItemId: itemId,
			Count: count,
		})
		rm.manager.sm.cc.UpdateSingleItem(user.UserId, updateFlag, itemId, count)
		return true
	}
	fmt.Println("update item err, id not exists ", itemId, count)
	return false
}

func (rm *room) GetUserInfoFromCache(user *defines.PlayerInfo) error {
	var cu proto.CacheUser
	err := rm.manager.sm.cc.GetUserInfoById(user.UserId, &cu)
	if err != nil {
		return err
	} else if int(user.UserId) != cu.Uid {
		return fmt.Errorf("cache user not same %v", user.Uid)
	} else {
		user.RoomId = uint32(cu.RoomId)
		user.Gold = cu.Gold
		user.RoomCard = cu.RoomCard
		user.Diamond = cu.Diamond
	}
	return nil
}

func (rm *room) SaveGameRecord(head, data []byte) int {
	return rm.manager.sm.cc.SaveGameRecord(head, data)
}

func (rm *room) SaveUserRecord(userid, id int) error {
	return rm.manager.sm.cc.SaveUserRecord(userid, id)
}

func (rm *room) UpdateUserInfo(info *defines.PlayerInfo, data *proto.GameUserPpUpdate) bool {
	if info == nil || data == nil {
		return false
	}
	if data.Gold != nil {
		rm.updateProp(info.UserId, defines.PpGold, *data.Gold)
	}
	if data.Score != nil {
		rm.updateProp(info.UserId, defines.PpScore, *data.Score)
	}
	if data.Diamond != nil {
		rm.updateProp(info.UserId, defines.PpDiamond, *data.Diamond)
	}
	if data.Item != nil && data.Item.ItemId != 0 {
		rm.updateUserItem(info, data.Item.ItemId, data.Item.Count)
	}
	return true
}
