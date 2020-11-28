package core

import "sync"

// Service 版本管理服务
type Service struct {
	sync.RWMutex

	files   []versionFile // 文件列表
	rootUrl string        // 文件URL
}

// versionFile 版本文件描述
type versionFile struct {
	Code string
	Size uint64
	MD5  string
}

// RespVersion 版本响应数据
type RespVersion struct {
	Code string
	Url  string
	MD5  string
	Size uint64
}

// ReqVersion 版本请求参数
type ReqVersion struct {
	// 版本号
	Code string
}