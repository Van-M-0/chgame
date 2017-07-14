package cacher

import "exportor/defines"

type cacheClient struct {

}

func newCacheClient() *cacheClient {
	return &cacheClient{}
}

func (cc *cacheClient) Start() {

}

func (cc *cacheClient) Stop() {

}

// ICacheClient
func (cc *cacheClient) SetCacheNotify(notify defines.ICacheNotify) {

}

func (cc *cacheClient) GetUserInfo(name string) {

}

// ICacheLoader



