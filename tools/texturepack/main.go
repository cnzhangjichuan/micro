package main

import (
	"io/ioutil"
	"os"

	"github.com/micro/packet"
	"github.com/micro/tools/texturepack/texture"
)

func main() {
	const src = `../Resources`
	const out = `../Publish`

	fs, err := ioutil.ReadDir(src)
	if err != nil {
		println("err:", err)
	}

	var res []string
	for _, f := range fs {
		if f.IsDir() {
			res = append(res, src+"/"+f.Name())
		}
	}

	tp := texture.NewPngCombination(2048, 2048, out)
	pack := packet.New(1024 * 1024 * 5)
	err = tp.Combination(res, pack)
	if err != nil {
		println("texture packet err", err.Error())
	} else {
		ioutil.WriteFile(out+"/textuepacket", pack.Data(), os.ModePerm)
		println("texture packet successful")
	}
	packet.Free(pack)
}
