package ines

import (
	"fmt"
	"io/ioutil"

	"git.letus.rocks/funsun/peridot/cartridge"
	"git.letus.rocks/funsun/peridot/common"
)

func ReadFile(filename string) common.Cartridge {
	data, _ := ioutil.ReadFile(filename)
	return Read([]uint8(data))
}

func Read(data []uint8) common.Cartridge {
	header := data[0:16]
	mapper := getMapperNumber(header)
	base := 16
	rpg := data[base : base+int(header[4])*16*1024]
	base = 16 + int(header[4])*16*1024
	chr := data[base : base+int(header[5])*8*1024]
	fmt.Println("mapper:", mapper)
	switch mapper {
	case 0x00:
		return new(cartridge.Mapper0).Init(rpg, chr)
	case 0x4a:
		return new(cartridge.MMC3).Init(rpg, chr)
	}
	panic("no mapper founded")
}

func getMapperNumber(header []uint8) uint8 {
	l := header[6] >> 4
	h := header[7] & 0xf0
	return h + l
}
