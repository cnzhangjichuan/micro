package types

type Logger interface {
	Log(fmt string, v ...interface{})
	Error(fmt string, v ...interface{})
}
