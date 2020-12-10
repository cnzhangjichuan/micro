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
func (s *Service) CheckVersion(dpo micro.Dpo) (interface{}, string) {
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
	s.RLock()
	for i := 0; i < len(s.files); i++ {
		if s.files[i].Code > req.Code {
			resp = append(resp, RespVersion{
				Code: s.files[i].Code,
				Size: s.files[i].Size,
				MD5:  s.files[i].MD5,
				Url:  s.rootUrl + s.files[i].Code + ".zip",
			})
		}
	}
	s.RUnlock()

	return &resp, ""
}

// UploadFile 上传版本文件
func (s *Service) UploadFile(reader io.Reader) error {
	s.Lock()

	// save file
	fileName := time.Now().Format("20060102150405")
	fd, err := os.Create(filepath.Join(root, fileName+".zip"))
	if err != nil {
		s.Unlock()
		return err
	}

	pack := packet.New(1024)
	_, err = io.CopyBuffer(fd, reader, pack.Allocate(1024))
	packet.Free(pack)
	fd.Close()

	// reload version files.
	s.reloadVersionFiles()
	s.Unlock()
	if err == io.EOF {
		return nil
	}
	return err
}

// reloadVersionFiles 重新加载文件配置
func (s *Service) ReloadVersionFiles() {
	s.Lock()
	s.reloadVersionFiles()
	s.Unlock()
}
