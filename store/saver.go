package store

import (
	"strconv"
	"strings"

	"github.com/micro/packet"
)

// NewSaver 创建回载器
func NewSaver(name string) *saver {
	if name == "" {
		return nil
	}

	// 创建数据表(仅表名)
	for i := 0; i < tableCount; i++ {
		env.db.Exec(strings.Join([]string{
			`create table `, name, strconv.Itoa(i),
			`(id varchar(20) unique,value text)`,
		}, ""))
	}

	return &saver{name: name}
}

type saver struct {
	name string
}

// Save 将数据保存到指定的数据表中
func (s *saver) Save(data interface{}, id string) {
	if s == nil {
		return
	}

	pack := packet.New(1024)
	_, err := pack.EncodeJSON(data, false, false)
	if err != nil {
		// log error
		return
	}
	tx := Ignore(string(pack.Data()))
	packet.Free(pack)

	seq := tableSuffix(id)
	id = Ignore(id)

	env.RLock()
	addSQLToQueen(strings.Join([]string{
		`insert into `, s.name, seq,
		` values('`, id, `','`, tx, `')`,
		` on conflict(id) do update set value='`, tx, `'`,
	}, ""))
	env.RUnlock()

	return
}

// Load 从数据表中加载数据
func (s *saver) Load(data interface{}, id string) bool {
	if s == nil {
		return false
	}

	env.RLock()
	if env.db == nil {
		env.RUnlock()
		return false
	}
	suffix := tableSuffix(id)
	r, err := env.db.Query(strings.Join([]string{
		`select value from `, s.name, suffix, ` where id='`, Ignore(id), `'`,
	}, ""))
	env.RUnlock()

	if err != nil {
		return false
	}
	if !r.Next() {
		r.Close()
		return false
	}

	var buf []byte
	r.Scan(&buf)
	r.Close()

	pack := packet.NewWithData(buf)
	err = pack.DecodeJSON(data)
	packet.Free(pack)

	return err == nil
}
