package ram

type RAM struct {
	mem []uint8
}

func (ram *RAM) Init(size uint16) *RAM {
	ram.mem = make([]uint8, size)
	return ram
}

func (ram *RAM) Read(addr uint16) uint8 {
	return ram.mem[addr]
}

func (ram *RAM) Write(addr uint16, val uint8) {
	ram.mem[addr] = val
}
