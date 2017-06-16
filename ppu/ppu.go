package ppu

import (
	"fmt"

	"git.letus.rocks/funsun/peridot/common"
)

// see https://wiki.nesdev.com/w/index.php/PPU_scrolling#Tile_and_attribute_fetching
// v: 0 yyy    N        N       YYYYY    XXXXX
//		  findY  NT-HIGH  NT-LOW  COARSEY  COARSEX
// name-table addr: 0010 1        1       11111    11111
//                  FIX  NT-HIGH  NT-LOW  COARSEY  COARSEX
func getNTAddr(v uint16) uint16 {
	return 0x2000 | v&0x0fff
	// return 0x2000 | v&0x03ff
}

// attr-table addr  0010 1        1       1111 111            111
//                  FIX  NT-HIGH NT-LOW   FIX  COARSEY-HIGH3  COARSEX-HIGH3

func getATAddr(v uint16) uint16 {
	return 0x23C0 | (v & 0x0C00) | ((v >> 4) & 0x38) | ((v >> 2) & 0x07)
	// return 0x23C0 | 0x0000 | ((v >> 4) & 0x38) | ((v >> 2) & 0x07)
}

// uint16
//X: NAMETABLE-LOW COARSEX FINEX
//	 A 	           BBBBB   CCC
//Y: NAMETABLE-HIGH COARSEY FINEY
//	 A 	           BBBBB   CCC

type PPU struct {

	// internel reg
	w           uint8
	skip        uint8
	v           uint16
	t           uint16
	fineX       uint16
	tfineX      uint16
	fineXCross  bool
	shiftRegL   uint16
	shiftRegH   uint16
	shiftAttr   uint16
	bgBase      uint16
	spriteBase  uint16
	spriteCache []uint16
	paletee     []uint8
	fb          [][]uint8
	counter     int

	bus           common.Bus
	oam           common.Bus
	oamAddr       uint16
	width, height int
	regs          []uint8
	//reg
	fLargeSprite bool
	fLargeStep   bool
	fNMI         bool

	fRender                  bool
	screen                   common.Screen
	nmi                      func()
	irq                      func()
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

	p.fb = make([][]uint8, p.height)
	p.paletee = make([]uint8, 0x20)
	for i := 0; i < p.height; i++ {
		p.fb[i] = make([]uint8, p.width)
	}

	p.regs = make([]uint8, 8)
	p.Tick = make(chan bool, 10)
	go p.render()
	return p
}

func (p *PPU) SetOAM(oam common.Bus) {
	p.oam = oam
}

func (p *PPU) txTovx() {
	p.v = (p.v & ^uint16(0x041f)) + (p.t & uint16(0x041f))
	p.fineX = p.tfineX
}

// see https://wiki.nesdev.com/w/index.php/PPU_scrolling#Wrapping_around
func (p *PPU) incX() {
	v := p.v
	if (v & 0x001F) == 31 {
		v &= ^uint16(0x001f)
		v ^= 0x0400
	} else {
		v += 1
	}
	p.fineXCross = false
	p.v = v
}

func (p *PPU) incY() {
	v := p.v
	if (v & 0x7000) != 0x7000 {
		v += 0x1000
	} else {
		v &= ^uint16(0x7000)
		y := (v & 0x03e0) >> 5
		switch y {
		case 29:
			y = 0
			v ^= 0x0800
		case 31:
			y = 0
		default:
			y++
		}
		v = (v & ^uint16(0x03e0)) | y<<5
	}
	p.v = v
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

const R = 1
const W = 2
const N = 0

func (p *PPU) Read(addr uint16) uint8 {
	// p.addr = addr
	// p.action = R
	// return p.buffer[addr]
	if addr == 2 {
		p.w = 0
	} else {
		fmt.Println("read ppu addr", addr)
	}
	// fmt.Println(p.regs[addr])
	return p.regs[addr]

}

func (p *PPU) incBaseAddr() {
	if p.fLargeStep {
		// nametable 一行正好32个
		p.v += 32
	} else {

		p.v += 1
	}
}

// var Echo = true

func (p *PPU) Write(addr uint16, val uint8) {
	if addr != 4 {
		fmt.Printf("write ppu 0x%x 0x%x \n", addr, val)
	}

	switch addr {
	case 0:

		p.fLargeStep = (val & 0x04) > 0
		p.spriteBase = uint16((val&0x08)>>3) * 0x1000
		p.bgBase = uint16((val&0x10)>>4) * 0x1000
		p.fLargeSprite = (val & 0x20) > 0
		p.fNMI = (val & 0x80) > 0
		// fmt.Printf("0x%x %v %v\n", val, (val&0x80) > 0, p.fNMI)
		p.t = p.t&(^uint16(0x0c00)) + uint16(val&0x03)<<10
	case 1:
		if val&0x18 > 0 {
			p.fRender = true
		} else {
			p.fRender = false
		}
	case 3:
		p.oamAddr = uint16(val)
	case 4:
		// fmt.Printf("write oam 0x%x 0x%x\n", p.oamAddr, val)
		p.oam.Write(p.oamAddr, val)
		p.oamAddr++
	case 5:
		if p.w == 0 { //write scroll x
			p.tfineX = uint16(val & 0x07)
			p.t = p.t&(^uint16(0x001f)) + uint16(val>>3)

		} else {
			p.t = p.t&(0x0fff) + uint16(val&0x07)
			p.t = p.t&(0xfc1f) + uint16(val>>3)
		}
		p.w ^= 0x01
	case 6:
		if p.w == 0 {
			p.t = (p.t & 0x00ff) + (uint16(val)&0x007f)<<8
		} else {
			p.t = (p.t & 0xff00) + uint16(val)
			p.v = p.t
		}
		p.w ^= 0x01
	case 7:
		// if  0x2000 <= p.baseAddr && p.baseAddr < 0x2fff {
		// if p.baseAddr == 0x27bd && val == 0xb7 {
		// 	common.Echo = false
		// }
		// if Echo {
		// if 0x2000 <= p.v && p.v < 0x2fff {
		// 	if (p.v-0x2000)%0x400 < 960 {
		// 		row := ((p.v - 0x2000) % 0x400) / 32
		// 		col := ((p.v - 0x2000) % 0x400) % 32
		// 		fmt.Printf("write nt%d, %d %d 0x%x\n", (p.v-0x2000)/0x400, row, col, val)
		// 	} else {
		// 		fmt.Printf("write at addr %x %x\n", p.v, val)
		// 	}
		//
		// } else {
		// 	fmt.Printf("write ppu addr %x %x\n", p.v, val)
		// }

		// }

		p.write(p.v, val)
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
		fmt.Println("Frame: ", count)
		// if count == 120 {
		// 	fmt.Println("kill")
		// 	return
		// }
		count++
		if p.fRender {
			p.preRenderScanline()
			for i := 0; i < 240; i++ {
				// fmt.Println("#", i)
				p.visibleScanline(i)
			}
			go p.screen.AddFrameBuffer(common.TileArrayToImage(p.fb))
		} else {
			for i := 0; i < 240; i++ {
				p.postRenderScanline(i)
			}
		}

		for i := 240; i < 260; i++ {
			p.postRenderScanline(i)
		}
	}
}

func (p *PPU) waitTick() {
	<-p.Tick
}

func (p *PPU) loadPalette() {
	for i := 0; i < 0x20; i++ {
		if i%4 == 0 && i != 0 {
			p.paletee[i] = p.paletee[0]
		} else {
			p.paletee[i] = p.bus.Read(uint16(0x3f00 + i))
		}
	}
}

func (p *PPU) renderPixel() {
	var paletteIndex uint8
	var attr uint8
	var hByte, lByte uint8
	if p.fineXCross {
		attr = uint8(p.shiftAttr & 0x00ff)
		hByte = uint8(p.shiftRegH & 0x00ff)
		lByte = uint8(p.shiftRegL & 0x00ff)
	} else {
		attr = uint8(p.shiftAttr >> 8)
		hByte = uint8(p.shiftRegH >> 8)
		lByte = uint8(p.shiftRegL >> 8)
	}

	paletteIndex = attr
	paletteIndex = (paletteIndex << 1) + (hByte&(uint8(0x80)>>p.fineX))>>(7-p.fineX)
	paletteIndex = (paletteIndex << 1) + (lByte&(uint8(0x80)>>p.fineX))>>(7-p.fineX)
	p.fb[p.counter>>8][p.counter&0x00ff] = p.paletee[paletteIndex]

	if p.fineX == 7 {
		p.fineXCross = true
		p.fineX = 0
	} else {
		p.fineX++
	}
	p.counter = (p.counter + 1) % (256 * 240)
}

func (p *PPU) preRenderScanline() {
	// 0
	if p.skip > 0 {
		p.waitTick()
	}
	p.skip ^= 0x01
	// 1
	p.waitTick()
	p.clearVBlank()

	// 2-257
	for i := 2; i < 258; i++ {
		p.waitTick()
	}

	// 258-320
	p.tryIRQ() // trim glitch in middle of screen
	p.tryIRQ()
	p.spriteEvalution()
	p.v = p.t

	// 321 - 336
	p.loadPalette()
	p.reloadShiftReg()

	// 337 - 340
	p.waitTick()
	p.waitTick()
	p.waitTick()
	p.waitTick()

}

func (p *PPU) visibleScanline(row int) {
	// 0
	p.waitTick()
	fmt.Println("row:", row)
	x := (p.v & 0x0400) >> 10
	x = (x << 5) + (p.v & 0x001f)

	// fmt.Printf("x %d ", x)
	y := (p.v & 0x0800) >> 11
	y = (y << 5) + ((p.v >> 5) & 0x001f)
	// fmt.Println("y", y)

	// 1-257
	p.renderBg()
	p.waitTick()
	p.incY()
	p.txTovx()
	// 258 - 320
	p.tryIRQ()
	p.spriteEvalution()
	// 321 - 336
	p.reloadShiftReg()

	// 337 - 340
	p.waitTick()
	p.waitTick()
	p.waitTick()
	p.waitTick()
}

func (p *PPU) postRenderScanline(row int) {
	if row == 241 {
		p.waitTick()
		p.waitTick()
		p.setVBlank()
		for i := 2; i < 340; i++ {
			p.waitTick()
		}
	} else {
		for i := 0; i < 340; i++ {
			p.waitTick()
		}
	}
}

func (p *PPU) renderBg() {
	var ntByte, atByte, lowBG, highBG uint8
	var tileAddr uint16
	var coarseX, coarseY, fineY uint16
	var attr uint16
	var offset uint16
	coarseY = (p.v & uint16(0x03E0)) >> 5
	fineY = p.v >> 12
	// fmt.Printf("0x%x\n", getATAddr(p.v))

	for i := 0; i < 32; i++ {
		// fetch +2 tile from this line
		coarseX = p.v & uint16(0x001f)
		p.waitTick()
		ntByte = p.read(getNTAddr(p.v))
		tileAddr = p.bgBase + uint16(ntByte)*16 + fineY
		p.renderPixel()
		p.waitTick()
		p.renderPixel()

		p.waitTick()
		atByte = p.read(getATAddr(p.v))
		p.renderPixel()
		p.waitTick()
		p.renderPixel()

		p.waitTick()
		lowBG = p.read(tileAddr)
		p.renderPixel()
		p.waitTick()
		p.renderPixel()

		p.waitTick()
		highBG = p.read(tileAddr + 8)
		p.renderPixel()
		p.waitTick()
		p.renderPixel()

		p.shiftRegL = (p.shiftRegL << 8) + uint16(lowBG)
		p.shiftRegH = (p.shiftRegH << 8) + uint16(highBG)
		offset = (coarseY & 0x02) + ((coarseX & 0x02) >> 1)
		attr = uint16(atByte&(0x03<<(offset*2))) >> (offset * 2)
		p.shiftAttr = (p.shiftAttr << 8) + attr
		p.incX()
	}
}

func (p *PPU) tryIRQ() {
	if p.bgBase == 0x0000 && p.spriteBase == 0x1000 && p.fRender {
		p.irq()
	}
}

func (p *PPU) spriteEvalution() {
	for i := 258; i < 321; i++ {
		p.waitTick()
	}

}

func (p *PPU) reloadShiftReg() {
	var ntByte, atByte, lowBG, highBG uint8
	var tileAddr uint16
	var coarseX, coarseY, fineY uint16
	var attr uint16
	var offset uint16
	coarseY = (p.v & uint16(0x03E0)) >> 5
	fineY = p.v >> 12
	p.waitTick()
	for i := 0; i < 2; i++ {
		// fetch +2 tile from this line
		coarseX = p.v & uint16(0x001f)
		p.waitTick()
		ntByte = p.read(getNTAddr(p.v))
		tileAddr = p.bgBase + uint16(ntByte)*16 + fineY
		p.waitTick()

		p.waitTick()
		atByte = p.read(getATAddr(p.v))
		p.waitTick()

		p.waitTick()
		lowBG = p.read(tileAddr)
		p.waitTick()

		p.waitTick()
		highBG = p.read(tileAddr + 8)
		p.waitTick()

		p.shiftRegL = (p.shiftRegL << 8) + uint16(lowBG)
		p.shiftRegH = (p.shiftRegH << 8) + uint16(highBG)
		offset = (coarseY & 0x02) + ((coarseX & 0x02) >> 1)
		attr = uint16(atByte&(0x03<<(offset*2))) >> (offset * 2)
		p.shiftAttr = (p.shiftAttr << 8) + attr
		p.incX()
	}

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

func (p *PPU) SetBus(bus common.Bus) {
	p.bus = bus
}

func (p *PPU) read(addr uint16) uint8 {
	return p.bus.Read(addr)
}

func (p *PPU) write(addr uint16, val uint8) {
	p.bus.Write(addr, val)
}
