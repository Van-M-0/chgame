package world

import (
	"exportor/defines"
	"github.com/jinzhu/gorm"
	"fmt"
	"mylog"
)

/*
*	table defines begin
*/
type T_GameArea struct {
	Id 			int 		`gorm:"primary_key;AUTO_INCREMENT"`
	Area 		string
	City 		string
	Province 	string
	Source 		string
	Database 	string
	Game 		string
}

/*	table defines end
*/

//CREATE DATABASE IF NOT EXISTS gamemaster default charset utf8 COLLATE utf8_general_ci;
type dbClient struct {
	opt 		*defines.DatabaseOption
	db 			*gorm.DB
	uri 		string
}

func newDbClient(cfg *defines.StartConfigFile) *dbClient {
	dc := &dbClient{}
	opt := &defines.DatabaseOption{
		Host: "127.0.0.1:3306",
		User: cfg.DbUser,
		Pass: cfg.DbPwd,
		Name: cfg.DbName,
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

func (dc *dbClient) InitTable() {
	if dc.db.HasTable(&T_GameArea{}) == false {
		dc.db.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8").CreateTable(&T_GameArea{})
		dc.db.Create(&T_GameArea{
			Area: "成都",
			City: "成都市",
			Province: "四川省",
			Source: "192.168.1.123",
			Database: "mygame",
			Game: "斗地主",
		}).Create(&T_GameArea{
			Area: "成都",
			City: "成都市",
			Province: "四川省",
			Source: "192.168.1.123",
			Database: "mygame",
			Game: "血战麻将",
		}).Create(&T_GameArea{
			Area: "绵阳",
			City: "绵阳市",
			Province: "四川省",
			Source: "192.168.1.123",
			Database: "mygame",
			Game: "绵阳麻将",
		})
	}
}
