package texture

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/micro/packet"
)

func NewPngCombination(width, height int, oDir string) *PngCombination {
	var p PngCombination
	p.Rectangle = &ImageRectangle{
		X:      0,
		Y:      0,
		With:   width,
		Height: height,
		Use:    false,
	}
	p.Group = 0
	p.oDir = oDir
	os.MkdirAll(p.oDir, os.ModePerm)
	return &p
}

type ImageItem struct {
	Name   string
	Group  int
	Offset image.Rectangle
	Data   image.Image
}

type ImageItems []ImageItem

func (items ImageItems) Len() int {
	return len(items)
}

func (items ImageItems) Less(i, j int) bool {
	return items[i].Data.Bounds().Dx() > items[j].Data.Bounds().Dy()
}

func (items ImageItems) Swap(i, j int) {
	items[i].Name, items[j].Name = items[j].Name, items[i].Name
	items[i].Data, items[j].Data = items[j].Data, items[i].Data
}

type ImageRectangle struct {
	X      int
	Y      int
	With   int
	Height int
	Use    bool
	Right  *ImageRectangle
	Down   *ImageRectangle
}

type PngCombination struct {
	Group     int
	Rectangle *ImageRectangle
	Images    []ImageItem
	oDir      string
}

func (p *PngCombination) Combination(res []string, pack *packet.Packet) error {
	for _, src := range res {
		items, err := p.readFileToItems(src)
		if err != nil {
			return err
		}
		p.Rectangle.X = 0
		p.Rectangle.Y = 0
		p.Rectangle.Use = false
		for i := range items {
			if err := p.setImageItemOffset(&items[i]); err != nil {
				return err
			}
		}
		_, name := filepath.Split(src)
		if err = p.saveItems(items, name, pack); err != nil {
			return err
		}
	}
	return nil
}

func (p *PngCombination) readFileToItems(src string) (items ImageItems, err error) {
	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".png") {
			return nil
		}
		fdd, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		data, err := png.Decode(bytes.NewReader(fdd))
		if err != nil {
			data, err = jpeg.Decode(bytes.NewReader(fdd))
		}
		if err != nil {
			return nil
		}
		fName := strings.Replace(path, `\`, `/`, -1)
		xCount := 0
		for i := len(fName) - 1; i >= 0; i-- {
			if fName[i] == '/' {
				xCount++
				if xCount >= 2 {
					fName = fName[i+1:]
					break
				}
			}
		}
		items = append(items, ImageItem{
			Name: fName,
			Data: data,
		})
		return err
	})
	// 按图片大小降序
	if err == nil {
		sort.Sort(items)
	}
	return
}

func (p *PngCombination) findFreeOffset(width, height int, pRect *ImageRectangle) *ImageRectangle {
	if pRect == nil {
		pRect = p.Rectangle
	}
	if pRect.Use {
		if pRect.Right != nil {
			if ret := p.findFreeOffset(width, height, pRect.Right); ret != nil {
				return ret
			}
		}
		if pRect.Down != nil {
			if ret := p.findFreeOffset(width, height, pRect.Down); ret != nil {
				return ret
			}
		}
		return nil
	}
	if pRect.With < width || pRect.Height < height {
		return nil
	}
	if pRect.With > width {
		pRect.Right = &ImageRectangle{
			X:      pRect.X + width,
			Y:      pRect.Y,
			With:   pRect.With - width,
			Height: height,
			Use:    false,
		}
	}
	if pRect.Height > height {
		pRect.Down = &ImageRectangle{
			X:      pRect.X,
			Y:      pRect.Y + height,
			With:   pRect.With,
			Height: pRect.Height - height,
			Use:    false,
		}
	}
	pRect.Use = true
	return pRect
}

func (p *PngCombination) setImageItemOffset(item *ImageItem) error {
	b := item.Data.Bounds()
	w := b.Dx()
	h := b.Dy()
	if w > p.Rectangle.With || h > p.Rectangle.Height {
		return errors.New("too larger image")
	}
	r := p.findFreeOffset(w, h, nil)
	if r == nil {
		p.Group += 1
		p.Rectangle.Use = false
		p.Rectangle.Right = nil
		p.Rectangle.Down = nil
		p.setImageItemOffset(item)
		return nil
	}
	item.Group = p.Group
	item.Offset.Min.X = r.X + 1
	item.Offset.Min.Y = r.Y + 1
	item.Offset.Max.X = r.X + w
	item.Offset.Max.Y = r.Y + h
	return nil
}

func (p *PngCombination) saveItems(items ImageItems, dstName string, pack *packet.Packet) error {
	for group := 0; group <= p.Group; group++ {
		// set size
		width, height := 0, 0
		count := uint32(0)
		for i := range items {
			if items[i].Group != group {
				continue
			}
			if width < items[i].Offset.Max.X {
				width = items[i].Offset.Max.X
			}
			if height < items[i].Offset.Max.Y {
				height = items[i].Offset.Max.Y
			}
			count += 1
		}
		m := image.NewRGBA(image.Rect(0, 0, width, height))
		var fileName string
		if group == 0 {
			fileName = fmt.Sprintf("%s.png", dstName)
		} else {
			fileName = fmt.Sprintf("%s%d.png", dstName, group)
		}
		pack.WriteString(fileName)
		pack.WriteU32(count)
		for i := range items {
			if items[i].Group != group {
				continue
			}
			draw.Draw(m, items[i].Offset, items[i].Data, image.ZP, draw.Over)
			pack.WriteString(items[i].Name)
			pack.WriteU32(uint32(items[i].Offset.Min.X))
			pack.WriteU32(uint32(items[i].Offset.Min.Y))
			pack.WriteU32(uint32(items[i].Offset.Dx()))
			pack.WriteU32(uint32(items[i].Offset.Dy()))
			println(items[i].Name, items[i].Offset.Min.X, items[i].Offset.Min.Y, items[i].Offset.Dx(), items[i].Offset.Dy())
		}
		fd, err := os.Create(filepath.Join(p.oDir, fileName))
		if err != nil {
			return err
		}
		err = png.Encode(fd, m)
		fd.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
