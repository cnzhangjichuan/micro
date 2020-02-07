package types

type User interface {
	GetUserId() string
	Access(string) bool
	OnLogout()
	SetTimeStamp(int64)
	GetTimeStamp() int64
}
