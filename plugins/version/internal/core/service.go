package core

import (
	"archive/zip"
	"fmt"
	"github.com/micro/packet"
	"github.com/micro/xutils"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	versionFileName = `./version`
	versionPacket   = `./packet.zip`
)

// Difference 比较目标差异文件
func (s *Service) Difference(dst string, only []string, sVersion bool) error {
	dst = strings.ReplaceAll(dst, `\`, `/`)
	tpx := strings.ReplaceAll(filepath.Join(dst, "_tmp_version"), `\`, `/`)

	// 移除发布文件
	os.RemoveAll(tpx)
	os.RemoveAll(versionPacket)

	// 解析现有版本的文件MD5值
	pack, err := packet.NewWithFile(versionFileName)
	if err != nil {
		err = nil
		pack = packet.New(2048)
	}
	fileMD5 := make(map[string]string, 256)
	for {
		fileName := pack.ReadString()
		if fileName == "" {
			break
		}
		fileMD5[fileName] = pack.ReadString()
	}

	// 迭代文件
	fs, err := ioutil.ReadDir(dst)
	if err != nil {
		return err
	}
	ps := len(dst)
	if !strings.HasSuffix(dst, `/`) {
		ps += 1
	}

	upd := false
	for _, f := range fs {
		if !s.containFile(f.Name(), only) {
			continue
		}
		ok, err := s.processFile(f, dst, ps, tpx, fileMD5)
		if err != nil {
			packet.Free(pack)
			return err
		}
		if ok {
			upd = true
		}
	}

	// 保存本次版本数据
	if upd && sVersion {
		pack.Reset()
		for fileName, md5Code := range fileMD5 {
			pack.WriteString(fileName)
			pack.WriteString(md5Code)
		}
		err = pack.SaveToFile(versionFileName)
		packet.Free(pack)
	}

	// 将临时文件打包
	if upd {
		pdf, err := os.Create(versionPacket)
		if err != nil {
			return err
		}
		w := zip.NewWriter(pdf)
		rs := len(tpx)
		if !strings.HasSuffix(tpx, `/`) {
			rs += 1
		}
		err = filepath.Walk(tpx, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			header, err := zip.FileInfoHeader(info)
			path = strings.ReplaceAll(path, `\`, `/`)
			header.Name = path[rs:]
			fmt.Println(header.Name)
			header.Method = zip.Deflate
			if err != nil {
				return err
			}
			writer, err := w.CreateHeader(header)
			if err != nil {
				return err
			}
			fd, err := os.Open(path)
			if err != nil {
				return err
			}
			_, err = io.Copy(writer, fd)
			fd.Close()
			return err
		})
		w.Close()
		pdf.Close()

		// 删除临时文件
		os.RemoveAll(tpx)
	}

	return err
}

// containFile 是否包含指定文件
func (s *Service) containFile(fileName string, only []string) bool {
	if len(only) == 0 {
		return true
	}
	return xutils.HasString(only, fileName)
}

// processFile 处理文件
func (s *Service) processFile(f os.FileInfo, cur string, ps int, tpx string, fileMD5 map[string]string) (bool, error) {
	if f.IsDir() {
		dr := filepath.Join(cur, f.Name())
		fs, err := ioutil.ReadDir(dr)
		if err != nil {
			return false, err
		}
		for _, c := range fs {
			if ok, err := s.processFile(c, dr, ps, tpx, fileMD5); err != nil {
				return ok, err
			}
		}
		return false, nil
	}

	// 处理文件
	fileName := filepath.Join(cur, f.Name())
	md5Code := s.getFileMD5(fileName)
	fn := strings.ReplaceAll(fileName[ps:], `\`, `/`)
	if oMd5Code, ok := fileMD5[fn]; ok && oMd5Code == md5Code {
		// 文件没有变化
		return false, nil
	}
	fileMD5[fn] = md5Code

	// 复件文件
	fmt.Println("copy file=>", fn)
	dst := filepath.Join(tpx, fn)
	os.MkdirAll(filepath.Dir(dst), os.ModePerm)
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return false, err
	}
	return true, ioutil.WriteFile(dst, data, os.ModePerm)
}
