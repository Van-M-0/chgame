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

	ItemConfigList []proto.ItemConfig
	ItemAreaList   []proto.ItemArea
}

func newMallService(lb *lobby) *mallService {
	ms := &mallService{}
	ms.lb = lb
	return ms
}

func (ms *mallService) start() {
	var r proto.MsLoadItemConfigReply
	ms.lb.dbClient.Call("DBService.LoadItemConfig", &proto.MsLoadItemConfigReply{}, &r)
	ms.itemsLock.Lock()
	ms.ItemConfigList = r.ItemConfigList
	ms.itemsLock.Unlock()

	fmt.Println("local item config", r)
}

func (ms *mallService) onUserLoadMalls(uid uint32, req *proto.ClientLoadMallList) {
	l := []proto.MallItem{}
	ms.itemsLock.Lock()
	for _, item := range ms.ItemConfigList {
		if item.Sell == 1 {
			m := proto.MallItem{
				Id: int(item.Itemid),
				Name: item.Itemname,
				Category: item.Category,
				BuyValue: item.Buyvalue,
				Nums: item.Nums,
			}
			l = append(l, m)
		}
	}
	ms.itemsLock.Unlock()

	ms.lb.send2player(uid, proto.CmdClientLoadMallList, &proto.ClientLoadMallListRet{
		Items: l,
	})
}

func (ms *mallService) OnUserBy(uid uint32, req *proto.ClientBuyReq) {
	var item proto.ItemConfig
	ms.itemsLock.Lock()
	for _, i := range ms.ItemConfigList {
		if int(i.Itemid) == req.ItemId {
			item = i
			break
		}
	}
	ms.itemsLock.Unlock()

	if item.Itemid == 0 {
		ms.lb.send2player(uid, proto.CmdClientBuyItem, &proto.ClientBuyMallItemRet{
			ErrCode: defines.ErrClientBuyItemNotExists,
		})
		return
	}

	user := ms.lb.userMgr.getUser(uid)
	fmt.Println("user buy", user)

	var v interface{}
	if item.Category == defines.MallItemCategoryDiamond {
		//v = ms.lb.userMgr.getUserProp(uid, defines.PpDiamond).(int)
	} else if item.Category == defines.MallItemCategoryGold {
		//v = ms.lb.userMgr.getUserProp(uid, defines.PpGold).(int64)
	} else if item.Category == defines.MallItemCategoryRoomCard {
		//v = ms.lb.userMgr.getUserProp(uid, defines.PpRoomCard).(int)
	}

	fmt.Println(v)
}

func (ms *mallService) GetItemConfig(itemid []int) []proto.ItemConfig {
	ms.itemsLock.Lock()
	defer func() {
		ms.itemsLock.Unlock()
	}()
	fmt.Println("get item config ", ms.ItemConfigList)
	items := []proto.ItemConfig{}
	for _, id := range itemid {
		for _, item := range ms.ItemConfigList {
			if item.Itemid == uint32(id) {
				items = append(items, item)
			}
		}
	}
	return items
}