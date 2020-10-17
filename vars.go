package micro

import "errors"

type errBisResp struct {
	ErrCode string
}

var (
	// apiNotFoundError 没有找到业务接口
	apiNotFoundError = &errBisResp{
		ErrCode: "ApiNotFound",
	}

	// notLoginError 没有登录
	noLoginError = &errBisResp{
		ErrCode: "NoLogin",
	}

	// errRPCNotFoundService RPC调用没有发现服务
	errRPCNotFoundService = errors.New("rpc: not found service")

	// errWSHDError WebSocket无效的Token
	errWSInvalidToken = errors.New(`ws: token invalid`)

	// errUploadError 文件上传错误
	errUploadError = errors.New("update: upload file painc")
)

var (
	httpUpgrade          = []byte("Upgrade: ")
	httpRPCUpgrade       = []byte("Upgrade: rpc")
	httpRegistryUpgrade  = []byte("Upgrade: registry")
	httpAuthorize        = []byte("Authorize: ")
	httpRemoteAddress    = []byte("Remote-Addr: ")
	httpRegistryPort     = []byte("ServerPort: ")
	httpPathStart        = []byte{' ', '/'}
	httpPathEnd          = []byte{' '}
	httpContentLength    = []byte("Content-Length: ")
	httpConnection       = []byte("Connection: ")
	httpKeepAlive        = []byte("keep-alive")
	httpAPI              = []byte("Api: ")
	httpRowAt            = []byte{'\r', '\n'}
	httpAcceptEncoding   = []byte("Accept-Encoding: ")
	httpAcceptGzlib      = []byte("zlib")
	httpUID              = []byte("UID: ")
	httpRanges           = []byte("Range: ")
	httpRespOk           = []byte("HTTP/1.1 200 OK\r\nConnection: keep-alive\r\n")
	httpRespOkAccess     = []byte("HTTP/1.1 200 OK\r\nConnection: keep-alive\r\nAccess-Control-Allow-Origin: *\r\nAccess-Control-Expose-Headers: Api\r\n")
	httpRespOk206        = []byte("HTTP/1.1 206 OK\r\nConnection: keep-alive\r\n")
	httpRespOk404        = []byte("HTTP/1.1 404 Not Found\r\nConnection: keep-alive\r\n")
	httpRespAccpetRanges = []byte("Accept-Ranges: bytes\r\n")
	httpRespRanges       = []byte("Content-Range: bytes ")
	httpRespEncoding     = []byte("Content-Encoding: zlib\r\n")
	httpRespContent0     = []byte("Content-Length: 0\r\n")
	httpRespStream       = []byte("Content-Type: application/octet-stream\r\n")
	httpRespJSON         = []byte("Content-Type: application/json; charset=utf-8\r\n")
	httpRespHTML         = []byte("Content-Type: text/html; charset=utf-8\r\n")
	httpRespCSS          = []byte("Content-Type: text/css; charset=utf-8\r\n")
	httpRespPNG          = []byte("Content-Type: image/png\r\n")
	httpRespJPG          = []byte("Content-Type: image/jpeg\r\n")
	httpRespGIF          = []byte("Content-Type: image/gif\r\n")
	httpRespICO          = []byte("Content-Type: image/x-icon\r\n")
	httpRespAppCache     = []byte("Content-Type: text/cache-manifest\r\n")
	httpRespErrorPrefix  = []byte(`{"ErrCode":"`)
	httpRespErrorSuffix  = []byte(`"}`)
	wsRespOK             = []byte("HTTP/1.1 101 Web Socket Protocol Handshake\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n")
	wsKey                = []byte(`Sec-WebSocket-Key: `)
	wsAccept             = []byte("Sec-WebSocket-Accept: ")
	wsProtocol           = []byte("Sec-WebSocket-Protocol: ")
	wsPing               = []byte{0x89, 0x00}
	wsPong               = []byte{0x8a, 0x00}
)
