package cacher

import "strconv"

func accountId(name string) string {
	return "accs."+name+":uid"
}

func users(uid int) string {
	return "users."+strconv.Itoa(uid)
}


