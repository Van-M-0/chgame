package proto

// account -> userid
type CacheUserId struct {
	Uid 		int
}

// userid -> cacheuser
type CacheUser struct {
	Account 	string
	Openid 		string
	Pwd 		string
	Uid 		int
	Name 		string
	Sex 		byte
	HeadImg 	string
	Diamond 	int
	RoomCard 	int
	Golden 		int64
}


