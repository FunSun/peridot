package ppu

import (
	"git.letus.rocks/funsun/peridot/common"
	_ "image/png"
)

type PPU struct {
	stream           chan uint16
	done             chan bool
	fb               *common.TileArray
	bus              common.Bus
	oam              common.Bus
	oamAddr          uint16
	xorigin, yorigin uint16
	width, height    uint16
	buffer, regs     []uint8
	//reg
	fLargeSprite             bool
	fLargeStep               bool
	fNMI                     bool
	fSkip                    bool
	counter                  int
	cycleCounter             int
	state                    int
	screen                   common.Screen
	nmi                      func()
	irq                      func()
	baseAddr                 uint16
	fSpriteOffset, fBgOffset bool
	vblank                   bool
	Tick                     chan bool

	addr   uint16
	data   uint8
	action int
}

func (p *PPU) Init() *PPU {

	p.width = 256
	p.height = 240
	p.stream = make(chan uint16, 16)
	p.done = make(chan bool)
	p.fb = new(common.TileArray).Init(p.width, p.height)
	p.cycleCounter = -10
	p.regs = make([]uint8, 8)
	p.Tick = make(chan bool, 10)
	go p.onTick()
	go p.render()
	return p
}

func (p *PPU) SetOAM(oam common.Bus) {
	p.oam = oam
}

func (p *PPU) onTick() {
	for {
		<-p.Tick
		p.updateState()
		// <-p.Tick
		// <-p.Tick
		// <-p.Tick
	}
}

// func (p *PPU) OnRising() {
// 	p.buffer = p.regs[0:len(p.regs)]
// }
//
// func (p *PPU) OnFalling() {
// 	if p.action != N {
// 		// fmt.Printf("PPU %d addr: 0x%x, val: 0x%x\n", p.action, p.addr, p.data)
// 		p.updateChanges()
// 		p.action = N
// 	}
// }

func (p *PPU) updateState() {
	if p.cycleCounter == 0 {
		// fmt.Println("render start")
	}
	if p.cycleCounter < 241*341 && p.cycleCounter%341 == 0 {
		row := p.cycleCounter / 341
		if row < 240 && row%8 == 0 {
			p.stream <- uint16(row / 8)
		}
		if row == 240 {
			// becarfule uint16 not have negative number
			for i := 63; i >= 0; i-- {
				p.addSpriteTile(p.fb, uint16(i))
			}
		}

	} else if p.cycleCounter == 241*341+1 {
		// for i := 0; i < len(p.fb.Buffer); i++ {
		// 	for j := 0; j < len(p.fb.Buffer[0]); j++ {
		// 		fmt.Printf("0x%x ", p.fb.Buffer[i][j])
		// 	}
		// }
		// fmt.Println("render end")
		go p.screen.AddFrameBuffer(common.TileArrayToImage(p.fb.Buffer))

		p.setVBlank()
	} else if p.cycleCounter == 260*341+1 {
		p.clearVBlank()
	} else if p.cycleCounter < 241*341 && (p.cycleCounter%341) == 260 {
		if (p.cycleCounter/341)%8 == 0 && p.cycleCounter < 240*341 {
			<-p.done
		}

		if p.irq != nil {
			// fmt.Printf("%d ", p.cycleCounter/341)
			p.irq()
		}
	}
	if p.cycleCounter == 0 && p.needSkip() {
		p.cycleCounter++
	}
	p.cycleCounter = (p.cycleCounter + 1) % (261 * 341)
}

func (p *PPU) updateChanges() {
}

func (p *PPU) setVBlank() {
	// for i := uint16(0); i < 0x20; i++ {
	// 	fmt.Printf("%x is %x\n", 0x3f00+i, p.read(0x3f00+i))
	// }
	p.regs[2] = p.regs[2] | (1 << 7)
	if p.fNMI {
		p.nmi()
	}
}

func (p *PPU) clearVBlank() {
	p.regs[2] = p.regs[2] & (^(uint8(1) << 7))
	p.oamAddr = 0
}

func (p *PPU) needSkip() bool {
	skip := p.fSkip
	p.fSkip = !p.fSkip
	return skip
}

const R = 1
const W = 2
const N = 0

func (p *PPU) Read(addr uint16) uint8 {
	// p.addr = addr
	// p.action = R
	// return p.buffer[addr]
	if addr == 2 {
		p.baseAddr = 0
	} else {
		// fmt.Println("addr is: ", addr)
	}
	// fmt.Println(p.regs[addr])
	return p.regs[addr]

}

func (p *PPU) incBaseAddr() {
	if p.fLargeStep {
		// nametable 一行正好32个
		p.baseAddr += 32
	} else {

		p.baseAddr += 1
	}
}

var Echo = true

func (p *PPU) Write(addr uint16, val uint8) {
	switch addr {
	case 0:
		// fmt.Printf("0x%x %v\n", val, (val&0x80) > 0)
		p.fLargeStep = (val & 0x04) > 0
		p.fSpriteOffset = (val & 0x08) > 0
		p.fBgOffset = (val & 0x10) > 0
		p.fLargeSprite = (val & 0x20) > 0
		p.fNMI = (val & 0x80) > 0
		// fmt.Printf("0x%x %v %v\n", val, (val&0x80) > 0, p.fNMI)
		base := val & 0x03
		switch base {
		case 0:
			p.xorigin = 0
			p.yorigin = 0
		case 1:
			p.xorigin = 0
			p.yorigin = 32
		case 2:
			p.xorigin = 30
			p.yorigin = 0
		case 3:
			p.xorigin = 30
			p.yorigin = 32
		}
	case 3:
		p.oamAddr = uint16(val)
	case 4:
		// fmt.Printf("write oam 0x%x 0x%x\n", p.oamAddr, val)
		p.oam.Write(p.oamAddr, val)
		p.oamAddr++
	case 5:
		// fmt.Println(val)
	case 6:
		p.baseAddr = (p.baseAddr << 8) + uint16(val)
	case 7:
		// if  0x2000 <= p.baseAddr && p.baseAddr < 0x2fff {
		if p.baseAddr == 0x27bd && val == 0xb7 {
			common.Echo = false
		}
		// if Echo {
		// 	fmt.Printf("write ppu %x %x\n", p.baseAddr, val)
		// }

		p.write(p.baseAddr, val)
		// if p.baseAddr == 0x27bd {
		// 	fmt.Printf("write ppu %x %x\n", p.baseAddr, val)
		// }

		p.incBaseAddr()
	}
	p.regs[p.addr] = val
}

var foo = false

var pre int64
var count int64

func (p *PPU) render() {
	for {
		i := <-p.stream
		for j := uint16(0); j*8 < p.width; j += 1 {
			p.addBgTile(p.fb, i, j)

		}
		p.done <- true
	}
	// background

	// 地址在前面的遮盖后面的

}

func (p *PPU) SetScreen(s common.Screen) {
	p.screen = s
}

func (p *PPU) SetNMI(nmi func()) {
	p.nmi = nmi
}

func (p *PPU) SetIRQ(irq func()) {
	p.irq = irq
}

func (p *PPU) addBgTile(fb *common.TileArray, x, y uint16) {
	// 这是我自己定的坐标系，肯能和nintendo的不一样，到时候转换一下就行了
	realX := (p.xorigin + x) % 60
	realY := (p.yorigin + y) % 64
	base := uint16(0)
	if realX < 30 {
		if realY < 32 {
			base = 0x2000
		} else {
			base = 0x2400
		}
	} else {
		if realY < 32 {
			base = 0x2800
		} else {
			base = 0x2c00
		}
	}

	inX := realX % 30
	inY := realY % 32

	patternIndex := p.read(base + inX*32 + inY)
	// if x == 29 && y == 29 {
	// 	fmt.Printf("base 0x%x 0x%x\n", base+inX*32+inY, patternIndex)
	// }

	// foo := (base + 960) + (inX/4)*8 + inY/4
	basePaleteeIndex := p.read((base + 960) + (inX/4)*8 + inY/4)
	paleteeOffet := ((inX%4)/2)*2 + (inY%4)/2
	paleteeIndex := (basePaleteeIndex & (0x3 << (paleteeOffet * 2))) >> (paleteeOffet * 2)
	// if x == 0 && y == 0 {
	// 	fmt.Println("###############")
	// }
	// fmt.Printf("%b ", 0x11<<(paleteeOffet*2))
	// if y == 31 {
	// 	fmt.Println("")
	// }
	pattern := p.getPattern(patternIndex, true)
	paletee := p.getPaletee(paleteeIndex, true)
	fb.AddTile(x*8, y*8, common.RenderTile(pattern, paletee))

}

func (p *PPU) addSpriteTile(fb *common.TileArray, index uint16) {
	// delay one scanline accorint to nesdev:oam page
	x := uint16(p.oam.Read(index*8)) + 1
	y := uint16(p.oam.Read(index*8 + 3))
	// sprite 不需要scroll

	// fmt.Println(index, p.oam.Read(index*8+3), p.oam.Read(index*8))
	patternIndex := p.oam.Read(index*8 + 1)
	paleteeIndex := p.oam.Read(index*8+2) & 0x0f
	paletee := p.getPaletee(paleteeIndex, false)
	if p.fLargeSprite {
		h := common.RenderTile(p.getPattern(patternIndex&0xfe, false), paletee)
		l := common.RenderTile(p.getPattern(patternIndex|0x01, false), paletee)
		fb.AddTile(x, y, append(h, l...))
	}
	fb.AddTile(x, y, common.RenderTile(p.getPattern(patternIndex, false), paletee))
}

// var counter uint16 = 0

func (p *PPU) getPattern(index uint8, bg bool) [][8]uint8 {
	pattern := make([][8]uint8, 8)
	base := 0x0000 + uint16(index)*16
	if bg && p.fBgOffset || !bg && p.fSpriteOffset {
		base += 0x1000
	}
	// base = counter * 16
	// counter = (counter + 1) % 512
	for i := uint16(0); i < 8; i++ {
		lByte := p.read(base + i)
		hByte := p.read(base + i + 8)
		for j := uint8(0); j < 8; j++ {
			pattern[i][j] = zipByte(hByte, lByte, j)
		}
	}
	return pattern
}

func (p *PPU) getPaletee(index uint8, bg bool) map[uint8]uint8 {
	paletee := map[uint8]uint8{}
	// paletee[0] = 0
	// paletee[1] = 1
	// paletee[2] = 9
	// paletee[3] = 2
	// return paletee
	paletee[0] = p.read(0x3f00)
	base := uint16(0x3f01)
	if !bg {
		base = uint16(0x3f11)
	}
	base += uint16(index) * 4
	paletee[1] = p.read(base)
	paletee[2] = p.read(base + 1)
	paletee[3] = p.read(base + 2)
	return paletee
}

func (p *PPU) SetBus(bus common.Bus) {
	p.bus = bus
}

func (p *PPU) read(addr uint16) uint8 {
	return p.bus.Read(addr)
}

func (p *PPU) write(addr uint16, val uint8) {
	p.bus.Write(addr, val)
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
