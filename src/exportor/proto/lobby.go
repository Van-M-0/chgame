package proto

// lobby proto 1000 - 2000

const (
	CmdClientLogin   = 1001
	CmdGuestLogin    = 1002
	CmdCreateAccount = 1003
)

type ClientLogin struct {
	Account 	string
}

type GuestLogin struct {
	Account 	string
}

type ClientLoginRet struct {
	ErrCode 	int
	Uid 		uint32
	Account 	string
	Name 		string
	UserId 		uint32
}

type CreateAccount struct {
	Name 		string
	Sex 		byte
}

type CreateAccountRet struct {
	ErrCode 	int
	Account 	string
	Pwd 		string
}