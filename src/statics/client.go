package statics

import (
	"mylog"
	"runtime/debug"
	"time"
	"sync"
	"rpcd"
	"tools"
)

type UserStatics struct {
	Register 		bool
	LoginCount 		int
	PriceCount 		int
	PriceAction		int
	RoomCard 		int
	Coin 			int
	device 			string
}

type UserStaticsSyncArg struct {
	Data 			map[uint32]UserStatics
}

type UserStaticsSyncReply struct {
	ErrCode 		string
}

type StaticsClient struct {
	req 		chan func()

	slock 		sync.Mutex
	storage 	map[uint32]*UserStatics

	master 	*rpcd.RpcdClient
}

func NewStaticsClient() *StaticsClient {
	return &StaticsClient{
		req: make(chan func(), 2048),
		storage: make(map[uint32]*UserStatics),
	}
}

func (sc *StaticsClient) Start() {
	safeCall := func(f func()) {
		defer func() {
			if r := recover(); r != nil {
				mylog.Debug("============statics client call exception==========")
				debug.PrintStack()
			}
		}()
		f()
	}

	sc.master = rpcd.StartClient(tools.GetMasterServiceHost())

	go func() {
		t := time.After(time.Duration(1) * time.Second)
		for {
			select {
			case f := <- sc.req:
				safeCall(f)
			case <- t:
				sc.sync()
				t = time.After(time.Duration(1) * time.Second)
			}
		}
	}()
}

func (sc *StaticsClient) sync() {

	arg := &UserStaticsSyncArg{
		Data: make(map[uint32]UserStatics),
	}

	sc.slock.Lock()
	for uid, info := range sc.storage {
		arg.Data[uid] = UserStatics{
			Register: info.Register,
			LoginCount: info.LoginCount,
			PriceAction: info.PriceAction,
			PriceCount: info.PriceCount,
			RoomCard: info.RoomCard,
			Coin: info.Coin,
		}
	}
	sc.slock.Unlock()

	rep := &UserStaticsSyncReply{}
	if err := sc.master.Call("StaticsServer.SyncData", arg, rep); err != nil {
		mylog.Info("sync data error")
	} else {
		sc.storage = make(map[uint32]*UserStatics)
	}
}

func (sc *StaticsClient) UserLogin(uid uint32, device string) {
	sc.req <- func() {
		sc.slock.Lock()
		if user, ok := sc.storage[uid]; ok {
			user.LoginCount++
		} else {
			sc.storage[uid] = &UserStatics{
				LoginCount: 1,
				device: device,
			}
		}
		sc.slock.Unlock()
	}
}

func (sc *StaticsClient) UserRegister(uid uint32, device string) {
	sc.req <- func() {
		sc.slock.Lock()
		if user, ok := sc.storage[uid]; ok {
			user.Register = true
		} else {
			sc.storage[uid] = &UserStatics{
				Register: true,
				device: device,
			}
		}
		sc.slock.Unlock()
	}
}

func (sc *StaticsClient) UserPay(uid uint32, item, count, price int) {
	sc.req <- func() {
		sc.slock.Lock()
		if user, ok := sc.storage[uid]; ok {
			user.PriceCount += price
			user.PriceAction++
		} else {
			sc.storage[uid] = &UserStatics{
				PriceCount: price,
				PriceAction: 1,
			}
		}
		sc.slock.Unlock()
	}
}

func (sc *StaticsClient) ConsumeRoomCard(uid uint32, count int) {
	sc.req <- func() {
		sc.slock.Lock()
		if user, ok := sc.storage[uid]; ok {
			user.RoomCard += count
		} else {
			sc.storage[uid] = &UserStatics{
				RoomCard: count,
			}
		}
		sc.slock.Unlock()
	}
}

func (sc *StaticsClient) ConsumeCoin(uid uint32, count int) {
	sc.req <- func() {
		sc.slock.Lock()
		if user, ok := sc.storage[uid]; ok {
			user.Coin += count
		} else {
			sc.storage[uid] = &UserStatics{
				Coin: count,
			}
		}
		sc.slock.Unlock()
	}
}