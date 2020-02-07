// 使用grom+postgres
package store

import (
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

/*
	// 连接范例
	db, err = gorm.Open("mysql", "<user>:<password>/<database>?charset=utf8&parseTime=True&loc=Local")
    if err != nil {
        panic(err)
    }

	// 建表范例
	type Table struct {
		ID        int    `gorm:"primary_key"`
		Ip        string `gorm:"type:varchar(20);not null;index:ip_idx"`
		Ua        string `gorm:"type:varchar(256);not null;"`
		Title     string `gorm:"type:varchar(128);not null;index:title_idx"`
		Hash      uint64 `gorm:"unique_index:hash_idx;"`
		...
	}

	if !db.HasTable(&Table{}) {
		if err := db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").CreateTable(&Table{}).Error; err != nil {
			panic(err)
		}
	}

 */

var Gorm *gorm.DB

// source:
// 	pg, host=%s port=%s user=%s password=%s dbname=%s sslmode=disable
func Init(source string, logMode bool) {
	var err error
	Gorm, err = gorm.Open("postgres", source)
	if err != nil {
		panic(err)
	}
	Gorm.LogMode(logMode)
}

// 数据分页查询载体
type Page struct {
	Index int
	Page  int
	List  interface{}

	pageSize int
}

func (p *Page) SetIndex(index int) *Page {
	if index < 1 {
		index = 1
	}
	p.Index = index
	return p
}

func (p *Page) SetPageSize(pageSize int) *Page {
	p.pageSize = pageSize
	return p
}

type PageFinder func(db *gorm.DB) (interface{}, error)

// 以分页的形式装载数据
func (p *Page) Load(db *gorm.DB, find PageFinder) error {
	var (
		index = p.Index
		pageSize = p.pageSize
	)

	if pageSize < 1 {
		pageSize = 10
	}

	// 设置数据条目
	var count int
	if err := db.Count(&count).Error; err != nil {
		return err
	}
	p.Page = (count-1)/pageSize + 1

	// 查询数据
	offset := (index-1)*pageSize
	data, err := find(db.Offset(offset).Limit(pageSize))
	if err != nil {
		return err
	}
	p.List = data

	return nil
}