package store

import (
	"database/sql"
	"strconv"
	"strings"
	"sync"

	// PG数据库驱动
	_ "github.com/lib/pq"
	"github.com/micro/xutils"
)

const (
	tableCount = 16
	queenSize  = tableCount
	queenCap   = 128
)

// funcBackupOnError 错误处理函数
type funcBackupOnError func(string, error)

// env 系统配置
var env struct {
	sync.RWMutex
	sync.WaitGroup

	db            *sql.DB
	tasks         []chan string
	backupOnError funcBackupOnError
}

// IsBackupOnErrorSetted 是否设置过存储失败回调
func IsBackupOnErrorSetted() bool {
	return env.backupOnError != nil
}

// SetBackupOnError 存储失败时，以SQL/ERROR为参数，调用这个回调
func SetBackupOnError(f funcBackupOnError) {
	env.backupOnError = f
}

// Init 打开数据库
// host=localhost port=5432 user=postgres password=postgres dbname=games_test sslmode=disable
// tabs 需要创建的表集合
// sqls 需要执行的SQL语句
func Init(resource string, tabs, sqls []string) (err error) {
	env.Lock()
	if env.db != nil {
		env.Unlock()
		return
	}
	env.db, err = sql.Open("postgres", resource)
	if err == nil {
		err = env.db.Ping()
	}
	if err != nil {
		env.db.Close()
		env.db = nil
		env.Unlock()
		return
	}

	// 创建数据表(仅表名)
	for _, tab := range tabs {
		for i := 0; i < tableCount; i++ {
			env.db.Exec(strings.Join([]string{
				`create table `, tab, strconv.Itoa(i),
				`(id varchar(20) unique,value text)`,
			}, ""))
		}
	}

	// 执行SQLs
	for _, s := range sqls {
		_, err := env.db.Exec(s)
		if err != nil && env.backupOnError != nil {
			env.backupOnError(s, err)
		}
	}

	// 启动
	env.tasks = make([]chan string, queenSize)
	for i := 0; i < queenSize; i++ {
		env.tasks[i] = make(chan string, queenCap)
		go func(db *sql.DB, q <-chan string) {
			env.Add(1)
			for SQL := range q {
				if _, err := db.Exec(SQL); err != nil {
					if env.backupOnError != nil {
						env.backupOnError(SQL, err)
					}
				}
			}
			env.Done()
		}(env.db, env.tasks[i])
	}

	env.Unlock()
	return
}

// Close 关闭连接
func Close() {
	// 关闭任务
	env.Lock()
	if env.tasks != nil {
		for i := 0; i < len(env.tasks); i++ {
			close(env.tasks[i])
		}
		env.tasks = nil
	}
	env.Unlock()

	// 等待所有任务完成
	env.Wait()

	// 关闭数据库连接
	env.Lock()
	if env.db != nil {
		env.db.Close()
		env.db = nil
	}
	env.Unlock()
}

// addSQL 添加SQL到队列中
func addSQLToQueen(SQL string) {
	if env.tasks == nil {
		return
	}
	select {
	case env.tasks[0] <- SQL:
	case env.tasks[1] <- SQL:
	case env.tasks[2] <- SQL:
	case env.tasks[3] <- SQL:
	case env.tasks[4] <- SQL:
	case env.tasks[5] <- SQL:
	case env.tasks[6] <- SQL:
	case env.tasks[7] <- SQL:
	case env.tasks[8] <- SQL:
	case env.tasks[9] <- SQL:
	case env.tasks[10] <- SQL:
	case env.tasks[11] <- SQL:
	case env.tasks[12] <- SQL:
	case env.tasks[13] <- SQL:
	case env.tasks[14] <- SQL:
	case env.tasks[15] <- SQL:
	}
}

// tableSuffix 获取表名后辍
func tableSuffix(id string) (suffix string) {
	idx := xutils.HashCode32(id) % tableCount
	suffix = strconv.Itoa(int(idx))
	return
}
