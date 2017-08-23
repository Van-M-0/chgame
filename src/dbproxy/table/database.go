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
	Userid 		uint32		`gorm:"primary_key;AUTO_INCREMENT:10000;not null"`
	Account 	string 		`gorm:"size:32;not null;default:'';index:acc_index"`
	OpenId 		string		`gorm:"index:openid_index"`
	Name 		string 		`gorm:"size:32;default:null"`
	Sex 		uint8		`gorm:"not null"`
	Headimg 	string		`gorm:"size:64;default:''"`
	Level 		uint8 		`gorm:"default:1"`
	Exp 		uint32		`gorm:"default:0"`
	Diamond 	uint32 		`gorm:"default:0"`
	RoomCard 	uint32		`gorm:"default:0"`
	Gold 		int64		`gorm:"default:0"`
	Score 		uint32		`gorm:"default:0"`
	Roomid 		uint32 		`gorm:"default:0"`
	History 	string 		`gorm:"size:4096;not null;default:''"`
	Agentid		int			`gorm:"default:0"`
	Regtime 	time.Time
	Ip 			string 		`gorm:"default:''"`
}

type T_UserItem struct {
	Userid 		uint32 		`gorm:"index:uid_index"`
	Itemid 		uint32 		`gorm:"index:item_index"`
	Count 		int 		`gorm:"not null; default:0"`
}

type T_ItemConfig struct {
	Itemid		uint32 		`gomr:"primary_key"`
	Itemname 	string		`gorm:"size:32"`
	Category 	int 		`gorm:"not null"`	//种类 1钻石，2房卡
	Nums 		int 		`gorm:"not null"`
	Sell 		int 		`gorm:"not null"`	//1 商店显示
	Buyvalue 	int 		`gorm:"not null"`
	GameKind 	int
	Description string 		`gorm:"default:''"`
}

/*
type T_MallItem struct {
	Itemid 		int 		`gorm:"primary_key;not null"`
	Itemname	string 		`gorm:"size:32;not null"`
	Category 	int 		`gorm:"not null"`
	Buyvalue 	int 		`gorm:"not null"`
	Nums 		int 		`gorm:"not null;default:1"`
	Limit 		int 		`gorm:"not null;default:0"`
}
*/

type T_Notice struct {
	Index 		int 		`gorm:"primary_key;AUTO_INCREMENT;not null"`
	Starttime 	time.Time
	Finishtime	time.Time
	Kind 		string 		`gorm:"size:16;default:'world'"`
	Content 	string 		`gorm:"size:128;not null"`
	Playtime 	int 		`gorm:"not null;default:0"`
	Playcount 	int			`gorm:"not null;default:1"`
}

type T_Gamelib struct {
	Id 			int			`gorm:"primary_key"`
	Name 		string
	Area 		string
	City 		string
	Province 	string
}

type T_Activity struct {
	Id 			int 		`gorm:"primary_key;"`
	Desc 		string
	Actype 		string
	Starttime 	time.Time
	Finishtime 	time.Time
	Rewardids 	string
}

type T_ActivityReward struct {
	Id 			int 		`gorm:"primary_key;"`
	RewardType 	string
	ItemId 		int
	Num 		int
}

type T_Userdata struct {
	Userid 		uint32 		`gorm:"primary_key"`
	Data 		[]byte		`gorm:"size:10240"`
}

type T_Quest struct {
	Id 			int			`gorm:"primary_key"`
	Title 		string
	Content 	string
	Type 		string
	MaxCount 	int
	RewardIds 	string
}

type T_QuestReward struct {
	Id 			int 		`gorm:"primary_key"`
	ItemId 		int
	Num 		int
}

type T_ActionForbid struct {
	Userid 		uint32		`gorm:"primary_key;not null"`
	SpeakForbid	int			`gorm:"default:0"`
	SfStartime 	time.Time
	SfFinishtime 	time.Time
	LoginForbid int			`gorm:"default:0"`
	LfStarttime time.Time
	LfFinishtime time.Time
}

type T_AuthInfo struct {
	Userid 		uint32		`gorm:"primary_key"`
	Phone 		string
	Idcard 		string
	Name 		string
}
