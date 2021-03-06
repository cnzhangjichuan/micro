package micro

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"time"

	"github.com/micro/packet"
	"github.com/micro/xutils"
)

type authorize struct {
	mask []byte
}

// Init 初始化
func (a *authorize) Init(mask string) {
	if mask == "" {
		mask = "JCZ2020"
	}
	a.mask = []byte(mask)
}

// NewCode 生成校验码
func NewCode(code string) string {
	return env.authorize.NewCode(code)
}

// NewToken 生成Token
func NewToken(code string) string {
	return env.authorize.NewToken(code)
}

// NewCode 成生校验码
func (a *authorize) NewCode(code string) string {
	if code == "" {
		code = xutils.GUID(0)
	}

	pack := packet.New(512)
	// add code
	pack.WriteString(code)
	ls := pack.Size()
	// add timestamp
	binary.BigEndian.PutUint64(pack.Allocate(8), uint64(time.Now().Unix()))
	// sign
	sign := md5.Sum(pack.Data())
	// mask
	pack.Mask(a.mask, 0, ls)

	// append md5 code
	pack.Seek(0, ls)
	pack.Write(sign[:])

	// hex encode
	src := pack.Data()
	capacity := hex.EncodedLen(pack.Size())
	dst := pack.Allocate(capacity)
	hex.Encode(dst, src)
	s := string(dst)
	packet.Free(pack)
	return s
}

// Check 校验码值是否合法
func (a *authorize) Check(as string) (code string, ok bool) {
	if as == "" {
		return
	}

	pack := packet.New(512)

	// hex decode
	src := xutils.UnsafeStringToBytes(as)
	_, err := hex.Decode(pack.Allocate(hex.DecodedLen(len(src))), src)
	if err != nil {
		packet.Free(pack)
		return
	}

	// md5 code
	sdx := pack.Size() - md5.Size
	vCode := pack.Slice(sdx, -1)
	last := sdx + 8
	pack.Seek(0, last)
	pack.Write(vCode)
	vCode = pack.Slice(last, -1)

	// unmask
	pack.Mask(a.mask, 0, sdx)
	code = pack.ReadString()

	// check vCode
	now := uint64(time.Now().Unix())
	dsw := now - 6
	for i := now; i >= dsw; i-- {
		pack.Seek(0, sdx)
		binary.BigEndian.PutUint64(pack.Allocate(8), i)
		sign := md5.Sum(pack.Slice(0, last))
		if bytes.Equal(vCode, sign[:]) {
			ok = true
			break
		}
	}
	packet.Free(pack)

	return
}

// NewToken 创建Token
func (a *authorize) NewToken(s string) string {
	pack := packet.New(512)
	pack.WriteString(s)
	size := pack.Size()
	// sign
	sign := md5.Sum(pack.Data())
	pack.Write(sign[:])
	// mask
	pack.Mask(a.mask, 0, size)

	// hex encode
	src := pack.Data()
	cpa := hex.EncodedLen(pack.Size())
	dst := pack.Allocate(cpa)
	hex.Encode(dst, src)
	token := string(dst)
	packet.Free(pack)
	return token
}

// CheckToken 校验Token是否合法
func (a *authorize) CheckToken(as string) (token string, ok bool) {
	if as == "" {
		return
	}

	pack := packet.New(512)
	// hex decode
	src := xutils.UnsafeStringToBytes(as)
	_, err := hex.Decode(pack.Allocate(hex.DecodedLen(len(src))), src)
	if err != nil {
		packet.Free(pack)
		return
	}

	// md5 code
	sdx := pack.Size() - md5.Size
	pack.Mask(a.mask, 0, sdx)
	sign := md5.Sum(pack.Slice(0, sdx))
	if !bytes.Equal(sign[:], pack.Slice(sdx, -1)) {
		packet.Free(pack)
		return
	}

	// unmask
	token, ok = pack.ReadString(), true
	packet.Free(pack)
	return
}

const (
	VERSION  = `version`
	LOGIN    = `login`
	LoginDBA = `loginDBA`
	REGISTER = `register`
	GUEST    = `loginCutGuest`
)

// CheckAPI 是否可以访问API
func (a *authorize) CheckAPI(uid, api string) bool {
	if uid != "" {
		return true
	}
	return api == VERSION || api == LOGIN || api == LoginDBA || api == REGISTER || api == GUEST
}
