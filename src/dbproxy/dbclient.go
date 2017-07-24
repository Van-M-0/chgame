package dbproxy

import (
	"exportor/defines"
	"fmt"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/jinzhu/gorm"
	"dbproxy/table"
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

func newDbClient() *dbClient {

	dc := &dbClient{}

	opt := &defines.DatabaseOption{
		Host: "127.0.0.1:3306",
		User: "root",
		Pass: "1",
		Name: "mygame",
		DetailLog: true,
		Singular: true,
	}

	uri := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8&parseTime=True",
		opt.User,
		opt.Pass,
		opt.Host,
		opt.Name,
	)

	fmt.Println("db proxy connection info ", uri)
	db, err := gorm.Open("mysql", uri)
	if err != nil {
		fmt.Println("create db proxy err ", err)
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

func (dc *dbClient) CreateTableForce(v...interface{}) {
	dc.db.DropTableIfExists(v...)
	dc.db.CreateTable(v...)
}

func (dc *dbClient) DropTable(v ...interface{}) {
	dc.db.DropTableIfExists(v...)
}

// logic handler

func (dc *dbClient) InitTable() {
	fmt.Println("init tables")
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
	dc.CreateTableIfNot(&table.T_Accounts{})
	dc.CreateTableIfNot(&table.T_Games{})
	dc.CreateTableIfNot(&table.T_GamesArchive{})
	dc.CreateTableIfNot(&table.T_Guests{})
	dc.CreateTableIfNot(&table.T_Message{})
	dc.CreateTableIfNot(&table.T_Rooms{})
	dc.CreateTableIfNot(&table.T_RoomUser{})
	dc.CreateTableIfNot(&table.T_Users{})
	dc.CreateTableIfNot(&table.T_MyTest{})
}

func (dc *dbClient) PreLoadData() {

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
	fmt.Println("add user info : ", userInfo)
	return dc.db.Create(userInfo).RowsAffected != 0
}

func (dc *dbClient) GetUserInfo(account string, userInfo *table.T_Users) bool {
	return dc.db.Where("account = ? ", account).
		Select("userid, account, name, sex, headimg, level, exp, coins, gems, roomid").
		Find(&userInfo).
		RowsAffected != 0
}

func (dc *dbClient) GetUserInfoByName(name string, users *table.T_Users) bool {
	return dc.db.Where("name = ?", name).
		Select("userid, account, name, sex, headimg, level, exp, coins, gems, roomid").
		Find(&users).
		RowsAffected != 0
}

func (dc *dbClient) GetUserInfoByUserid(userid uint32, userInfo *table.T_Users) bool {
	return dc.db.Where("userid = ? ", userid).
		Select("userid, account, name, sex, headimg, level, exp, coins, gems, roomid").
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
