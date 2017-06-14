package main

import (
	"git.letus.rocks/funsun/peridot/controller"
	"git.letus.rocks/funsun/peridot/cpu"
	"git.letus.rocks/funsun/peridot/ines"
	"git.letus.rocks/funsun/peridot/motherboard"
	"git.letus.rocks/funsun/peridot/ppu"
	"git.letus.rocks/funsun/peridot/ram"
	"git.letus.rocks/funsun/peridot/screen"
)

func main() {

	mb := new(motherboard.MotherBoard).Init()
	c := new(cpu.CPU).Init()
	c.SetBus(mb.CPUBus)
	p := new(ppu.PPU).Init()
	mb.AddCPU(c.Tick)
	mb.AddPPU(p.Tick)
	cpuRAM := new(ram.RAM).Init(2048)
	mockAPU := new(ram.RAM).Init(32)
	oam := new(ram.RAM).Init(1024)
	p.SetOAM(oam)
	ctrl := new(controller.Controller).Init()
	s := new(screen.Screen).Init(1000, 800, ctrl)
	p.SetScreen(s)
	p.SetBus(mb.PPUBus)
	vram := new(ram.RAM).Init(8 * 1024)
	// TODO： 存在mirror的问题
	mb.PPUBus.AddMapping(0x2000, 0x2000, vram, true)
	// d := new(dma.DMA).Init(cpuRAM, oam, 0x0000)
	// reference https://en.wikibooks.org/wiki/NES_Programming/Memory_Map
	mb.CPUBus.AddMapping(0x0000, 0x0800, cpuRAM, true)
	mb.CPUBus.AddMapping(0x0800, 0x0800, cpuRAM, true)
	mb.CPUBus.AddMapping(0x1000, 0x0800, cpuRAM, true)
	mb.CPUBus.AddMapping(0x1800, 0x0800, cpuRAM, true)
	mb.CPUBus.AddMapping(0x2000, 8, p, true)
	// mb.CPUBus.AddMapping(0x4014, 1, d, true)
	mb.CPUBus.AddMapping(0x4000, 0x14, mockAPU, true)
	mb.CPUBus.AddMapping(0x4015, 1, mockAPU, true)
	mb.CPUBus.AddMapping(0x4049, 1, mockAPU, true)

	mb.CPUBus.AddMapping(0x4016, 2, ctrl, true)

	p.SetNMI(c.NMI)
	// c.AddRising(p.OnRising)
	// c.AddFalling(p.OnFalling)
	cart := ines.ReadFile("./test.nes")
	cart.SetCPURouter(mb.CPUBus)
	cart.SetPPURouter(mb.PPUBus)
	p.SetIRQ(cart.IRQ)
	cart.SetIRQ(c.IRQ)
	s.Show()
	mb.Start()
}