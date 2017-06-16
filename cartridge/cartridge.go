package cartridge

import "github.com/funsun/peridot/common"

type cpuBus struct {
	m common.ComplexBus
}

func (c *cpuBus) Write(addr uint16, val uint8) {
	c.m.CPUWrite(addr, val)
}

func (c *cpuBus) Read(addr uint16) uint8 {
	return c.m.CPURead(addr)
}

type ppuBus struct {
	m common.ComplexBus
}

func (p *ppuBus) Read(addr uint16) uint8 {
	return p.m.PPURead(addr)
}

func (p *ppuBus) Write(addr uint16, val uint8) {
	p.m.PPUWrite(addr, val)
}

type MMC3 struct {
	cpuRouter, ppuRouter                           common.Router
	cpuBus, ppuBus                                 common.Bus
	cpu0, cpu1, cpu2, cpu3                         func() uint8
	ppu0, ppu1, ppu2, ppu3, ppu4, ppu5, ppu6, ppu7 func() uint8
	bankData                                       func(uint8)
	r0, r1, r2, r3, r4, r5, r6, r7, r8             uint8
	rpg, chr                                       []uint8
	battery                                        []uint8
	irq                                            func()
	fIRQ                                           bool
	irqcounter                                     uint8
	irqlatch                                       uint8
}

func (m *MMC3) Init(rpg, chr []uint8) *MMC3 {
	m.rpg = rpg
	m.chr = chr
	m.cpuBus = &cpuBus{m}
	m.ppuBus = &ppuBus{m}
	m.battery = make([]uint8, 0x2000)
	m.bankSelect(0)
	return m
}

func (m *MMC3) GetCHR() []uint8 {
	return m.chr
}

func (m *MMC3) GetRPG() []uint8 {
	return m.rpg
}

func (m *MMC3) IRQ() {
	// fmt.Println(m.irqcounter)
	if m.irqcounter == 0 {
		if m.irq != nil && m.fIRQ {
			m.irq()
		}
		m.irqcounter = m.irqlatch
		return
	}
	m.irqcounter--
}

func (m *MMC3) SetIRQ(irq func()) {
	m.irq = irq
}

func (m *MMC3) SetCPURouter(r common.Router) {
	m.cpuRouter = r
	r.AddMapping(0x6000, 0x2000, m.cpuBus, false)
	r.AddMapping(0x8000, 0x8000, m.cpuBus, false)
}

func (m *MMC3) SetPPURouter(r common.Router) {
	m.ppuRouter = r
	r.AddMapping(0x0000, 0x2000, m.ppuBus, false)
}

func (m *MMC3) CPUWrite(addr uint16, val uint8) {
	if 0x8000 <= addr && addr < 0xa000 {
		if (addr % 2) == 0 {
			// fmt.Printf("bank select 0x%x\n", val)
			m.bankSelect(val)
			return
		}
		// fmt.Printf("bank data 0x%x\n", val)
		m.bankData(val)
		// fmt.Printf("R0-5, %d %d %d %d %d %d\n", m.r0, m.r1, m.r2, m.r3, m.r4, m.r5)
		return
	} else if 0xa000 <= addr && addr <= 0xbfff {
		if (addr % 2) == 0 {
			m.mirroring(val)
			return
		}
		m.ramProtect(val)
		return
	} else if 0xc000 <= addr && addr <= 0xdfff {
		if (addr % 2) == 0 {
			m.irqLatch(val)
			return
		}
		m.irqReload()
		return
	} else if 0xe000 <= addr && addr <= 0xffff {
		if (addr % 2) == 0 {
			m.irqDisable()
			return
		}
		m.irqEnable()
		return
	} else if 0x6000 <= addr && addr <= 0x7fff { // battery ram
		m.battery[addr-0x6000] = val
		return
	}
	panic("wrong addr")
}

func (m *MMC3) CPURead(addr uint16) uint8 {
	var val uint8
	if 0x8000 <= addr && addr <= 0x9fff {
		val = m.findRPG(m.cpu0, addr-0x8000)
	} else if 0xa000 <= addr && addr <= 0xbfff {
		val = m.findRPG(m.cpu1, addr-0xa000)
	} else if 0xc000 <= addr && addr <= 0xdfff {
		val = m.findRPG(m.cpu2, addr-0xc000)
	} else if 0xe000 <= addr && addr <= 0xffff {
		val = m.findRPG(m.cpu3, addr-0xe000)
	} else if 0x6000 <= addr && addr <= 0x7fff {
		val = m.battery[addr-0x6000]
	}

	return val
}

func (m *MMC3) PPUWrite(addr uint16, val uint8) {
	// fmt.Printf("write ppu 0x%x to 0x%x\n", val, addr)
	// return
	var reg func() uint8
	offset := uint16(0)
	if 0x0000 <= addr && addr <= 0x03ff {
		reg = m.ppu0
		offset = addr - 0x0000
	} else if 0x0400 <= addr && addr <= 0x07ff {
		reg = m.ppu1
		offset = addr - 0x0400
	} else if 0x0800 <= addr && addr <= 0x0bff {
		reg = m.ppu2
		offset = addr - 0x0800
	} else if 0x0c00 <= addr && addr <= 0x0fff {
		reg = m.ppu3
		offset = addr - 0x0c00
	} else if 0x1000 <= addr && addr <= 0x13ff {
		reg = m.ppu4
		offset = addr - 0x1000
	} else if 0x1400 <= addr && addr <= 0x17ff {
		reg = m.ppu5
		offset = addr - 0x1400
	} else if 0x1800 <= addr && addr <= 0x1bff {
		reg = m.ppu6
		offset = addr - 0x1800
	} else if 0x1c00 <= addr && addr <= 0x1fff {
		reg = m.ppu7
		offset = addr - 0x1c00
	}
	m.writeHR(reg, offset, val)
}

func (m *MMC3) PPURead(addr uint16) uint8 {
	if 0x0000 <= addr && addr <= 0x03ff {
		return m.findCHR(m.ppu0, addr-0x0000)
	} else if 0x0400 <= addr && addr <= 0x07ff {
		return m.findCHR(m.ppu1, addr-0x0400)
	} else if 0x0800 <= addr && addr <= 0x0bff {
		return m.findCHR(m.ppu2, addr-0x0800)
	} else if 0x0c00 <= addr && addr <= 0x0fff {
		return m.findCHR(m.ppu3, addr-0x0c00)
	} else if 0x1000 <= addr && addr <= 0x13ff {
		return m.findCHR(m.ppu4, addr-0x1000)
	} else if 0x1400 <= addr && addr <= 0x17ff {
		return m.findCHR(m.ppu5, addr-0x1400)
	} else if 0x1800 <= addr && addr <= 0x1bff {
		return m.findCHR(m.ppu6, addr-0x1800)
	} else if 0x1c00 <= addr && addr <= 0x1fff {
		return m.findCHR(m.ppu7, addr-0x1c00)
	}
	panic("wrong addr")
}

func (m *MMC3) bankSelect(val uint8) {
	r := val & 0x07
	switch r {
	case 0:
		m.bankData = func(val uint8) { m.r0 = val }
	case 1:
		m.bankData = func(val uint8) { m.r1 = val }
	case 2:
		m.bankData = func(val uint8) { m.r2 = val }
	case 3:
		m.bankData = func(val uint8) { m.r3 = val }
	case 4:
		m.bankData = func(val uint8) { m.r4 = val }
	case 5:
		m.bankData = func(val uint8) { m.r5 = val }
	case 6:
		m.bankData = func(val uint8) { m.r6 = val }
	case 7:
		m.bankData = func(val uint8) { m.r7 = val }
	}

	if (val & 0x80) == 0 {
		m.ppu0 = func() uint8 { return m.r0 & 0xfe }
		m.ppu1 = func() uint8 { return m.r0 | 0x01 }
		m.ppu2 = func() uint8 { return m.r1 & 0xfe }
		m.ppu3 = func() uint8 { return m.r1 | 0x01 }
		m.ppu4 = func() uint8 { return m.r2 }
		m.ppu5 = func() uint8 { return m.r3 }
		m.ppu6 = func() uint8 { return m.r4 }
		m.ppu7 = func() uint8 { return m.r5 }
	} else {
		m.ppu0 = func() uint8 { return m.r2 }
		m.ppu1 = func() uint8 { return m.r3 }
		m.ppu2 = func() uint8 { return m.r4 }
		m.ppu3 = func() uint8 { return m.r5 }
		m.ppu4 = func() uint8 { return m.r0 & 0xfe }
		m.ppu5 = func() uint8 { return m.r0 | 0x01 }
		m.ppu6 = func() uint8 { return m.r1 & 0xfe }
		m.ppu7 = func() uint8 { return m.r1 | 0x01 }
	}

	if (val & 0x40) == 0 {
		m.cpu0 = func() uint8 { return m.r6 }
		m.cpu1 = func() uint8 { return m.r7 }
		m.cpu2 = func() uint8 { return uint8(len(m.rpg)/(8*1024) - 2) }
		m.cpu3 = func() uint8 { return uint8(len(m.rpg)/(8*1024) - 1) }
	} else {
		m.cpu0 = func() uint8 { return uint8(len(m.rpg)/(8*1024) - 2) }
		m.cpu1 = func() uint8 { return m.r7 }
		m.cpu2 = func() uint8 { return m.r6 }
		m.cpu3 = func() uint8 { return uint8(len(m.rpg)/(8*1024) - 1) }
	}
}

func (m *MMC3) mirroring(val uint8) {}

func (m *MMC3) ramProtect(val uint8) {}

func (m *MMC3) irqLatch(val uint8) {
	m.irqlatch = val
}

func (m *MMC3) irqReload() {
	m.irqcounter = m.irqlatch
}

func (m *MMC3) irqDisable() {
	m.fIRQ = false
}

func (m *MMC3) irqEnable() {
	m.fIRQ = true
}

func (m *MMC3) findRPG(r func() uint8, offset uint16) uint8 {
	return m.rpg[int(r())*8*1024+int(offset)]
}

func (m *MMC3) findCHR(r func() uint8, offset uint16) uint8 {
	return m.chr[int(r())*1024+int(offset)]
}

func (m *MMC3) writeHR(r func() uint8, offset uint16, val uint8) {
	m.chr[int(r())*1024+int(offset)] = val
}
