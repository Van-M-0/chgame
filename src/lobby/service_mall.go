package lobby

import (
	"exportor/defines"
	"exportor/proto"
	"sync"
	"time"
	"strconv"
	"crypto/md5"
	"fmt"
	"mylog"
)

type orderKey struct {
	order 		string
}

type orderVal struct {
	user 		uint32
	uid 		uint32
	acc 		string
	item 		int
	tm	 		time.Time
}

type MallService struct {
	lb 			*lobby
	itemsLock 	sync.RWMutex

	ItemConfigList []proto.ItemConfig
	ItemAreaList   []proto.ItemArea

	orderLock 	sync.RWMutex
	orders 		map[orderKey]orderVal
}

func newMallService(lb *lobby) *MallService {
	ms := &MallService{}
	ms.lb = lb
	ms.orders = make(map[orderKey]orderVal)
	return ms
}

func (ms *MallService) start() {
	var r proto.MsLoadItemConfigReply
	ms.lb.dbClient.Call("DBService.LoadItemConfig", &proto.MsLoadItemConfigReply{}, &r)
	ms.itemsLock.Lock()
	ms.ItemConfigList = r.ItemConfigList
	ms.itemsLock.Unlock()

	//mylog.Debug("local item config", r)
}

func (ms *MallService) onUserLoadMalls(uid uint32, req *proto.ClientLoadMallList) {
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

func (ms *MallService) getItem(itemid int) proto.ItemConfig {
	ms.itemsLock.Lock()
	defer ms.itemsLock.Unlock()
	for _, i := range ms.ItemConfigList {
		if int(i.Itemid) == itemid{
			return i
		}
	}
	return proto.ItemConfig{}
}

func (ms *MallService) OnUserBy(uid uint32, req *proto.ClientBuyReq) {
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

	item := ms.getItem(req.Item)
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

func (ms *MallService) getOrder(tm time.Time, user uint32, uid uint32,item int) string {
	seed := strconv.FormatInt(tm.Unix(), 10) + strconv.FormatUint(uint64(user * uid * uint32(item)), 10)
	has := md5.Sum([]byte(seed))
	return fmt.Sprintf("%x", has)
}

func (ms *MallService) OnUserPreyInfo(uid uint32, req *proto.UserPrepay) {
	user := ms.lb.userMgr.getUser(uid)

	if user == nil {
		ms.lb.send2player(uid, proto.CmdUserPreayInfo, &proto.UserPrepayRet{
			ErrCode: defines.ErrComononUserNotIn,
		})
		return
	}

	if req.Item	< 0 {
		ms.lb.send2player(uid, proto.CmdUserPreayInfo, &proto.UserPrepayRet{
			ErrCode: defines.ErrClientBuyInvalid,
		})
		return
	}

	item := ms.getItem(req.Item)
	if item.Itemid == 0 {
		ms.lb.send2player(uid, proto.CmdUserPreayInfo, &proto.UserPrepayRet{
			ErrCode: defines.ErrClientBuyItemNotExists,
		})
		return
	}

	if item.Category != defines.MallItemCategoryDiamond {
		ms.lb.send2player(uid, proto.CmdUserPreayInfo, &proto.UserPrepayRet{
			ErrCode: defines.ErrClientBuyInvalid,
		})
		return
	}

	if item.Buyvalue != req.Price {
		ms.lb.send2player(uid, proto.CmdUserPreayInfo, &proto.UserPrepayRet{
			ErrCode: defines.ErrClientBuyItemMoneyNotEnough,
		})
		return
	}

	tm := time.Now()
	order := ms.getOrder(tm, user.userId, user.uid, req.Item)

	ms.orderLock.Lock()
	ms.orders[orderKey{
		order: order,
	}]	= orderVal{
		user: user.userId,
		uid: user.uid,
		acc: user.account,
		item: req.Item,
		tm: tm,
	}
	ms.orderLock.Unlock()

	mylog.Info("create order ", tm, order, ms.orders[orderKey{order:order}])

	ms.lb.send2player(uid, proto.CmdUserPreayInfo, &proto.UserPrepayRet{
		ErrCode: defines.ErrCommonSuccess,
		OrderId: order,
	})
}

func (ms *MallService) UserPayReturn(req *proto.MsPayNotifyArg, res *proto.MsPayNotifyReply) error {
	key := orderKey{order: req.Order}
	ms.orderLock.Lock()
	order, ok := ms.orders[key]
	ms.orderLock.Unlock()
	if ok {
		od := ms.getOrder(order.tm, order.user, order.uid, order.item)
		if od != req.Order {
			mylog.Info("order not same ?", req.Order, od, order)
		} else {

			ms.orderLock.Lock()
			delete(ms.orders, key)
			ms.orderLock.Unlock()

			item := ms.getItem(order.item)
			if item.Itemid == 0 {
				mylog.Debug("item no more exists ", order.item)
				return nil
			}

			user := ms.lb.userMgr.getUserByAcc(order.acc)
			if user != nil {
				ms.lb.userMgr.updateUserProp(user, defines.PpDiamond, item.Nums)
			} else {
				mylog.Debug("user not online, update to db", order)
				var res proto.MsUpdateUserPropReply
				ms.lb.dbClient.Call("DBService.UpdateUserProp", &proto.MsUpdateUserPropArg{
					UserId: order.user,
					Key: "diamond",
					Diamond: item.Nums,
				}, &res)
			}
		}
	} else {
		mylog.Info("order not exists, ", order)
	}

	return nil
}

func (ms *MallService) GetItemConfig(itemid []int) []proto.ItemConfig {
	ms.itemsLock.Lock()
	defer func() {
		ms.itemsLock.Unlock()
	}()

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
