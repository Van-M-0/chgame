//go:binary-only-package-my

package dbproxy

import "exportor/defines"

func NewDbProxy(opt *defines.DbProxyOption) defines.IDbProxy {
	return newDBProxyServer(opt)
}