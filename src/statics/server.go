package statics

import (
	"time"
	"mylog"
	"runtime/debug"
	"rpcd"
	"tools"
	"exportor/proto"
	"encoding/json"
	"sort"
	"os"
	"io/ioutil"
	"exportor/defines"
	"fmt"
	"errors"
)

const (
	SaveDayCount = 90
	TimeKeyFormat = "2006-01-02"
)

type Actions struct {
	Register 		bool
	Logined 		bool
	Payed 			bool
}

type DayStatics struct {
	Staticsed 			bool
	StartCount 			int			//启动次数
	StartCountNotRepeat int			//启动去重复次数
	StartCountAdd 		int			//新增启动次数

	LoginCount			int			//登陆次数
	RegisterCount 		int			//注册次数

	DAU 				int			//日活
	WAU					int			//周活
	MAU 				int			//月活

	RoomCardConsume 	int			//房卡消耗
	CoinConsume 		int			//金币消耗

	PayUserCount 		int		//总付费用户
	PayUserCountAdd	 	int		//新增付费用户
	PayCount 			int		//总付费次数
	PayNewUserCount 	int		//新增用户付费金额
	PayPriceCount 		int		//所有用户付费金额

	UserAction 			map[uint32]*Actions
}

type UserConsumes struct {
	RoomCard 			int64
	Coin 				int64
	PayCount 			int64
}

type HistoryStatics struct {
	RegisterCount 		int
	Consumes 			map[uint32]*UserConsumes
}

type StaticsServer struct {
	Days 	map[string]*DayStatics
	History HistoryStatics

	req 	chan func()
	db 		*rpcd.RpcdClient
}

type storageLayout struct {
	Days 	map[string]*DayStatics
	History HistoryStatics
}

func NewStaticsServer() *StaticsServer {
	return &StaticsServer{
		Days: make(map[string]*DayStatics),
		History: HistoryStatics{
			Consumes: make(map[uint32]*UserConsumes),
		},
		req: make(chan func(), 1024),
	}
}

func (ss *StaticsServer) readStorage() {

	file := defines.WorkDir + "statics.data"
	mylog.Debug("statics storage file ", file)


	if !tools.FileExists(file) {
		if _, err := os.Create(file); err != nil {
			panic("create statics.data error " + file + ", " + err.Error())
			return
		}
	}

	content, err := ioutil.ReadFile(file)
	if err != nil {
		panic("read statics.data error " + err.Error())
	}

	sl := storageLayout{}
	if err := json.Unmarshal(content, &sl); err != nil {
		return
	}

	fmt.Println(sl)

	ss.Days = sl.Days
	ss.History = sl.History
}

func (ss *StaticsServer) saveStorage() {
	sl := storageLayout{
		Days: ss.Days,
		History: ss.History,
	}

	stream, err := json.Marshal(sl)
	if err != nil {
		mylog.Info("save storage data error ", err)
		return
	}

	if err := ioutil.WriteFile(defines.WorkDir + "statics.data", stream, os.ModeAppend); err != nil {
		mylog.Info("write storage data error ", err)
	}
}

func (ss *StaticsServer) Start() {
	ss.readStorage()
	safeCall := func(f func()) {
		defer func() {
			if r := recover(); r != nil {
				mylog.Debug("============statics client call exception==========")
				debug.PrintStack()
			}
		}()
		f()
	}

	ss.db = rpcd.StartClient(tools.GetDbServiceHost())

	go func() {
		t := time.After(time.Duration(10) * time.Second)
		for {
			select {
			case f := <- ss.req:
				safeCall(f)
			case <- t:
				for k, statics := range ss.Days {
					key := time.Now().Format(TimeKeyFormat)
					if k < key {
						if statics.Staticsed {
							continue
						}
						statics.Staticsed = true
						ss.History.RegisterCount += statics.RegisterCount
					}
				}
				t = time.After(time.Duration(10) * time.Second)
			}
		}
	}()
}

func (ss *StaticsServer) Stop() {
	ss.saveStorage()
}

func (ss *StaticsServer) SyncData(arg *UserStaticsSyncArg, reply *UserStaticsSyncReply) error {
	fmt.Println("sync data ")
	ss.req <- func() {
		fmt.Println("sync data ", arg, reply)

		key := time.Now().Format(TimeKeyFormat)

		if _, ok := ss.Days[key]; !ok {
			ss.Days[key] = &DayStatics{
				UserAction: make(map[uint32]*Actions),
			}
		}
		days := ss.Days[key]

		for uid	, info := range arg.Data {
			if _, ok := days.UserAction[uid]; !ok {
				days.UserAction[uid] = &Actions{}
			}

			if _, ok := ss.History.Consumes[uid]; !ok {
				ss.History.Consumes[uid] = &UserConsumes{}
			}

			if info.Register {
				days.RegisterCount++
				days.UserAction[uid].Register = true
			}

			if info.LoginCount > 0 {
				days.LoginCount += info.LoginCount

				if days.UserAction[uid].Logined == false {
					days.DAU++
				}
				days.UserAction[uid].Logined = true
			}

			if info.RoomCard > 0 {
				days.RoomCardConsume += info.RoomCard
				ss.History.Consumes[uid].RoomCard += int64(info.RoomCard)
			}

			if info.Coin > 0 {
				days.CoinConsume += info.Coin
				ss.History.Consumes[uid].Coin += int64(info.Coin)
			}

			if info.PriceCount > 0 {
				days.PayUserCount += info.PriceAction
				if !days.UserAction[uid].Payed {
					days.PayUserCountAdd++
					days.UserAction[uid].Payed = true
				}
				days.PayCount += info.PriceAction

				if 	days.UserAction[uid].Register {
					days.PayNewUserCount += info.PriceCount
				}
				days.PayPriceCount += info.PriceCount

				ss.History.Consumes[uid].PayCount += int64(info.PriceCount)
			}
		}
	}

	reply.ErrCode = "ok"
	return nil
}

type QueryArgs struct {
	Typo 		string

	AgentType 	string
	AgentId 	uint32
	StartDay 	string
	FinishDay 	string

	QueryDay 	string
}

func (ss *StaticsServer) Query(a *QueryArgs) ([]byte, error) {
	retChan := make(chan bool, 1)
	var (
		ret []byte
		e 	error
	)

	ss.req <- func() {
		if a.Typo == "agentincome" {
			var rep proto.MsGetAgentUserReply
			ss.db.Call("DBService.GetAgentUser", &proto.MsGetAgentUserReq{
				AgentId: a.AgentId,
				AgentType: a.AgentType,
			}, &rep)

			type AgentIncome struct {
				DayTime 			string
				TotalUserCount 		int
				RoomCardConsume 	int64
				CoinConsume			int64
				PayCount			int64
			}

			as := []*AgentIncome{}
			for i := a.StartDay; i <= a.FinishDay; {
				a := &AgentIncome{}
				a.DayTime = i
				for uid, _ := range rep.Uids {
					a.TotalUserCount ++
					if user, ok := ss.History.Consumes[uid]; ok {
						a.PayCount += user.PayCount
						a.RoomCardConsume += user.RoomCard
						a.CoinConsume += user.Coin
					}
				}
				as = append(as, a)
				t, _ := time.Parse(TimeKeyFormat, i)
				t = t.Add(time.Duration(time.Hour) * 24)
				i = t.Format(TimeKeyFormat)
				fmt.Println("time i ", i)
			}

			ret, e = json.Marshal(as)
			retChan <- true

		} else if a.Typo == "statics" {

			if _, ok := ss.Days[a.QueryDay]; ok {

				type DaySort struct {
					dateKey 	string
					statics 	*DayStatics
				}
				sortDayArr := []DaySort	{}

				for dateKey, dayStatics := range ss.Days {
					sortDayArr = append(sortDayArr, DaySort{
						dateKey: dateKey,
						statics: dayStatics,
					})
				}

				sort.Slice(sortDayArr, func(i, j int) bool {
					return sortDayArr[i].dateKey < sortDayArr[j].dateKey
				})

				pivoit := -1
				wi := 0
				mi := 0
				for i := 0; i < len(sortDayArr); i++ {
					if sortDayArr[i].dateKey == a.QueryDay {
						pivoit = i
						break
					}
					if i >= 7 {
						wi++
					}
					if i >= 31 {
						mi++
					}
				}

				if pivoit != -1 {
					sc := sortDayArr[pivoit].statics
					sr := DayStatics{
						StartCount: sc.StartCount,
						StartCountNotRepeat: sc.StartCountNotRepeat,
						StartCountAdd: sc.StartCountAdd,
						LoginCount: sc.LoginCount,
						RegisterCount: sc.RegisterCount,
						DAU: sc.DAU,
						RoomCardConsume: sc.RoomCardConsume,
						CoinConsume: sc.CoinConsume,
						PayUserCount: sc.PayUserCount,
						PayUserCountAdd: sc.PayUserCountAdd,
						PayCount: sc.PayCount,
						PayNewUserCount: sc.PayNewUserCount,
						PayPriceCount: sc.PayPriceCount,
					}

					for i := pivoit; i >= mi; i-- {
						if i - wi > 0 {
							sr.WAU += sortDayArr[i].statics.DAU
						}
						sr.MAU += sr.WAU
					}
					ret, e = json.Marshal(&sr)
				} else {
					e = errors.New("not find")
				}
			} else {
				e = errors.New("not find 1")
			}

			retChan <- true
		}
	}

	<- retChan
	return ret, e
}
