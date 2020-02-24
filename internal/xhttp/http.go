package xhttp

import (
	"bytes"
	"github.com/cnzhangjichuan/micro/internal/its"
	"github.com/cnzhangjichuan/micro/xutils"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// handle http request
func Handle(w http.ResponseWriter, r *http.Request) {
	const (
		API = `/api`
		ES  = `/event-source/`
	)

	path := r.URL.Path

	switch {
	default:
		handleResource(w, r, path)

	case strings.HasPrefix(path, API):
		handleApi(w, r, path[len(API):])

	case strings.HasPrefix(path, ES):
		handleEventSource(w, path[len(ES):])

	}
}

// process event-source interface.
func handleEventSource(w http.ResponseWriter, uid string) {
	const (
		TOM    = time.Second * 5
		TICKER = time.Minute
	)

	// user
	if uid == "" {
		logError("not found logon user")
		return
	}

	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		logError("event-source hijacking error: %v", err)
		return
	}
	defer conn.Close()

	// handshake
	conn.SetWriteDeadline(time.Now().Add(TOM))
	if _, err = conn.Write(env.esp.handshake); err != nil {
		logError("event-source handshake error: %v", err)
		return
	}

	// retry settings
	if _, err = conn.Write(env.esp.retry); err != nil {
		logError("event-source set retry error: %v", err)
		return
	}

	// push data
	var (
		msgChan   = make(chan interface{}, 10)
		timer     = time.NewTicker(TICKER)
		packet    []byte
		isClosing = false
	)

	env.esp.Set(uid, msgChan)
	for {
		packet = nil
		select {
		case <-timer.C:
			packet = env.esp.heartbeat
		case msg := <-msgChan:
			if msg != nil {
				packet, err = xutils.MarshalJson(msg)
				if err == nil {
					isClosing = bytes.Equal(packet, env.esp.closer)
					if !isClosing {
						packet = bytes.Join([][]byte{env.esp.header, packet, env.esp.footer}, nil)
					}
				}
			}
		}
		if packet == nil || err != nil {
			break
		}

		conn.SetWriteDeadline(time.Now().Add(TOM))
		// close
		if isClosing {
			conn.Write(packet)
			break
		}

		// data
		if _, err = conn.Write(packet); err != nil {
			logError("push data error %v", err)
			break
		}
	}

	// release resource
	timer.Stop()
	env.esp.Del(uid, msgChan)
}

// process static file.
func handleResource(w http.ResponseWriter, r *http.Request, path string) {
	const (
		SHARE    = `/share`
		APPCACHE = `.appcache`
	)

	if path == "" || path == "/" {
		r.URL.Path = "/app.html"
	}
	switch {
	case strings.HasSuffix(path, APPCACHE):
		w.Header().Set("Content-Type", "text/cache-manifest")
	case strings.HasPrefix(path, SHARE):
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	env.fs.ServeHTTP(w, r)
}

// process api interface.
func handleApi(w http.ResponseWriter, r *http.Request, path string) {
	handler, dpo, err := findHandler(r, path)

	if err != nil {
		w.Header().Set(`Content-Type`, `application/json; charset=utf-8`)
		w.Write(xutils.UnsafeStringToBytes(`{"Error":"` + err.Error() + `"}`))
		return
	}

	// execute service
	err = handler.Func(dpo)

	// set response data type
	if dpo.df != "" {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment;filename="+dpo.df)
		if err != nil {
			logError("download error %v", err)
			return
		}
		bw := &downloadWriter{
			buf: make([]byte, 0, 10240),
		}
		excelHelper := xutils.ExcelHelper{
			Writer: bw,
		}
		excelHelper.Write(dpo.resp)
		w.Header().Set("Content-Length", strconv.Itoa(bw.Len()))
		bw.Flush(w)
	} else {
		w.Header().Set(`Content-Type`, `application/json; charset=utf-8`)
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
}

func findHandler(r *http.Request, path string) (handler *its.Handler, dpo *httpDpo, err error) {
	const SessionId = `Session`

	var ok bool

	// find handler
	handler, ok = env.handlers[path]

	if !ok {
		err = xutils.NewError(`Api not found!`)
		return
	}

	// get userdata
	sid := r.Header.Get(SessionId)
	user, ok := env.usp.Get(sid)
	if handler.Permit != "" {
		if !ok {
			// not login
			// set response data type
			err = xutils.NewError(`Need Login`)
			return
		} else if !user.Access(handler.Permit) {
			// permit not matchable
			// set response data type
			err = xutils.NewError(`Permission denied!`)
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

	dpo = &httpDpo{
		r:       r,
		reqData: unescape(data),
		user:    user,
		resp:    nil,
		df:      "",
	}

	return
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

type downloadWriter struct {
	buf []byte
}

func (d *downloadWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	d.buf = append(d.buf, p...)
	return
}

func (d *downloadWriter) Len() int {
	return len(d.buf)
}

func (d *downloadWriter) Flush(w io.Writer) {
	offset := 0
	l := d.Len()
	for offset < l {
		n, err := w.Write(d.buf[offset:])
		if err != nil {
			break
		}
		offset += n
	}
}
