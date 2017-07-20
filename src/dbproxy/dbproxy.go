//go:binary-only-package

package dbproxy

import "exportor/defines"

func NewDbProxy() defines.IDbProxy {
	return newDBProxyServer()
}