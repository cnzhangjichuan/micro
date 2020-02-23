package types

import (
	"mime/multipart"
)

// data process object
type Dpo interface {
	Request(interface{}) error
	Response(interface{})
	FileName(string)
	GetUser() User
	BindUser(User)
	UnBindUser()
	MoveFileTo(string, string) (string, error)
	DeleteFile(string) error
	ProcessFile(string, func(multipart.File, *multipart.FileHeader) error) error
	SetRoom(string)
	Close()
	SendMessage(message interface{}, user ...string)
	SendRoomMessage(interface{})
	Proxy(string) error
}
