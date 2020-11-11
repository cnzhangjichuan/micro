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
	pack.WriteString(code)
	ls := pack.Size()
	binary.BigEndian.PutUint64(pack.Allocate(8), uint64(time.Now().Unix()))

	// mask
	pack.Mask(a.mask, 0, pack.Size())
	sign := md5.Sum(pack.Data())

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
	last := sdx+8
	pack.Seek(0, last)
	pack.Write(vCode)
	vCode = pack.Slice(last, -1)

	// unmask
	pack.Mask(a.mask, 0, sdx)
	code = pack.ReadString()

	// check vCode
	now := uint64(time.Now().Unix())
	dsw := now - 3
	for i := now; i >= dsw; i-- {
		pack.Seek(0, sdx)
		binary.BigEndian.PutUint64(pack.Allocate(8), i)
		pack.Mask(a.mask, 0, last)
		sign := md5.Sum(pack.Slice(0, last))
		if bytes.Equal(vCode, sign[:]) {
			ok = true
			break
		}
		pack.Mask(a.mask, 0, sdx)
	}
	packet.Free(pack)

	return
}

// NewToken 创建Token
func (a *authorize) NewToken(s string) string {
	pack := packet.New(512)
	pack.WriteString(s)
	// mask
	pack.Mask(a.mask, 0, pack.Size())
	sign := md5.Sum(pack.Data())
	pack.Write(sign[:])

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
	vCode := pack.Slice(sdx, -1)
	sign := md5.Sum(pack.Slice(0, sdx))
	if !bytes.Equal(sign[:], vCode) {
		packet.Free(pack)
		return
	}

	// unmask
	pack.Mask(a.mask, 0, sdx)
	token, ok = pack.ReadString(), true
	packet.Free(pack)
	return
}

// CheckAPI 是否可以访问API
func (a *authorize) CheckAPI(uid, api string) bool {
	const (
		VERSION  = `version`
		LOGIN    = `login`
		REGISTER = `register`
	)
	if uid != "" {
		return true
	}
	return api == VERSION || api == LOGIN || api == REGISTER
}
