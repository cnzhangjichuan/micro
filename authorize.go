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

// NewCode 成生校验码
func NewCode(code string) string {
	return env.authorize.NewCode(code)
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
		return
	}

	// md5 code
	sdx := pack.Size() - md5.Size
	pack.Seek(sdx, -1)
	pins := pack.Copy()
	vCode := pins.Data()

	// unmask
	pack.Seek(0, -1)
	pack.Mask(a.mask, 0, sdx)
	code = pack.ReadString()

	// check vCode
	now := uint64(time.Now().Unix())
	dsw := now - 3
	for i := now; i >= dsw; i-- {
		pack.Seek(0, sdx)
		binary.BigEndian.PutUint64(pack.Allocate(8), i)
		pack.Mask(a.mask, 0, pack.Size())
		sign := md5.Sum(pack.Data())
		if bytes.Equal(vCode, sign[:]) {
			ok = true
			break
		}
		pack.Mask(a.mask, 0, sdx)
	}
	packet.Free(pins)
	packet.Free(pack)

	return
}

func (a *authorize) CheckAPI(uid, api string) bool {
	if uid != "" {
		return true
	}

	return api == "login" || api == "register" || api == "loginDBA" || api == "version"
}
