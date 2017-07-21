//go:binary-only-package-my

package cacher

import "exportor/defines"

func NewCacheClient(group string) defines.ICacheClient {
	return newCacheClient(group)
}



