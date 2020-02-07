package xutils

import (
	"errors"
	"fmt"
)

func NewError(e interface{}) error {
	return errors.New(fmt.Sprintf("%v", e))
}
