package lobby

import (
	"exportor/defines"
	"exportor/proto"
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

	//mylog.Debug("local item config", r)
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
	user := ms.lb.userMgr.getUser(uid)
	//mylog.Debug("user buy", user)

	if user == nil {
		ms.lb.send2player(uid, proto.CmdClientBuyItem, &proto.ClientBuyMallItemRet{
			ErrCode: defines.ErrComononUserNotIn,
		})
		return
	}

	if req.Item	< 0 {
		ms.lb.send2player(uid, proto.CmdClientBuyItem, &proto.ClientBuyMallItemRet{
			ErrCode: defines.ErrClientBuyInvalid,
		})
		return
	}

	ms.itemsLock.Lock()
	for _, i := range ms.ItemConfigList {
		if int(i.Itemid) == req.Item {
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

	if item.Category == defines.MallItemCategoryGold || item.Category == defines.MallItemCategoryItem {
		if item.Buyvalue > user.diamond {
			ms.lb.send2player(uid, proto.CmdClientBuyItem, &proto.ClientBuyMallItemRet{
				ErrCode: defines.ErrClientBuyItemMoneyNotEnough,
			})
			return
		} else if ms.lb.userMgr.updateUserProp(user, defines.PpDiamond, int(-item.Buyvalue)) == false {
			ms.lb.send2player(uid, proto.CmdClientBuyItem, &proto.ClientBuyMallItemRet{
				ErrCode: defines.ErrClientBuyConsumeErr,
			})
			return
		}
	}

	if item.Category == defines.MallItemCategoryGold {
		ms.lb.userMgr.updateUserProp(user, defines.PpGold, int64(item.Nums))
	} else if item.Category == defines.MallItemCategoryDiamond {
		ms.lb.userMgr.updateUserProp(user, defines.PpDiamond, item.Nums)
	} else if item.Category == defines.MallItemCategoryItem {
		ms.lb.userMgr.updateUserItem(user, item.Itemid, item.Nums)
	}

	//mylog.Debug("client buy item success ", item)

	ms.lb.send2player(uid, proto.CmdClientBuyItem, &proto.ClientBuyMallItemRet{
		ErrCode: defines.ErrCommonSuccess,
	})
}

func (ms *mallService) GetItemConfig(itemid []int) []proto.ItemConfig {
	ms.itemsLock.Lock()
	defer func() {
		ms.itemsLock.Unlock()
	}()
	//mylog.Debug("get item config ", ms.ItemConfigList)
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