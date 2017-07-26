package lobby

import (
	"exportor/defines"
	"exportor/proto"
	"cacher"
	"fmt"
	"sync"
)

type mallService struct {
	lb 			*lobby
	itemsLock 	sync.RWMutex
	items 		map[int]*proto.MallItem
	itemList 	[]*proto.MallItem
	cc 			defines.ICacheClient
}

func newMallService(lb *lobby) *mallService {
	ms := &mallService{}
	ms.cc = cacher.NewCacheClient("mall")
	ms.items = make(map[int]*proto.MallItem)
	ms.itemList = make([]*proto.MallItem, 0)
	ms.lb = lb
	return ms
}

func (ms *mallService) start() {
	ms.cc.Start()
	ms.lb.bpro.Register(defines.ChannelTypeMall, defines.ChannelUpdateMall, func(data interface{}) {
		ms.itemUpdate(data)
	})
	ms.lb.bpro.Register(defines.ChannelTypeMall, defines.ChannelLoadMall, func(data interface{}) {
		ms.itemLoaded(data)
	})
}

func (ms *mallService) itemUpdate(data interface{}) {
	items, ok := data.(*proto.PMMallItemUdpate)
	if !ok {
		fmt.Println("mall item update error")
		return
	}

	for _, item := range items.Items {
		ms.items[item.Id] = &item
	}
	ms.makeMallItemList()
}

func (ms *mallService) itemLoaded(data interface{}) {
	items, ok := data.(*proto.PMMallItemLoaded)
	if !ok {
		fmt.Println("mall item update error")
		return
	}

	for _, item := range items.Items {
		ms.items[item.Id] = &item
	}
	ms.makeMallItemList()
}

func (ms *mallService) makeMallItemList() {
	items := []*proto.MallItem{}
	for _, item := range ms.items {
		items = append(items, item)
	}
	ms.itemList = items
}

func (ms *mallService) OnUserLoadItem(info *defines.PlayerInfo, item *proto.ClientLoadMallItem) {
	ms.lb.send2player(info.Uid, proto.CmdClientLoadMallItem, &proto.ClientLoadMallItemRet{
		Items: ms.itemList,
	})
}

func (ms *mallService) OnUserBy(info *defines.PlayerInfo, req *proto.ClientBuyReq) {
	_, ok := ms.items[req.ItemId]
	if !ok {
		ms.lb.send2player(info.Uid, proto.CmdClientBuyItem, &proto.ClientBuyMallItemRet{
			ErrCode: defines.ErrClientBuyItemNotExists,
		})
	}
}