package iap

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/micro/xutils"
)

type Verifier struct {
	sharedSecretKey string
}

// Init 初始化实例
// sharedSecretKey 共享密钥
func (v *Verifier) Init(sharedSecretKey string) {
	v.sharedSecretKey = sharedSecretKey
}

var (
	statusPrefix      = []byte(`"status":`)
	productPrefix     = []byte(`"product_id":`)
	transactionPrefix = []byte(`"original_transaction_id":`)
)

// Verify 验证票据是否合法
func (v *Verifier) Verify(receiptData string, skipSandbox bool) (productId, transactionId string, err error) {
	var (
		url         = backendVerifyUrl
		requestData = strings.Join([]string{
			`{"receipt-data":"`, receiptData, `","password":"`, v.sharedSecretKey, `"}`,
		}, "")
		resp *http.Response
		data []byte
	)

	for i := 0; i < 2; i++ {
		resp, err = http.Post(url, `application/x-www-form-urlencoded`, bytes.NewReader([]byte(requestData)))
		if err != nil {
			return
		}
		data, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return
		}
		status, ok := v.propertyValue(data, statusPrefix)
		if !ok {
			err = errOthers
			return
		}
		switch xutils.ParseI32(status, -1) {
		default:
			err = errOthers
			return
		case respCodeSuccess:
			productId, _ = v.propertyValue(data, productPrefix)
			transactionId, _ = v.propertyValue(data, transactionPrefix)
			return
		case respCode21000:
			err = err21000
			return
		case respCode21002:
			err = err21002
			return
		case respCode21003:
			err = err21003
			return
		case respCode21004:
			err = err21004
			return
		case respCode21005:
			err = err21005
			return
		case respCode21006:
			err = err21006
			return
		case respCode21007:
			if skipSandbox {
				err = err21007
				return
			}
			url = sandboxVerifyUrl
		case respCode21008:
			err = err21008
			return
		case respCode21010:
			err = err21010
			return
		}
	}
	return
}

// propertyValue 获取对应的JSON值
func (v *Verifier) propertyValue(data, prefix []byte) (s string, ok bool) {
	i := bytes.Index(data, prefix)
	if i < 0 {
		return
	}
	data = data[i+len(prefix):]
	i = bytes.IndexByte(data, ',')
	if i < 0 {
		i = bytes.IndexByte(data, '}')
	}
	if i < 0 {
		i = len(data)
	}
	s, ok = string(bytes.Trim(bytes.TrimSpace(data[:i]), `"`)), true
	return
}
