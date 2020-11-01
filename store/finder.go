package store

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
)

var errNotInitialized = errors.New("store not initialized")

// Query 通过SQL查询数据
func Query(load func(*sql.Rows) error, SQL string, args ...interface{}) error {
	env.RLock()
	if env.db == nil {
		env.RUnlock()
		return errNotInitialized
	}
	r, err := env.db.Query(SQL, args...)
	env.RUnlock()

	if err != nil {
		return err
	}
	defer r.Close()
	for r.Next() {
		if err := load(r); err != nil {
			return err
		}
	}
	return nil
}

// Execute 执行SQL
func Execute(SQL string) {
	env.RLock()
	addSQLToQueen(SQL)
	env.RUnlock()
}

// ExecuteNow 立即执行SQL
func ExecuteNow(SQL string) error {
	env.RLock()
	if env.db == nil {
		env.RUnlock()
		return errNotInitialized
	}
	_, err := env.db.Exec(SQL)
	env.RUnlock()
	if err != nil && env.backupOnError != nil {
		env.backupOnError(SQL, err)
	}
	return err
}

// Ignore 一定程度防止SQL注入
func Ignore(s string) string {
	return strings.Replace(s, "'", "", -1)
}

// 置换SQL参数
func SetParam(sq, name string, value interface{}) string {
	switch v := value.(type) {
	case string:
		return strings.Replace(sq, name, Ignore(v), 1)
	case int:
		return strings.Replace(sq, name, strconv.Itoa(v), 1)
	case int8:
		return strings.Replace(sq, name, strconv.Itoa(int(v)), 1)
	case int16:
		return strings.Replace(sq, name, strconv.Itoa(int(v)), 1)
	case int32:
		return strings.Replace(sq, name, strconv.Itoa(int(v)), 1)
	case int64:
		return strings.Replace(sq, name, strconv.Itoa(int(v)), 1)
	case float32:
		return strings.Replace(sq, name, strconv.FormatFloat(float64(v), 'f', -1, 64), 1)
	case float64:
		return strings.Replace(sq, name, strconv.FormatFloat(v, 'f', -1, 64), 1)
	case uint:
		return strings.Replace(sq, name, strconv.FormatUint(uint64(v), 10), 1)
	case uint8:
		return strings.Replace(sq, name, strconv.FormatUint(uint64(v), 10), 1)
	case uint32:
		return strings.Replace(sq, name, strconv.FormatUint(uint64(v), 10), 1)
	case uint64:
		return strings.Replace(sq, name, strconv.FormatUint(v, 10), 1)
	case bool:
		if v {
			return strings.Replace(sq, name, "1", 1)
		}
		return strings.Replace(sq, name, "0", 1)
	}
	return strings.Replace(sq, name, "", 1)
}
