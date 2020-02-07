package its

import (
	"github.com/cnzhangjichuan/micro/types"
	"mime/multipart"
)

type BaseDpo struct{}

func (b *BaseDpo) Request(v interface{}) error                     { return nil }
func (b *BaseDpo) Response(resp interface{})                       {}
func (b *BaseDpo) GetUser() types.User                             { return nil }
func (b *BaseDpo) BindUser(u types.User)                           {}
func (b *BaseDpo) UnBindUser()                                     {}
func (b *BaseDpo) MoveFileTo(name, dstName string) (string, error) { return "", nil }
func (b *BaseDpo) DeleteFile(name string) error                    { return nil }
func (b *BaseDpo) ProcessFile(name string, f func(multipart.File, *multipart.FileHeader) error) error {
	return nil
}
func (b *BaseDpo) SetRoom(room string)                             {}
func (b *BaseDpo) Close()                                          {}
func (b *BaseDpo) SendMessage(message interface{}, user ...string) {}
func (b *BaseDpo) SendRoomMessage(message interface{})             {}
func (b *BaseDpo) Proxy(string) error                              { return nil }
