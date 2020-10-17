package packet

import (
	"io/ioutil"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigFastest

// EncodeJSON 序列化对象为json格式的数据
func (p *Packet) EncodeJSON(v interface{}, isGzip, prefix bool) (bool, error) {
	return p.EncodeJSONWithAPI(v, isGzip, prefix, nil)
}

// EncodeJSONWithAPI 序列化对象为json格式的数据
func (p *Packet) EncodeJSONWithAPI(v interface{}, isGzip, prefix bool, api []byte) (ok bool, err error) {
	if prefix {
		p.w++
	}
	w := p.w
	if len(api) > 0 {
		p.Write(api)
	}
	stream := json.BorrowStream(p)
	stream.WriteVal(v)
	if stream.Error != nil {
		json.ReturnStream(stream)
		err = stream.Error
		return
	}
	stream.Flush()
	json.ReturnStream(stream)

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

// DecodeJSON 反序列化json数据
func (p *Packet) DecodeJSON(v interface{}) error {
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
