package proto

type PMLoadUser struct {
	Acc 		string
	Create 		bool
	Guest 		bool
}

type PMLoadUserFinish struct {
	Acc 		string
	Err 		error
	Code 		int
}

type PMCreateAccount struct {
	Name 		string
	Sex 		byte
	Pwd 		string
}

type PMCreateAccountFinish struct {
	Err 		int
	Account 	string
	Pwd 		string
}
