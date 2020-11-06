package core

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/micro"
	"github.com/micro/packet"
)

const root = `./assets/resource/versions`

// versionService 版本服务
type VersionService struct {
	sync.RWMutex

	files   []VersionFile // 文件列表
	rootUrl string        // 文件URL
}

type VersionFile struct {
	Code string
	Size uint64
	MD5  string
}

// Init 初始化版本文件列表
func (v *VersionService) Init(rootUrl string) {
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
	v.files = make([]VersionFile, 0, 32)
	v.ReloadVersionFiles()
}

type RespVersion struct {
	Code string
	Url  string
	MD5  string
	Size uint64
}

type ReqVersion struct {
	// 版本号
	Code string
}

// CheckVersion 校验版本号
func (v *VersionService) CheckVersion(dpo micro.Dpo) (interface{}, string) {
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
func (v *VersionService) UploadFile(reader io.Reader) error {
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
func (v *VersionService) ReloadVersionFiles() {
	v.Lock()
	v.reloadVersionFiles()
	v.Unlock()
}

func (v *VersionService) reloadVersionFiles() {
	fs, err := ioutil.ReadDir(root)
	if err != nil {
		micro.Debug("read version files error", err)
		return
	}
	v.files = v.files[:0]
	for _, f := range fs {
		v.files = append(v.files, VersionFile{
			Code: v.getNameWithoutExt(f.Name()),
			Size: uint64(f.Size()),
			MD5:  v.getFileMD5(filepath.Join(root, f.Name())),
		})
	}
}

// getNameWithoutStu
func (v *VersionService) getNameWithoutExt(fileName string) string {
	for i := len(fileName) - 1; i >= 0; i-- {
		if fileName[i] == '.' {
			return fileName[:i]
		}
	}
	return fileName
}

// getFileMD5 获取文件的MD5值
func (v *VersionService) getFileMD5(fileName string) string {
	fd, err := os.Open(fileName)
	if err != nil {
		return ""
	}
	defer fd.Close()

	pack := packet.New(1024)
	md := md5.New()
	buf := pack.Allocate(1024)
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
