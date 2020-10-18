package micro

import (
	"errors"
	"io"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/micro/packet"
	"github.com/micro/xutils"
)

type uploader struct {
	baseChain
}

// Handle 处理Conn
func (c *uploader) Handle(conn net.Conn, name string, pack *packet.Packet) bool {
	if name != "uploader" {
		return false
	}

	auc := pack.HTTPHeaderValue(httpAuthorize)
	n := xutils.ParseI64(pack.HTTPHeaderValue(httpContentLength), 0)

	// authorize check
	api, ok := env.authorize.Check(auc)
	if !ok {
		pack.BeginWrite()
		pack.WriteU32(0)
		pack.WriteString("trespassing")
		pack.EndWrite()
		pack.FlushToConn(conn)
		return true
	}

	// 查找API
	f, ok := env.uploadFunc[api]
	if !ok {
		pack.BeginWrite()
		pack.WriteU32(0)
		pack.WriteString("not found api")
		pack.EndWrite()
		pack.FlushToConn(conn)
		return true
	}

	// 通知远端，开始上传文件
	pack.BeginWrite()
	pack.WriteU32(1)
	pack.EndWrite()
	pack.FlushToConn(conn)

	// 开始接收数据
	err := f(io.LimitReader(conn, n))
	pack.BeginWrite()
	if err != nil {
		pack.WriteString("upload file error:" + err.Error())
	} else {
		pack.WriteString(`file has been uploaded.`)
	}
	pack.EndWrite()
	pack.FlushToConn(conn)

	return true
}

type uploadFunc func(io.Reader) error

// RegisterUploadFunc 文件上传
func RegisterUploadFunc(name string, f uploadFunc) {
	env.uploadFunc[name] = func(reader io.Reader) (err error) {
		defer func() {
			er := recover()
			if er == nil {
				return
			}
			pack := packet.New(1024)
			buf := pack.Allocate(1024)
			buf = buf[:runtime.Stack(buf, false)]
			Debug("\nupload error: %v\n%s\n\n", er, buf)
			packet.Free(pack)
			err = errUploadError
		}()
		err = f(reader)
		return
	}
}

// 上传文件
func requestUploadService(api, fileName string) error {
	conn, err := net.DialTimeout("tcp", localeAddress(), time.Second)
	if err != nil {
		return errors.New("service not found, it may be closed")
	}

	fd, err := os.Open(fileName)
	if err != nil {
		conn.Close()
		return err
	}
	pack := packet.New(1024)

	// 发送请求
	pack.Write([]byte("Upgrade: uploader"))
	pack.Write(httpRowAt)
	// check code
	pack.Write(httpAuthorize)
	pack.Write(xutils.UnsafeStringToBytes(env.authorize.NewCode(api)))
	pack.Write(httpRowAt)
	// file size
	pack.Write(httpContentLength)
	fs, _ := fd.Stat()
	fileSize := fs.Size()
	pack.Write(xutils.ParseIntToBytes(fileSize))
	pack.Write(httpRowAt)
	pack.Write(httpRowAt)
	if _, err = pack.FlushToConn(conn); err != nil {
		packet.Free(pack)
		fd.Close()
		conn.Close()
		return errors.New("signal couldn't be sent. service may be closed")
	}

	// 接收确认信息
	err = pack.ReadConn(conn)
	if err != nil {
		packet.Free(pack)
		fd.Close()
		conn.Close()
		return errors.New("signal ok couldn't be recv. service may be closed")
	}
	if pack.ReadU32() != 1 {
		err = errors.New(pack.ReadString())
		packet.Free(pack)
		fd.Close()
		conn.Close()
		return err
	}

	// 开始发送数据
	conn.SetWriteDeadline(time.Time{})
	total := float64(fileSize)
	Logf("upload file[%s]...", api)
	_, err = pack.CopyReaderToConnWithProgress(conn, fd, fileSize, func(uploaded int64) {
		LogOrigin("uploaded...%0.2f%%", 100*float64(uploaded)/total)
	})
	fd.Close()
	LogNextLine()
	if err != nil {
		packet.Free(pack)
		conn.Close()
		return err
	}

	// 接收发送结果
	err = pack.ReadConn(conn)
	conn.Close()
	if err != nil {
		packet.Free(pack)
		return errors.New("signal cannot be received. service may be closed")
	}
	err = errors.New(pack.ReadString())
	packet.Free(pack)
	return err
}
