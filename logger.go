package micro

import (
	"fmt"
	"io"
	"log"
	"strings"
)

// SetLogger 设置日志输出源
func SetLogger(w io.Writer) {
	if env.log == nil {
		env.log = log.New(w, "[micro]", log.Ltime)
	} else {
		env.log.SetOutput(w)
	}
}

// Debug 调试
func Debug(fmt string, args ...interface{}) {
	if !strings.HasSuffix(fmt, "\n") {
		fmt += "\n"
	}
	env.log.Printf(fmt, args...)
}

// Logf 记录日志
func Logf(fmt string, args ...interface{}) {
	if !strings.HasSuffix(fmt, "\n") {
		fmt += "\n"
	}
	env.log.Printf(fmt, args...)
}

// Log 记录日志
func Log(s ...interface{}) {
	env.log.Println(s...)
}

// LogOrigin 在原位置输出内容
func LogOrigin(f string, args ...interface{}) {
	w := env.log.Writer()
	if w == nil {
		return
	}
	w.Write([]byte(fmt.Sprintf("\r"+f, args...)))
}

// LogNextLine 定位到下一行输出
func LogNextLine() {
	w := env.log.Writer()
	if w == nil {
		return
	}
	w.Write([]byte{'\n'})
}

// Errorf 错误
func Errorf(fmt string, args ...interface{}) {
	if !strings.HasSuffix(fmt, "\n") {
		fmt += "\n"
	}
	env.log.Printf(fmt, args...)
}

// Error 错误
func Error(args ...interface{}) {
	env.log.Println(args...)
}
