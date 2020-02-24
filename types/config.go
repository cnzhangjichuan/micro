package types

import "time"

type EnvConfig struct {
	Id            string
	Port          string
	Mask          string
	Register      string
	ConnCacheSize int
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
}
