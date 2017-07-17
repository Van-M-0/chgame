package proto

// lobby proto 1000 - 2000

const (
	CmdClientLogin 				= 1001
	CmdClientLoginRet			= 1002
)

type ClientLogin struct {
	Account 	string
}

type ClientLoginRet struct {
	ErrCode 	uint8
	Account 	string
	Name 		string
	UserId 		uint32
}