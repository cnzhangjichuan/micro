package micro

import (
	"bytes"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/micro/packet"
	"github.com/micro/xutils"
)

type http struct {
	baseChain

	asm     sync.RWMutex
	assets  map[string][]byte
	dpoPool sync.Pool
}

// Init 初始化
func (h *http) Init() {
	h.assets = make(map[string][]byte, 64)
	h.dpoPool.New = func() interface{} {
		return &httpDpo{}
	}
}

// Reload 重新加载资源
func (h *http) Reload() {
	h.asm.Lock()
	for k := range h.assets {
		delete(h.assets, k)
	}
	h.asm.Unlock()
}

const assets = `./assets`

// Handle 处理Conn
func (h *http) Handle(conn net.Conn, name string, pack *packet.Packet) bool {
	const (
		WT = time.Second * 10
		RT = time.Second * 30
	)

	if name != "" {
		return false
	}

	// 设置超时时长
	pack.SetTimeout(RT, WT)
	var (
		cac    dpoCache
		remote = ""
	)

	// 获取远端地址
	remote = pack.HTTPHeaderValue(httpRemoteAddress)
	if remote == "" {
		remote = conn.RemoteAddr().String()
	}

	for {
		// 获取资源路径
		path := string(pack.DataBetween(httpPathStart, httpPathEnd))

		// api
		api := pack.HTTPHeaderValue(httpAPI)

		// 是否支持zlib
		v := pack.DataBetween(httpAcceptEncoding, httpRowAt)
		isZlib := bytes.Index(v, httpAcceptGzlib) >= 0

		// 是否长连
		v = pack.DataBetween(httpConnection, httpRowAt)
		isKeepAlive := bytes.Index(v, httpKeepAlive) >= 0

		// 处理资源
		if api == "" || path != "" {
			if err := h.sendResource(conn, pack, path, isZlib); err != nil {
				break
			}
		} else {
			if cac == nil {
				cac = createDpoCache()
			}
			if err := h.callAPI(conn, pack, api, remote, isZlib, cac); err != nil {
				break
			}
		}

		// 如果不是长连，直接断开连结
		if !isKeepAlive {
			break
		}

		// 读入下一个报头
		if err := pack.ReadHTTPHeader(conn); err != nil {
			break
		}
	}
	freeDpoCache(cac)
	return true
}

// callAPI 调用api
func (h *http) callAPI(conn net.Conn, pack *packet.Packet, api, remote string, isZlib bool, cac map[string]interface{}) error {
	var (
		resp    interface{}
		errCode string
	)

	// 用户标识
	uid := pack.HTTPHeaderValue(httpUID)
	if uid != "" && !env.cache.Has(uid) {
		uid = ""
	}

	// 读取消息体
	pack.ReadHTTPBody(conn)

	// 调用业务接口
	if !env.authorize.CheckAPI(uid, api) {
		errCode = noLoginError.ErrCode
	} else {
		if bis, ok := findBis(api); !ok {
			errCode = apiNotFoundError.ErrCode
		} else {
			dpo := h.createDpo()
			dpo.uid = uid
			dpo.pack = pack
			dpo.cache = cac
			dpo.SetRemote(remote)
			resp, errCode = bis(dpo)
			h.freeDpo(dpo)
		}
	}

	pack.Reset()
	pack.Write(httpRespOkAccess)
	pack.Write(httpRespJSON)
	pack.Write(httpAPI)
	pack.Write(xutils.UnsafeStringToBytes(api))
	pack.Write(httpRowAt)

	// 业务错误
	if errCode != "" {
		s := pack.Size()
		pack.Write(httpRespErrorPrefix)
		pack.Write(xutils.UnsafeStringToBytes(errCode))
		pack.Write(httpRespErrorSuffix)
		e := pack.Size()
		pack.Write(httpContentLength)
		pack.Write(xutils.ParseIntToBytes(int64(e - s)))
		pack.Write(httpRowAt)
		pack.Write(httpRowAt)
		pack.MoveToEnd(s, e)
		_, err := pack.FlushToConn(conn)
		return err
	}

	// 无响应数据
	if resp == nil {
		pack.Write(httpRespContent0)
		pack.Write(httpRowAt)
		_, err := pack.FlushToConn(conn)
		return err
	}

	// 返回业务数据
	s := pack.Size()
	ok, _ := pack.EncodeJSON(resp, isZlib, false)
	e := pack.Size()
	pack.Write(httpContentLength)
	pack.Write(xutils.ParseIntToBytes(int64(e - s)))
	pack.Write(httpRowAt)
	if ok {
		pack.Write(httpRespEncoding)
	}
	pack.Write(httpRowAt)
	pack.MoveToEnd(s, e)
	_, err := pack.FlushToConn(conn)
	return err
}

// sendResource 发送静态资源
func (h *http) sendResource(conn net.Conn, pack *packet.Packet, path string, isZlib bool) error {
	if path == "" {
		path = "index.html"
	}

	// 读取Range头
	rgs := pack.HTTPHeaderValue(httpRanges)

	// 读取消息体
	pack.ReadHTTPBody(conn)
	pack.Reset()

	// 加载静态资源(大数据)
	if !env.config.AssetsCache || strings.HasPrefix(path, "resource") {
		fd, err := os.Open(filepath.Join(assets, path))
		if err != nil {
			// 404
			pack.Write(httpRespOk404)
			// 文档类型
			h.setContentType(pack, path)
			pack.Write(httpRespAccpetRanges)
			pack.Write(httpRespContent0)
			pack.Write(httpRowAt)
			_, err := pack.FlushToConn(conn)
			return err
		}

		fst, err := fd.Stat()
		if err != nil {
			fd.Close()
			// 404
			pack.Write(httpRespOk404)
			// 文档类型
			h.setContentType(pack, path)
			pack.Write(httpRespAccpetRanges)
			pack.Write(httpRespContent0)
			pack.Write(httpRowAt)
			_, err := pack.FlushToConn(conn)
			return err
		}
		fSize := fst.Size()
		from, to := h.parseRanges(rgs, fSize)
		if to > 0 {
			fd.Seek(from, 0)
			// 206
			pack.Write(httpRespOk206)
			h.setContentType(pack, path)
			pack.Write(httpRespRanges)
			pack.Write(xutils.ParseIntToBytes(from))
			pack.WriteByte('-')
			pack.Write(xutils.ParseIntToBytes(to))
			pack.WriteByte('/')
			pack.Write(xutils.ParseIntToBytes(fSize))
			pack.Write(httpRowAt)
			fSize = to - from + 1
		} else {
			// 200
			pack.Write(httpRespOk)
			h.setContentType(pack, path)
			pack.Write(httpRespAccpetRanges)
		}
		pack.Write(httpContentLength)
		pack.Write(xutils.ParseIntToBytes(fSize))
		pack.Write(httpRowAt)
		pack.Write(httpRowAt)
		if _, err := pack.FlushToConn(conn); err != nil {
			fd.Close()
			return err
		}
		_, err = pack.CopyReaderToConn(conn, fd, to+1)
		fd.Close()
		return err
	}

	// 加载静态资源(小数据)
	pack.Write(httpRespOk)
	isZlib = h.setContentType(pack, path) && isZlib
	data, isZlib := h.loadResource(path, isZlib)
	if isZlib {
		pack.Write(httpRespEncoding)
	}
	pack.Write(httpContentLength)
	pack.Write(xutils.ParseIntToBytes(int64(len(data))))
	pack.Write(httpRowAt)
	pack.Write(httpRowAt)
	pack.Write(data)

	_, err := pack.FlushToConn(conn)
	return err
}

// parseRanges 处理Ranges数据
// bytes=
func (h *http) parseRanges(rgs string, size int64) (from, to int64) {
	if len(rgs) < 6 || rgs[:6] != "bytes=" {
		from, to = 0, -1
		return
	}
	rgs = rgs[6:]
	i := strings.IndexByte(rgs, '-')
	if i == 0 {
		to = xutils.ParseI64(rgs, 0)
		from, to = size+to, size-1
	} else if i < 0 {
		from = xutils.ParseI64(rgs, 0)
		to = size - 1
	} else {
		from = xutils.ParseI64(rgs[:i], 0)
		to = xutils.ParseI64(rgs[i+1:], size-1)
	}

	if from > to || to >= size {
		from, to = 0, -1
	}
	return
}

// loadResource 加载静态资源
func (h *http) loadResource(path string, isZlib bool) ([]byte, bool) {
	var (
		key      = path
		nZlibKey = ""
	)
	if isZlib {
		key = key + ".zlb"
	}

	// load from cache
	h.asm.RLock()
	if data, ok := h.assets[key]; ok {
		h.asm.RUnlock()
		return data, isZlib
	}
	if isZlib {
		nZlibKey = path + ".nzlb"
		if data, ok := h.assets[nZlibKey]; ok {
			h.asm.RUnlock()
			return data, false
		}
	}
	h.asm.RUnlock()

	// load from disk
	h.asm.Lock()
	if data, ok := h.assets[key]; ok {
		h.asm.Unlock()
		return data, isZlib
	}
	if isZlib {
		if data, ok := h.assets[nZlibKey]; ok {
			h.asm.Unlock()
			return data, false
		}
	}
	data, err := ioutil.ReadFile(filepath.Join(assets, path))
	if err != nil {
		h.asm.Unlock()
		return nil, false
	}
	if isZlib {
		pack := packet.NewWithData(data)
		isZlib = pack.Compress(0)
		if !isZlib {
			copy(data, pack.Data())
			key = nZlibKey
			h.assets[path] = data
		}
		packet.Free(pack)
	}
	h.assets[key] = data
	h.asm.Unlock()

	return data, isZlib
}

// 设置文档类型
func (h *http) setContentType(pack *packet.Packet, path string) bool {
	switch {
	default:
		pack.Write(httpRespStream)
		return false
	case strings.HasSuffix(path, ".json"):
		pack.Write(httpRespJSON)
		return true
	case strings.HasSuffix(path, ".html"):
		pack.Write(httpRespHTML)
		return true
	case strings.HasSuffix(path, ".css"):
		pack.Write(httpRespCSS)
		return true
	case strings.HasSuffix(path, ".png"):
		pack.Write(httpRespPNG)
		return false
	case strings.HasSuffix(path, ".jpg"):
		pack.Write(httpRespJPG)
		return false
	case strings.HasSuffix(path, ".gif"):
		pack.Write(httpRespGIF)
		return false
	case strings.HasSuffix(path, ".ico"):
		pack.Write(httpRespICO)
		return false
	case strings.HasSuffix(path, ".appcache"):
		pack.Write(httpRespAppCache)
		return true
	}
}

// createDpo 创建处理对象
func (h *http) createDpo() *httpDpo {
	return h.dpoPool.Get().(*httpDpo)
}

// freeDpo 释放处理对象
func (h *http) freeDpo(dpo *httpDpo) {
	if dpo == nil {
		return
	}
	dpo.pack = nil
	dpo.release()
	h.dpoPool.Put(dpo)
}

type httpDpo struct {
	baseDpo

	pack *packet.Packet
}

// Parse 获取客户端参数
func (h *httpDpo) Parse(v interface{}) {
	if h.pack == nil {
		return
	}
	err := h.pack.DecodeJSON(v)
	if err != nil {
		Debug("http dpo parse data error: %v", err)
	}
}
