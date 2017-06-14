package main

import (
	"fmt"
	"git.letus.rocks/funsun/peridot/cartridge"
	"git.letus.rocks/funsun/peridot/common"
	"git.letus.rocks/funsun/peridot/controller"
	"git.letus.rocks/funsun/peridot/ines"
	"git.letus.rocks/funsun/peridot/screen"
	"os"
	"strconv"
)

var palette = map[uint8]uint8{
	0: 0x0d,
	1: 0x00,
	2: 0x10,
	3: 0x20,
}

func main() {
	filename := os.Args[1]
	index, _ := strconv.Atoi(os.Args[2])
	y, _ := strconv.Atoi(os.Args[3]) //width
	x, _ := strconv.Atoi(os.Args[4]) // height

	mmc3 := ines.ReadFile(filename)
	chr := mmc3.(*cartridge.MMC3).GetCHR()
	fmt.Printf("TILE NUM: %d (0x%x)\n", len(chr)/16, len(chr)/16)
	s := new(screen.Screen).Init(1000, 800, new(controller.Controller).Init())
	s.Show()
	s.AddFrameBuffer(common.TileArrayToImage(makePattenTable(chr, index, x, y).Buffer))
	ch := make(chan bool)
	<-ch
}

func makePattenTable(chr []uint8, index, x, y int) *common.TileArray {
	fb := new(common.TileArray).Init(uint16(y)*8, uint16(x)*8)
	for i := uint16(0); i < uint16(x); i++ {
		for j := uint16(0); j < uint16(y); j++ {
			// 注意页和1K的区别
			tileNum := i*uint16(y) + j + uint16(index)
			base := int(tileNum) * 16
			tile := common.RenderTile(common.PatternToTile(chr[base:base+16]), palette)
			fb.AddTile(i*8, j*8, tile)
		}
	}
	return fb
}
