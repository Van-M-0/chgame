//go:binary-only-package

package cacher

import "exportor/defines"

func NewCacheClient(group string) defines.ICacheClient {
	return newCacheClient(group)
}



