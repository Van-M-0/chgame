package dbproxy

import (
	"exportor/defines"
	"fmt"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/jinzhu/gorm"
	"dbproxy/table"
	"time"
	"mylog"
)

//CREATE DATABASE IF NOT EXISTS mygame default charset utf8 COLLATE utf8_general_ci;
type dbClient struct {
	opt 		*defines.DatabaseOption
	db 			*gorm.DB
	uri 		string
}

func InitTables() {
	dc := newDbClient()
	dc.InitTable()
}

func Test() {
	/*
	dc := newDbClient()

	type TestAAA struct {
		UserId 		int			`gorm:"primary_key;"`
		Conf 		[]byte		`gorm:"size:4096"`
	}

	dc.CreateTableIfNot(&TestAAA{})


	str := "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890" +
		"1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890" +
		"12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901222222"
	d := []byte(str)

	dc.db.Save(&TestAAA{
		UserId: 1001,
		Conf: d,
	})


	var a TestAAA
	dc.db.Where("user_id = ? ", 1001).Find(&a)

	mylog.Debug(a, len(a.Conf) , string(a.Conf))
	*/

/*
	dc := newDbClient()

	dc.CreateTableIfNot(&table.T_MallItem{})

	dc.db.Create(&table.T_MallItem{
		Itemid: 1,
		Itemname: "种类1-钻石-商品1",
		Category: 1,
		Buyvalue: 100,
		Nums: 1,
	}).Create(&table.T_MallItem{
		Itemid: 201,
		Itemname: "种类2-金币-商品1",
		Category: 2,
		Buyvalue: 100,
		Nums: 10,
	}).Create(&table.T_MallItem{
		Itemid: 301,
		Itemname: "种类3-房卡-商品1",
		Category: 3,
		Buyvalue: 100,
		Nums: 100,
	})
	*/

	type t_Myrank struct {
		UserId 		int
		Score 		int
	}

/*
	dc := newDbClient()
	dc.CreateTableIfNot(&t_Myrank{})
	dc.db.Create(&t_Myrank{
		UserId: 1,
		Score: 100,
	}).Create(&t_Myrank{
		UserId: 2,
		Score: 200,
	}).Create(&t_Myrank{
		UserId: 3,
		Score: 300,
	}).Create(&t_Myrank{
		UserId: 4,
		Score: 400,
	}).Create(&t_Myrank{
		UserId: 5,
		Score: 500,
	})

	var ranks []t_Myrank
	dc.db.Order("score desc").Limit(3).Find(&ranks)
	mylog.Debug(ranks)
	*/
}

func newDbClient() *dbClient {

	dc := &dbClient{}

	opt := &defines.DatabaseOption{
		Host: "127.0.0.1:3306",
		User: DbProxyOptoins.User,
		Pass: DbProxyOptoins.Pwd,
		Name: DbProxyOptoins.Name,
		DetailLog: true,
		Singular: true,
	}

	uri := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8&parseTime=True",
		opt.User,
		opt.Pass,
		opt.Host,
		opt.Name,
	)

	mylog.Debug("db proxy connection info ", uri)
	db, err := gorm.Open("mysql", uri)
	if err != nil {
		mylog.Debug("create db proxy err ", err)
		return nil
	}

	if opt.DetailLog {
		db.LogMode(true)
	}

	if opt.Singular {
		db.SingularTable(true)
	}

	dc.opt = opt
	dc.db = db
	dc.uri = uri
	dc.InitTable()
	return dc
}

func (dc *dbClient) CreateTable(v ...interface{}) {
	dc.db.CreateTable(v...)
}

func (dc *dbClient) CreateTableIfNot(v ...interface{}) {
	for _, m := range v {
		if dc.db.HasTable(m) == false {
			dc.db.CreateTable(m).Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8")
		}
	}
}

func (dc *dbClient) CreateTableIfNotWithOption(v interface{}, options string) {
	if dc.db.HasTable(v) == false {
		dc.db.Set("gorm:table_options", options).CreateTable(v)
	}
}

func (dc *dbClient) CreateTableForce(v...interface{}) {
	dc.db.DropTableIfExists(v...)
	dc.db.CreateTable(v...)
}

func (dc *dbClient) DropTable(v ...interface{}) {
	dc.db.DropTableIfExists(v...)
}

// logic handler

func (dc *dbClient) InitTable() {
	mylog.Debug("init tables")
/*
	dc.DropTable(&table.T_Accounts{})
	dc.DropTable(&table.T_Games{})
	dc.DropTable(&table.T_GamesArchive{})
	dc.DropTable(&table.T_Guests{})
	dc.DropTable(&table.T_Message{})
	dc.DropTable(&table.T_Rooms{})
	dc.DropTable(&table.T_RoomUser{})
	dc.DropTable(&table.T_Users{})
	dc.DropTable(&table.T_MyTest{})
*/
	if !dc.db.HasTable(&table.T_ItemConfig{}) {
		dc.CreateTable(&table.T_ItemConfig{})
		dc.db.Create(&table.T_ItemConfig{
			Itemid: 1,
			Itemname: "钻石",
			Category: 1,
			Nums: 10,
			Sell: 1,
			Buyvalue: 10,
			GameKind: 1,
			Description: "钻石-1",
		}).Create(&table.T_ItemConfig{
			Itemid: 2,
			Itemname: "房卡",
			Category: 2,
			Nums: 10,
			Sell: 1,
			Buyvalue: 10,
			GameKind: 1,
			Description: "钻石-1",
		}).Create(&table.T_ItemConfig{
			Itemid: 3,
			Itemname: "房卡-物品",
			Category: 3,
			Nums: 10,
			Sell: 1,
			Buyvalue: 10,
			GameKind: 1,
			Description: "道具",
		})
	}

	if !dc.db.HasTable(&table.T_Gamelib{}) {
		dc.db.CreateTable(&table.T_Gamelib{})
		dc.db.Create(&table.T_Gamelib{
			Id: 1,
			Name: "血战麻将",
			Area: "成都",
			City: "成都市",
			Province: "四川省",
		})
	}

	if !dc.db.HasTable(&table.T_Activity{}) {
		dc.db.CreateTable(&table.T_Activity{})
		dc.db.Create(&table.T_Activity{
			Id: 101,
			Desc: "首充有礼",
			Actype: "always",
			Starttime: time.Now(),
			Finishtime: time.Now(),
			Rewardids: "1",
		}).Create(&table.T_Activity{
			Id: 102,
			Desc: "1、玩家参与活动需加入官方用户微信群并关注“约8棋牌平台”公众号；\n2、每周日23:59，“约8棋牌平台”更新一组牌型；\n3、玩家需进入“约8棋牌平台”输入“本周牌型”即可显示出当周的奖励牌型图例；\n4、玩家在游戏完成当周的奖励牌型后，截该奖励牌型的结算界面发送至内测微信群；\n5、并且提供自己微信号，由官方人员进行记录，以便不删档测试后发放奖励；\n6、每个玩家每天只能领取一次奖励，不能重复获得奖励；\n7、更多活动详细介绍请关注“约8棋牌平台”微信公众号，点击“热门活动”查看。",
			Actype: "always",
			Starttime: time.Now(),
			Finishtime: time.Now(),
			Rewardids: "2",
		})

		dc.db.CreateTable(&table.T_ActivityReward{})
		dc.db.Create(&table.T_ActivityReward{
			Id: 1,
			RewardType: "addition",
			Num: 1,
		}).Create(&table.T_ActivityReward{
			Id: 2,
			RewardType: "addition",
			ItemId: 1,
			Num: 5,
		})
	}

	if !dc.db.HasTable(&table.T_Quest{}) {
		dc.db.CreateTable(&table.T_Quest{})
		dc.db.Create(&table.T_Quest{
			Id: 101,
			Title: "分享有礼"	,
			Content: "分享，即可获得丰厚礼品",
			Type: "counter",
			MaxCount: 1,
			RewardIds: "1",
		})

		dc.db.CreateTable(&table.T_QuestReward{})
		dc.db.Create(&table.T_QuestReward{
			Id: 1,
			ItemId: 1,
			Num: 1,
		})
	}

	if !dc.db.HasTable(&table.T_Notice{}) {
		dc.db.CreateTable(&table.T_Notice{})
		dc.db.Create(&table.T_Notice{
			Index: 20,
			Starttime: time.Now(),
			Finishtime: time.Now().Add(time.Hour * 240),
			Kind: "notice",
			Content:"每日一播",
			Playtime: 3600,
			Playcount: 1,
		})
	}

	dc.CreateTableIfNot(&table.T_Accounts{})
	dc.CreateTableIfNot(&table.T_Games{})
	dc.CreateTableIfNot(&table.T_GamesArchive{})
	dc.CreateTableIfNot(&table.T_Guests{})
	dc.CreateTableIfNot(&table.T_Rooms{})
	dc.CreateTableIfNot(&table.T_RoomUser{})
	dc.CreateTableIfNotWithOption(&table.T_Users{}, "ENGINE=InnoDB CHARSET=utf8 AUTO_INCREMENT=100000")
	dc.CreateTableIfNot(&table.T_Notice{})
	dc.CreateTableIfNot(&table.T_UserItem{})
	dc.CreateTableIfNot(&table.T_Userdata{})
	dc.CreateTableIfNot(&table.T_ActionForbid{})
	dc.CreateTableIfNot(&table.T_AuthInfo{})
	dc.CreateTableIfNot(&table.T_Club{})
	dc.CreateTableIfNot(&table.T_ClubMember{})
	dc.CreateTableIfNot(&table.T_Agents{})
}

// t_accounts : account info
func (dc *dbClient) GetAccountInfo(account string, accInfo *table.T_Accounts) bool {
	return dc.db.Where(&table.T_Accounts{Account: account}).First(accInfo).RowsAffected != 0
}

func (dc *dbClient) AddAccountInfo(accInfo *table.T_Accounts) bool {
	return dc.db.Create(accInfo).RowsAffected != 0
}

// t_users : user info
func (dc *dbClient) AddUserInfo(userInfo *table.T_Users) bool {
	mylog.Debug("add user info : ", userInfo)
	return dc.db.Create(userInfo).RowsAffected != 0
}

func (dc *dbClient) GetUserInfo(account string, userInfo *table.T_Users) bool {
	return dc.db.Where("account = ? ", account).
		Find(&userInfo).
		RowsAffected != 0
}

func (dc *dbClient) GetUserInfoByName(name string, users *table.T_Users) bool {
	return dc.db.Where("name = ?", name).
		Find(&users).
		RowsAffected != 0
}

func (dc *dbClient) GetUserInfoByUserid(userid uint32, userInfo *table.T_Users) bool {
	return dc.db.Where("userid = ? ", userid).
		Find(&userInfo).
		RowsAffected != 0
}

func (dc *dbClient) ModifyUserInfo(userid uint32, userInfo *table.T_Users) bool {
	return dc.db.Model(&table.T_Users{}).
		Where("userid = ?", userid).
		Update(userInfo).
		RowsAffected != 0
}

func (dc *dbClient) GetUserHistoryByUserid(userid uint32, userInfo *table.T_Users) bool {
	return dc.db.Where("userid = ? ", userid).
		Select("history").
		Find(&userInfo).
		RowsAffected != 0
}

func (dc *dbClient) GetUserGemsByUserid(userid uint32, userInfo *table.T_Users) bool {
	return dc.db.Where("userid = ? ", userid).
		Select("gems").
		Find(&userInfo).
		RowsAffected != 0
}

func (dc *dbClient) GetUserBaseInfo(userid uint32, userInfo *table.T_Users) bool {
	return dc.db.Where("userid = ? ", userid).
		Select("name, sex, headimg").
		Find(&userInfo).
		RowsAffected != 0
}

// t_rooms : room info
func (dc *dbClient) GetRoomInfo(roomid string, roomInfo *table.T_Rooms) bool {
	return dc.db.Where(&table.T_Rooms{Id: roomid}).First(roomInfo).RowsAffected != 0
}
