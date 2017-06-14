package common

import (
	"image"
	"image/color"
)

type TileArray struct {
	w, h   uint16
	Buffer [][]uint8
}

func (fb *TileArray) Init(w, h uint16) *TileArray {
	fb.w = w
	fb.h = h
	fb.Buffer = make([][]uint8, h)
	for i := uint16(0); i < h; i++ {
		fb.Buffer[i] = make([]uint8, w)
	}
	return fb
}

// 这里坐标和nes也是行x列y， 到opengl的转换是由screen做的
func (fb *TileArray) AddTile(x, y uint16, tile [][8]uint8) {
	// for big sprite

	endX := x + uint16(len(tile))
	endY := y + uint16(8)

	if endX > uint16(len(fb.Buffer)) {
		endX = uint16(len(fb.Buffer))
	}
	if endY > uint16(len(fb.Buffer[0])) {
		endY = uint16(len(fb.Buffer[0]))
	}

	for i := uint16(0); x+i < endX; i++ {
		for j := uint16(0); y+j < endY; j++ {
			fb.Buffer[x+i][y+j] = tile[i][j]
		}
	}
}

func PatternToTile(pattern []uint8) [][8]uint8 {
	expanded := make([][8]uint8, 8)
	for i := uint16(0); i < 8; i++ {
		lByte := pattern[i]
		hByte := pattern[i+8]
		for j := uint8(0); j < 8; j++ {
			expanded[i][j] = zipByte(hByte, lByte, j)
		}
	}
	return expanded
}

func zipByte(hByte, lByte uint8, idx uint8) uint8 {
	h := (hByte & (1 << (7 - idx))) > 0
	l := (lByte & (1 << (7 - idx))) > 0
	res := uint8(0)
	if h {
		res += 2
	}
	if l {
		res += 1
	}
	return res
}

func TileArrayToImage(fb [][]uint8) image.Image {
	// use golang image coordinates
	res := image.NewRGBA(image.Rect(0, 0, len(fb[0]), len(fb)))
	for i := 0; i < len(fb); i++ {
		for j := 0; j < len(fb[i]); j++ {
			res.Set(j, i, NES_COLOR_MAP[fb[i][j]])
		}
	}
	return res
}

var NES_COLOR_MAP = []*color.RGBA{
	&color.RGBA{124, 124, 124, 0},
	&color.RGBA{0, 0, 252, 0},
	&color.RGBA{0, 0, 188, 0},
	&color.RGBA{68, 40, 188, 0},
	&color.RGBA{148, 0, 132, 0},
	&color.RGBA{168, 0, 32, 0},
	&color.RGBA{168, 16, 0, 0},
	&color.RGBA{136, 20, 0, 0},
	&color.RGBA{80, 48, 0, 0},
	&color.RGBA{0, 120, 0, 0},
	&color.RGBA{0, 104, 0, 0},
	&color.RGBA{0, 88, 0, 0},
	&color.RGBA{0, 64, 88, 0},
	&color.RGBA{0, 0, 0, 0},
	&color.RGBA{0, 0, 0, 0},
	&color.RGBA{0, 0, 0, 0},
	&color.RGBA{188, 188, 188, 0},
	&color.RGBA{0, 120, 248, 0},
	&color.RGBA{0, 88, 248, 0},
	&color.RGBA{104, 68, 252, 0},
	&color.RGBA{216, 0, 204, 0},
	&color.RGBA{228, 0, 88, 0},
	&color.RGBA{248, 56, 0, 0},
	&color.RGBA{228, 92, 16, 0},
	&color.RGBA{172, 124, 0, 0},
	&color.RGBA{0, 184, 0, 0},
	&color.RGBA{0, 168, 0, 0},
	&color.RGBA{0, 168, 68, 0},
	&color.RGBA{0, 136, 136, 0},
	&color.RGBA{0, 0, 0, 0},
	&color.RGBA{0, 0, 0, 0},
	&color.RGBA{0, 0, 0, 0},
	&color.RGBA{248, 248, 248, 0},
	&color.RGBA{60, 188, 252, 0},
	&color.RGBA{104, 136, 252, 0},
	&color.RGBA{152, 120, 248, 0},
	&color.RGBA{248, 120, 248, 0},
	&color.RGBA{248, 88, 152, 0},
	&color.RGBA{248, 120, 88, 0},
	&color.RGBA{252, 160, 68, 0},
	&color.RGBA{248, 184, 0, 0},
	&color.RGBA{184, 248, 24, 0},
	&color.RGBA{88, 216, 84, 0},
	&color.RGBA{88, 248, 152, 0},
	&color.RGBA{0, 232, 216, 0},
	&color.RGBA{120, 120, 120, 0},
	&color.RGBA{0, 0, 0, 0},
	&color.RGBA{0, 0, 0, 0},
	&color.RGBA{252, 252, 252, 0},
	&color.RGBA{164, 228, 252, 0},
	&color.RGBA{184, 184, 248, 0},
	&color.RGBA{216, 184, 248, 0},
	&color.RGBA{248, 184, 248, 0},
	&color.RGBA{248, 164, 192, 0},
	&color.RGBA{240, 208, 176, 0},
	&color.RGBA{252, 224, 168, 0},
	&color.RGBA{248, 216, 120, 0},
	&color.RGBA{216, 248, 120, 0},
	&color.RGBA{184, 248, 184, 0},
	&color.RGBA{184, 248, 216, 0},
	&color.RGBA{0, 252, 252, 0},
	&color.RGBA{248, 216, 248, 0},
	&color.RGBA{0, 0, 0, 0},
	&color.RGBA{0, 0, 0, 0},
}

func RenderTile(pattern [][8]uint8, paletee map[uint8]uint8) [][8]uint8 {
	tile := make([][8]uint8, 8)
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			tile[i][j] = paletee[pattern[i][j]]
		}
	}
	return tile
}
