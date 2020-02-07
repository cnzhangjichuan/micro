package xhttp

import (
	"github.com/cnzhangjichuan/micro/internal/its"
	"github.com/cnzhangjichuan/micro/types"
	"github.com/cnzhangjichuan/micro/xutils"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type httpDpo struct {
	its.BaseDpo

	r       *http.Request
	reqData []byte
	user    types.User
	resp    interface{}
}

func (h *httpDpo) Request(v interface{}) error {
	return xutils.UnmarshalJson(h.reqData, v)
}

func (h *httpDpo) GetUser() types.User {
	return h.user
}

func (h *httpDpo) BindUser(u types.User) {
	env.usp.Set(u)
	h.user = u
}

func (h *httpDpo) UnBindUser() {
	if h.user != nil {
		env.usp.Del(h.user.GetUserId())
		h.user = nil
	}
}

func (h *httpDpo) Response(resp interface{}) {
	h.resp = resp
}

func (h *httpDpo) DeleteFile(name string) error {
	const PREFIX = `./assets/share`

	return os.Remove(filepath.Join(PREFIX, name))
}

func (h *httpDpo) MoveFileTo(name, dstName string) (string, error) {
	const PREFIX = `./assets/share`

	f, hd, err := h.r.FormFile(name)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if dstName == "" {
		dstName = strconv.Itoa(int(time.Now().Unix()))
	}

	// add suffix
	dot := strings.LastIndex(hd.Filename, ".")
	if dot >= 0 {
		dstName = dstName + hd.Filename[dot:]
	}

	// add prefix for user
	if h.user != nil {
		dstName = h.user.GetUserId() + "/" + dstName
	}

	// create dirs
	dst := filepath.Join(PREFIX, dstName)
	os.MkdirAll(filepath.Dir(dst), os.ModePerm)

	// create new file
	fd, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	_, err = io.Copy(fd, f)
	return dstName, err
}

func (h *httpDpo) ProcessFile(name string, f func(multipart.File, *multipart.FileHeader) error) error {
	fd, hd, err := h.r.FormFile(name)
	if err != nil {
		return err
	}
	defer fd.Close()

	return f(fd, hd)
}

func (h *httpDpo) Proxy(host string) error {
	switch {
	default:
		h.r.URL.Scheme = "http"
		h.r.URL.Host = host
	case strings.HasPrefix(host, "http://"):
		h.r.URL.Scheme = "http"
		h.r.URL.Host = host[7:]
	case strings.HasPrefix(host, "https://"):
		h.r.URL.Scheme = "https"
		h.r.URL.Host = host[8:]
	}

	resp, err := http.DefaultTransport.RoundTrip(h.r)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	h.resp = data

	return nil
}
