package its

import "fmt"

func NewDefaultLogger() *defaultLogger {
	return &defaultLogger{}
}

type defaultLogger struct {
}

func (s *defaultLogger) Log(f string, v ...interface{}) {
	fmt.Printf(f+"\n", v...)
}

func (s *defaultLogger) Error(f string, v ...interface{}) {
	fmt.Printf(f+"\n", v...)
}
