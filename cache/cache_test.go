package cache

import (
	"fmt"
	"testing"
)

func TestCacheAll(t *testing.T) {
	var c Cache
	c.Init(1024)

	c.Set("xxxxxxxxxxxxxxx", []byte(`I am in ShangHai, i want go home.`), 0)
	c.Set("cccccccccccccc", []byte(`I am in ShangHai, i study in shangwai.`), 0)

	data, ok := c.Get("xxxxxxxxxxxxxxx")
	fmt.Println(ok, string(data))
}
