package core

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/micro"
	"github.com/micro/packet"
)

const root = `./assets/resource/versions`

// Service 初始化版本服务
func (v *Service) Init(rootUrl string) {
	// 文件下载地址
	if rootUrl == "" {
		rootUrl = "http://127.0.0.1:9000/resource/versions"
	}
	if !strings.HasSuffix(rootUrl, "/") {
		rootUrl += "/"
	}
	v.rootUrl = rootUrl

	// 加载版本文件
	os.MkdirAll(root, os.ModePerm)
	v.files = make([]versionFile, 0, 32)
	v.reloadVersionFiles()
}

// reloadVersionFiles 加载版本文件列表
func (v *Service) reloadVersionFiles() {
	fs, err := ioutil.ReadDir(root)
	if err != nil {
		micro.Debug("read version files error", err)
		return
	}
	v.files = v.files[:0]
	for _, f := range fs {
		v.files = append(v.files, versionFile{
			Code: v.getNameWithoutExt(f.Name()),
			Size: uint64(f.Size()),
			MD5:  v.getFileMD5(filepath.Join(root, f.Name())),
		})
	}
}

// getNameWithoutExt 获取文件名称(不包括扩展名)
func (v *Service) getNameWithoutExt(fileName string) string {
	for i := len(fileName) - 1; i >= 0; i-- {
		if fileName[i] == '.' {
			return fileName[:i]
		}
	}
	return fileName
}

// getFileMD5 获取文件的MD5值
func (v *Service) getFileMD5(fileName string) string {
	fd, err := os.Open(fileName)
	if err != nil {
		return ""
	}
	defer fd.Close()

	pack := packet.New(1024)
	buf := pack.Allocate(1024)
	md := md5.New()
	for {
		n, err := fd.Read(buf)
		if n > 0 {
			md.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
	packet.Free(pack)

	return hex.EncodeToString(md.Sum(nil))
}
