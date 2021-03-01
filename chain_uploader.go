package micro

import (
	"errors"
	"io"
	"net"
	"os"
	"runtime"
	"strings"
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
		pack.WriteString("not found api")
		pack.EndWrite()
		pack.FlushToConn(conn)
		return true
	}

	// 接收body体
	if err := pack.ReadHTTPBodyStream(conn); err != nil {
		pack.BeginWrite()
		pack.WriteString("invalid body data")
		pack.EndWrite()
		pack.FlushToConn(conn)
		return true
	}

	// 开始处理数据
	err := f(pack)
	pack.BeginWrite()
	if err != nil {
		pack.WriteString("upload file error:" + err.Error())
	} else {
		pack.WriteString(`file has been uploaded`)
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

// RequestUploadService 上传文件
func RequestUploadService(remoteAddress, mask string, apiAndFiles []string) {
	env.authorize.Init(mask)
	for i := 1; i < len(apiAndFiles); i += 2 {
		err := requestUploadService(apiAndFiles[i-1], apiAndFiles[i], remoteAddress)
		Log(err)
	}
	Log("all files uploaded!")
}

// 上传文件
func requestUploadService(api, fileName, remoteAddress string) error {
	fd, err := os.Open(fileName)
	if err != nil {
		return err
	}
	// 解析路径
	var path string
	if i := strings.Index(remoteAddress, "/"); i > 0 {
		path = remoteAddress[i:]
		remoteAddress = remoteAddress[:i]
	}
	conn, err := net.DialTimeout("tcp", remoteAddress, time.Second*3)
	if err != nil {
		fd.Close()
		return errors.New("service not found, it may be closed")
	}

	pack := packet.New(1024)

	// 发送请求
	if path != "" {
		pack.Write([]byte(`GET ` + path + ` HTTP/1.1`))
		pack.Write(httpRowAt)
		pack.Write([]byte(`Host: ` + remoteAddress))
		pack.Write(httpRowAt)
	}
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
