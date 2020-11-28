package core

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/micro"
	"github.com/micro/packet"
)

// CheckVersion 校验版本号
func (v *Service) CheckVersion(dpo micro.Dpo) (interface{}, string) {
	var (
		resp = make([]RespVersion, 0, 8)
		req  ReqVersion
	)

	// 获取参数
	dpo.Parse(&req)

	// 没有版本号
	if req.Code == "" {
		return &resp, ""
	}

	// 将版本号大于传入版本号的文件加入列表中
	v.RLock()
	for i := 0; i < len(v.files); i++ {
		if v.files[i].Code > req.Code {
			resp = append(resp, RespVersion{
				Code: v.files[i].Code,
				Size: v.files[i].Size,
				MD5:  v.files[i].MD5,
				Url:  v.rootUrl + v.files[i].Code + ".zip",
			})
		}
	}
	v.RUnlock()

	return &resp, ""
}

// UploadFile 上传版本文件
func (v *Service) UploadFile(reader io.Reader) error {
	v.Lock()

	// save file
	fileName := time.Now().Format("20060102150405")
	fd, err := os.Create(filepath.Join(root, fileName+".zip"))
	if err != nil {
		v.Unlock()
		return err
	}

	pack := packet.New(1024)
	_, err = io.CopyBuffer(fd, reader, pack.Allocate(1024))
	packet.Free(pack)
	fd.Close()

	// reload version files.
	v.reloadVersionFiles()
	v.Unlock()
	if err == io.EOF {
		return nil
	}
	return err
}

// reloadVersionFiles 重新加载文件配置
func (v *Service) ReloadVersionFiles() {
	v.Lock()
	v.reloadVersionFiles()
	v.Unlock()
}
