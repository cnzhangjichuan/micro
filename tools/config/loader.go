package config

import (
	"errors"
	"reflect"

	"github.com/micro/packet"
)

type OnSetupFunc func(int) interface{}

// Load 加载数据
func (e *excel) Load(name string, onSetup OnSetupFunc) error {
	e.RLock()
	defer e.RUnlock()

	pack, err := packet.NewWithFile(e.path)
	if err != nil {
		return errors.New(`excel: ` + err.Error())
	}
	for {
		n := pack.ReadString()
		if n == "" {
			break
		}
		s := pack.ReadI64()
		if n != name {
			pack.Skip(int(s))
			continue
		}
		names := pack.ReadStrings()
		size := int(pack.ReadI64())
		arrayValue := reflect.ValueOf(onSetup(size))
		for i := 0; i < size; i++ {
			value := arrayValue.Index(i)
			for _, n := range names {
				f := value.FieldByName(n)
				switch f.Kind() {
				case reflect.String:
					f.SetString(pack.ReadString())
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					f.SetInt(pack.ReadI64())
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					f.SetUint(pack.ReadU64())
				case reflect.Float32:
					f.SetFloat(float64(pack.ReadF32()))
				case reflect.Float64:
					f.SetFloat(pack.ReadF64())
				case reflect.Struct:
					m := f.Addr().MethodByName("Parse")
					if m.IsValid() {
						m.Call([]reflect.Value{reflect.ValueOf(pack.ReadString())})
					} else {
						pack.ReadString()
					}
				case reflect.Ptr:
					f.Set(reflect.New(f.Type().Elem()))
					m := f.MethodByName("Parse")
					if m.IsValid() {
						m.Call([]reflect.Value{reflect.ValueOf(pack.ReadString())})
					} else {
						pack.ReadString()
					}
				}
			}
		}
		packet.Free(pack)
		return nil
	}

	// not found data
	packet.Free(pack)
	return errors.New(`excel: not found ` + name)
}
