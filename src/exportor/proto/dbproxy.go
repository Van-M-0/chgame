package proto

type PMLoadUser struct {
	Acc 		string
}

type PMLoadUserFinish struct {
	Acc 		string
	Err 		error
	Code 		int
}


