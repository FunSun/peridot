package cartridge

import "git.letus.rocks/funsun/peridot/common"

type Mapper0 struct {
	cpuRouter, ppuRouter common.Router
	cpuBus, ppuBus       common.Bus
	rpg, chr             []uint8
	irq                  func()
}

func (m *Mapper0) Init(rpg, chr []uint8) *Mapper0 {
	m.rpg = rpg
	m.chr = chr
	m.cpuBus = &cpuBus{m}
	m.ppuBus = &ppuBus{m}
	// for i := 0; i < 0xff; i++ {
	// 	fmt.Printf("0xdb%x: 0x%x\n", i, m.rpg[0xdb00-0xc000+i])
	// }
	// panic("foo")
	return m
}

func (m *Mapper0) IRQ() {}

func (m *Mapper0) SetIRQ(irq func()) {
	m.irq = irq
}

func (m *Mapper0) SetCPURouter(r common.Router) {
	r.AddMapping(0x8000, 0x4000, m.cpuBus, true)
	r.AddMapping(0xc000, 0x4000, m.cpuBus, true)
}
func (m *Mapper0) SetPPURouter(r common.Router) {
	r.AddMapping(0x0000, 0x2000, m.ppuBus, false)
}

func (m *Mapper0) PPURead(addr uint16) uint8 {
	return m.chr[addr]
}
func (m *Mapper0) PPUWrite(addr uint16, val uint8) {
	m.chr[addr] = val
}
func (m *Mapper0) CPURead(addr uint16) uint8 {
	return m.rpg[addr]
}
func (m *Mapper0) CPUWrite(uint16, uint8) {
	panic("TODO")
}
