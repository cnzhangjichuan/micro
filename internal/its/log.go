package its

import (
	"log"
)

func NewDefaultLogger() *defaultLogger {
	return &defaultLogger{}
}

type defaultLogger struct {
}

func (s *defaultLogger) Log(f string, v ...interface{}) {
	log.Printf(f, v...)
}

func (s *defaultLogger) Error(f string, v ...interface{}) {
	log.Printf(f, v...)
}
