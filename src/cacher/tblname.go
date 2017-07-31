package cacher

import "strconv"

func accountId(name string) string {
	return "accs."+name+":uid"
}

func users(uid int) string {
	return "users."+strconv.Itoa(uid)
}

func servers(id int) string {
	return "servers."+strconv.Itoa(id)
}

func serversPattern() string {
	return "servers.*"
}

func ciduserid(uid uint32) string {
	return "cids."+strconv.Itoa(int(uid))
}
