package iap

import (
	"bytes"
	"github.com/micro/xutils"
	"io/ioutil"
	"net/http"
	"strings"
)

type Verifier struct {
	sharedSecretKey string
}

// New 创建验证实例
// sharedSecretKey 共享密钥
func New(sharedSecretKey string) *Verifier {
	return &Verifier{
		sharedSecretKey: sharedSecretKey,
	}
}

var (
	statusPrefix = []byte(`"status":`)
)

// 验证票据是否合法
func (v *Verifier) Verify(receiptData string, skipSandbox bool) error {
	var (
		url         = backendVerifyUrl
		requestData = strings.Join([]string {
			`{"receipt-data":"`, receiptData, `","password":"`, v.sharedSecretKey, `"}`,
		}, "")
	)

	for i := 0; i < 2; i++ {
		resp, err := http.Post(url, `application/x-www-form-urlencoded`, bytes.NewReader([]byte(requestData)))
		if err != nil {
			return err
		}
		data, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}
		i := bytes.Index(data, statusPrefix)
		if i < 0 {
			return errOthers
		}
		data = data[i+len(statusPrefix):]
		i = bytes.IndexByte(data, ',')
		if i < 0 {
			i = bytes.IndexByte(data, '}')
		}
		code := xutils.ParseI32(string(data[:i]), -1)
		switch code {
		default:
			return errOthers
		case respCodeSuccess:
			return nil
		case respCode21000:
			return err21000
		case respCode21002:
			return err21002
		case respCode21003:
			return err21003
		case respCode21004:
			return err21004
		case respCode21005:
			return err21005
		case respCode21006:
			return err21006
		case respCode21007:
			if skipSandbox {
				return err21007
			}
			url = sandboxVerifyUrl
		case respCode21008:
			return err21008
		case respCode21010:
			return err21010
		}
	}
	return errOthers
}
