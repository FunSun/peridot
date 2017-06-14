package rom

type ROM struct {
	mem []uint8
}

func (rom *ROM) Init(size uint16) *ROM {
	rom.mem = make([]uint8, size)
	return rom
}

func (rom *ROM) Read(addr uint16) uint8 {
	return rom.mem[addr]
}

func (rom *ROM) Write(addr uint16, val uint8) {
	panic("cannot write")
}

func (rom *ROM) LoadData(start uint16, data []uint8) {
	for i := uint16(0); i < uint16(len(data)); i++ {
		rom.mem[start+i] = data[i]
	}
}
