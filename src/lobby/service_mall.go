package lobby

import (
	"exportor/defines"
	"exportor/proto"
	"fmt"
	"sync"
)

type mallService struct {
	lb 			*lobby
	itemsLock 	sync.RWMutex
	items 		map[int]*proto.MallItem
	itemList 	[]*proto.MallItem

	ItemConfigList []proto.ItemConfig
	ItemAreaList   []proto.ItemArea
}

func newMallService(lb *lobby) *mallService {
	ms := &mallService{}
	ms.items = make(map[int]*proto.MallItem)
	ms.itemList = make([]*proto.MallItem, 0)
	ms.lb = lb
	return ms
}

func (ms *mallService) start() {
	var res proto.MsLoadMallItemListReply
	ms.lb.dbClient.Call("DBService.LoadMallItem", &proto.MsLoadMallItemListArg{}, &res)

	var r proto.MsLoadItemConfigReply
	ms.lb.dbClient.Call("DBService.LoadItemConfig", &proto.MsLoadItemConfigReply{}, &r)

	ms.itemsLock.Lock()
	ms.itemList = res.Malls
	ms.items = make(map[int]*proto.MallItem)
	for _, n := range ms.itemList {
		ms.items[n.Id] = n
	}
	ms.ItemConfigList = r.ItemConfigList
	ms.ItemAreaList = r.ItemAreaList
	ms.itemsLock.Unlock()
	fmt.Println("ns notices map", ms.items)
}

func (ms *mallService) onUserLoadMalls(uid uint32, req *proto.ClientLoadMallList) {
	l := make([]proto.MallItem, len(ms.itemList))
	ms.itemsLock.Lock()
	for i := 0; i < len(ms.itemList); i++ {
		l[i] = *ms.itemList[i]
	}
	ms.itemsLock.Unlock()

	ms.lb.send2player(uid, proto.CmdClientLoadMallList, &proto.ClientLoadMallListRet{
		Items: l,
	})
}

func (ms *mallService) OnUserBy(uid uint32, req *proto.ClientBuyReq) {
	var item proto.MallItem
	ms.itemsLock.Lock()
	pItem, ok := ms.items[req.ItemId]
	if ok {
		item = *pItem
	}
	ms.itemsLock.Unlock()

	if item.Id == 0 {
		ms.lb.send2player(uid, proto.CmdClientBuyItem, &proto.ClientBuyMallItemRet{
			ErrCode: defines.ErrClientBuyItemNotExists,
		})
	}

	var v interface{}
	if item.Category == defines.MallItemCategoryDiamond {
		v = ms.lb.userMgr.getUserProp(uid, defines.PpDiamond).(int)
	} else if item.Category == defines.MallItemCategoryGold {
		v = ms.lb.userMgr.getUserProp(uid, defines.PpGold).(int64)
	} else if item.Category == defines.MallItemCategoryRoomCard {
		v = ms.lb.userMgr.getUserProp(uid, defines.PpRoomCard).(int)
	}
	fmt.Println(v)
}