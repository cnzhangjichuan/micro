package packet

import "sync"

var bytesEnv struct {
	chunks [16]sync.Pool
}

func init() {
	capacity := uint32(16)
	creator := func(capacity uint32) func() interface{} {
		return func() interface{} {
			return make([]byte, 0, capacity)
		}
	}
	for i := 0; i < 16; i++ {
		bytesEnv.chunks[i].New = creator(capacity)
		capacity <<= 1
	}
}

// getBytes 获取(cap >= n)缓冲区
// 初始长度为0
func getBytes(n int) []byte {
	var i int
	if n <= 16 {
		i = 0
	} else if n <= 32 {
		i = 1
	} else if n <= 64 {
		i = 2
	} else if n <= 128 {
		i = 3
	} else if n <= 256 {
		i = 4
	} else if n <= 512 {
		i = 5
	} else if n <= 1024 {
		// 1K
		i = 6
	} else if n <= 2048 {
		// 2K
		i = 7
	} else if n <= 4096 {
		// 4K
		i = 8
	} else if n <= 8192 {
		// 8K
		i = 9
	} else if n <= 16384 {
		// 16K
		i = 10
	} else if n <= 32768 {
		// 32K
		i = 11
	} else if n <= 65536 {
		// 64K
		i = 12
	} else if n <= 131072 {
		// 128K
		i = 13
	} else if n <= 262144 {
		// 256K
		i = 14
	} else if n <= 524288 {
		// 512K
		i = 15
	} else {
		i = -1
	}
	if i < 0 {
		return make([]byte, n)
	}
	b := bytesEnv.chunks[i].Get().([]byte)
	return b[:0]
}

// putBytes 放入缓冲区到池中
func putBytes(b []byte) {
	var (
		i int
		cp = cap(b)
	)
	if cp >= 524288 {
		// 512K
		i = 15
	} else if cp >= 262144 {
		// 256K
		i = 14
	} else if cp >= 131072 {
		// 128K
		i = 13
	} else if cp >= 65536 {
		// 64K
		i = 12
	} else if cp >= 32768 {
		// 32K
		i = 11
	} else if cp >= 16384 {
		// 16K
		i = 10
	} else if cp >= 8192 {
		// 8K
		i = 9
	} else if cp >= 4096 {
		// 4K
		i = 8
	} else if cp >= 2048 {
		// 2K
		i = 7
	} else if cp >= 1024 {
		// 1K
		i = 6
	} else if cp >= 512 {
		i = 5
	} else if cp >= 256 {
		i = 4
	} else if cp >= 128 {
		i = 3
	} else if cp >= 64 {
		i = 2
	} else if cp >= 32 {
		i = 1
	} else if cp >= 16 {
		i = 0
	} else {
		return
	}
	bytesEnv.chunks[i].Put(b)
}
