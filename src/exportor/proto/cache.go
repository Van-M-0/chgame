package proto

// account -> userid
type CacheUserId struct {
	Uid 		int
}

// userid -> cacheuser
type CacheUser struct {
	Account 	string		`redis:"account"`
	Openid 		string		`redis:"openid"`
	Pwd 		string		`redis:"pwd"`
	Uid 		int			`redis:"uid"`
	Name 		string		`redis:"name"`
	Sex 		byte		`redis:"sex"`
	HeadImg 	string		`redis:"headimg"`
	Diamond 	int			`redis:"diamond"`
	RoomCard 	int			`redis:"roomcard"`
	Gold 		int64		`redis:"gold"`
	RoomId 		int 		`redis:"roomid"`
}


