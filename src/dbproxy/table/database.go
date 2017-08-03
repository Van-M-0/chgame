package table

import "time"

type T_Accounts struct {
	Account 	string		`gorm:"primary_key"`
	Password	string
}

type T_Games struct {
	RoomUuid	string		`gorm:"size:20;type:char(20);primary_key"`
	GameIndex 	int			`gorm:"type:int unsigned;primary_key"`
	BaseInfo 	string		`gorm:"size:1024;varchar(1024);not null"`
	CreateTime 	int			`gorm:"not null"`
	Snapshots	string 		`gorm:"size:2048;varchar(2048);default:null"`
	ActionRecords string 	`gorm:"varchar(2048);default:null"`
	Result 		string 		`gorm:"default:null"`
}

type T_GamesArchive struct {
	RoomUuid	string		`gorm:"size:20;type:char(20);primary_key"`
	GameIndex 	int			`gorm:"type:int unsigned;primary_key"`
	BaseInfo 	string		`gorm:"size:1024;varchar(1024);not null"`
	CreateTime 	int			`gorm:"not null"`
	Snapshots	string 		`gorm:"size:2048;varchar(2048);default:null"`
	ActionRecords string 	`gorm:"varchar(2048);default:null"`
	Result 		string 		`gorm:"default:null"`
}

type T_Guests struct {
	GuestAccount string 	`gorm:"size:255;type:varchar(255);primary_key"`
}

type T_Rooms struct {
	Uuid 		string 		`gorm:"size:20;type:char(20);not null;primary_key"`
	Id 			string 		`gorm:"size:8;not null"`
	BaseInfo 	string 		`gorm:"size:256;type:varchar(256);not null;default:0"`
	CreateTime 	int 		`gorm:"not null"`
	NumOfTurns 	int 		`gorm:"not null;default:'0'"`
	NextButton	int 		`gorm:"not null;default:'0'"`
	User1		int			`gorm:"not null"`
	User2		int			`gorm:"not null"`
	User3		int			`gorm:"not null"`
	User4		int			`gorm:"not null"`
}

type T_RoomUser struct {
	UserIndex 	int 		`gorm:"primary_key;AUTO_INCREMENT"`
	UserID 		int 		`gorm:"not null"`
	UserIcon 	string 		`gorm:"size:128;not null;default:''"`
	UserName 	string 		`gorm:"size:32;not null;default:''"`
	UserScore 	int 		`gorm:"not null;default:0"`
}

type T_Users struct {
	Userid 		uint32		`gorm:"primary_key;AUTO_INCREMENT;not null"`
	Account 	string 		`gorm:"size:32;not null;default:'';index:acc_index"`
	OpenId 		string		`gorm:"index;openid_index"`
	Name 		string 		`gorm:"size:32;default:null"`
	Sex 		uint8
	Headimg 	string
	Level 		uint8 		`gorm:"default:1"`
	Exp 		uint32		`gorm:"default:0"`
	Diamond 	uint32 		`gorm:"default:0"`
	RoomCard 	uint32		`gorm:"default:0"`
	Gold 		int64		`gorm:"default:0"`
	Score 		uint32		`gorm:"default:0"`
	Roomid 		uint32 		`gorm:"default:0"`
	History 	string 		`gorm:"size:4096;not null;default:''"`
}

type T_UserItem struct {
	Itemid 		uint32 		`gorm:"primary_key"`
	Userid 		int 		`gorm:"index:user_index;not null"`
	Count 		int 		`gorm:"not null; default:0"`
}

type T_ItemConfig struct {
	Itemid		uint32 		`gomr:"primary_key"`
	Itemname 	string		`gorm:"size:32"`
	Category 	int 		`gorm:"not null"`	//种类 1钻石，2房卡
	Nums 		int 		`gorm:"not null"`
	Sell 		int 		`gorm:"not null"`	//1 商店显示
	Buyvalue 	int 		`gorm:"not null"`
	Area	 	int 		`gorm:"not null"`	//所在区域
	Description string 		`gorm:"default:'';"`
}

type T_ItemArea struct {
	Area 		int 		`gorm:"not null"`
	Gamelib 	int 		`gorm:"not null"`
}

type T_MallItem struct {
	Itemid 		int 		`gorm:"primary_key;not null"`
	Itemname	string 		`gorm:"size:32;not null"`
	Category 	int 		`gorm:"not null"`
	Buyvalue 	int 		`gorm:"not null"`
	Nums 		int 		`gorm:"not null;default:1"`
	Limit 		int 		`gorm:"not null;default:0"`
}

type T_Notice struct {
	Index 		int 		`gorm:"primary_key;AUTO_INCREMENT;not null"`
	Starttime 	time.Time
	Finishtime	time.Time
	Kind 		string 		`gorm:"size:16;default:'world'"`
	Content 	string 		`gorm:"size:128;not null"`
	Playtime 	int 		`gorm:"not null;default:0"`
	Playcount 	int			`gorm:"not null;default:1"`
}