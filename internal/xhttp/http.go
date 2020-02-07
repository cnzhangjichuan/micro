package xhttp

import (
	"github.com/cnzhangjichuan/micro/xutils"
	"io"
	"net/http"
	"strings"
)

// handle http request
func Handle(w http.ResponseWriter, r *http.Request) {
	const (
		API   = `/api`
		SHARE = `/share`
	)

	path := r.URL.Path

	switch {
	default:
		if path == "" || path == "/" {
			r.URL.Path = "/app.html"
		}
		if strings.HasPrefix(path, SHARE) {
			w.Header().Set("Content-Type", "application/octet-stream")
		}
		env.fs.ServeHTTP(w, r)

	case strings.HasPrefix(path, API):
		handle(w, r, path[len(API):])

	}
}

func handle(w http.ResponseWriter, r *http.Request, path string) {
	const SessionId = `Session`

	// set response data type
	w.Header().Set(`Content-Type`, `application/json; charset=utf-8`)

	// find handler
	handler, ok := env.handlers[path]
	if !ok {
		// not found handler
		w.Write(xutils.UnsafeStringToBytes(`{"Error":"Api is not found!"}`))
		return
	}

	// get userdata
	sid := r.Header.Get(SessionId)
	user, ok := env.usp.Get(sid)
	if handler.Permit != "" {
		if !ok {
			// not login
			w.Write(xutils.UnsafeStringToBytes(`{"Error":"Need Login"}`))
			return
		} else if !user.Access(handler.Permit) {
			// permit not matchable
			w.Write(xutils.UnsafeStringToBytes(`{"Error":"Permission denied!"}`))
			return
		}
	}

	// load dpo
	var data []byte
	switch {
	default:
		data = xutils.UnsafeStringToBytes(r.URL.RawQuery)
	case r.ContentLength > 0:
		ct := strings.TrimLeft(r.Header.Get("Content-Type"), " ")
		if ct != "" && strings.HasPrefix(ct, "multipart/form-data;") {
			data = xutils.UnsafeStringToBytes(r.FormValue(`data`))
		} else {
			data = make([]byte, r.ContentLength)
			io.ReadFull(r.Body, data)
			r.Body.Close()
		}
	}

	dpo := &httpDpo{
		r:       r,
		reqData: unescape(data),
		user:    user,
		resp:    nil,
	}

	// call handler
	err := handler.Func(dpo)

	// response
	if err != nil {
		w.Write(xutils.UnsafeStringToBytes(`{"Error":"` + strings.Replace(err.Error(), `"`, `'`, -1) + `"}`))
		return
	}
	if dpo.resp != nil {
		data, err := xutils.MarshalJson(dpo.resp)
		if err != nil {
			w.Write(xutils.UnsafeStringToBytes(`{"Error":"` + strings.Replace(err.Error(), `"`, `'`, -1) + `"}`))
		} else {
			w.Write(data)
		}
	}
}

func sax(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

// 解码地址栏数据
func unescape(s []byte) []byte {
	var (
		cx = 0
		px = 0
		sx = len(s)
	)

	for px < sx {
		switch s[px] {
		case '%':
			s[cx] = sax(s[px+1])<<4 | sax(s[px+2])
			px += 3
			cx += 1
		default:
			s[cx] = s[px]
			cx += 1
			px += 1
		}
	}
	return s[:cx]
}
