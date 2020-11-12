package store

import (
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/micro/packet"
)

// NewSaver 创建加载器
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
	r := env.db.QueryRow(strings.Join([]string{
		`select value from `, s.name, suffix, ` where id='`, Ignore(id), `'`,
	}, ""))
	env.RUnlock()

	var buf []byte
	if err := r.Scan(&buf); err != nil {
		return false
	}

	pack := packet.NewWithData(buf)
	err := pack.DecodeJSON(data)
	packet.Free(pack)

	return err == nil
}

// NewSingleSaver 创建单表加载器
func NewSingleSaver(name string) *singleSaver {
	if name == "" {
		return nil
	}

	// 创建数据表
	env.db.Exec(strings.Join([]string{
		`create table `, name,
		`(id varchar(20) unique,value text)`,
	}, ""))

	return &singleSaver{
		insert: strings.Join([]string{
			`insert into `, name,
			` values($1,$2)`,
			` on conflict(id) do update set value=$3`,
		}, ""),
		query: strings.Join([]string{
			`select value from `, name, ` where id=$1`,
		}, ""),
	}
}

type singleSaver struct {
	insert string
	query  string
}

// Save 保存数据
func (s *singleSaver) Save(id string, data []byte) (ok bool) {
	if s == nil {
		return
	}
	env.RLock()
	if env.db == nil {
		env.RUnlock()
		return
	}

	v := hex.EncodeToString(data)
	_, err := env.db.Exec(s.insert, id, v, v)
	env.RUnlock()
	if err != nil {
		if env.backupOnError != nil {
			env.backupOnError(s.insert, err)
		}
	} else {
		ok = true
	}
	return
}

// Find 从表数据库中查询数据
func (s *singleSaver) Find(id string) (data []byte, ok bool) {
	if s == nil {
		return
	}
	env.RLock()
	if env.db == nil {
		env.RUnlock()
		return
	}
	r := env.db.QueryRow(s.query, id)
	env.RUnlock()

	err := r.Scan(&data)
	ok = err == nil
	if ok {
		n, _ := hex.Decode(data, data)
		data = data[:n]
	}
	return
}
