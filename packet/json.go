package packet

import (
	"encoding/json"
	"io/ioutil"
)

// EncodeJSON 编码为json字符串
func (p *Packet) EncodeJSON(v interface{}, isGzip, prefix bool) (bool, error) {
	return p.EncodeJSONApi(v, isGzip, prefix, nil)
}

// EncodeJSONApi 编码为json字符串
func (p *Packet) EncodeJSONApi(v interface{}, isGzip, prefix bool, api []byte) (ok bool, err error) {
	if prefix {
		p.w++
	}
	w := p.w
	if len(api) > 0 {
		p.Write(api)
	}

	data, er := json.Marshal(v)
	if er != nil {
		err = er
		return
	}
	p.Write(data)

	if isGzip {
		ok = p.Compress(w)
		if prefix {
			if ok {
				p.buf[w-1] = 1
			} else {
				p.buf[w-1] = 0
			}
		}
	} else if prefix {
		p.buf[w-1] = 0
	}
	return
}

// DecodeJSON json解码
func (p *Packet) DecodeJSON(v interface{}) error {
	if p.r >= p.w {
		return nil
	}
	if p.buf[p.r] == 1 {
		p.r++
		p.UnCompress(p.r)
	} else if p.buf[p.r] == 0 {
		p.r++
	}
	return json.Unmarshal(p.Data(), v)
}

// LoadConfig 加载配置文件
func (p *Packet) LoadConfig(path string, config interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	p.Write(data)
	return p.DecodeJSON(config)
}
