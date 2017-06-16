package cpu

import (
	"sync"

	"git.letus.rocks/funsun/peridot/common"
)

type CPU struct {
	// p-flag
	N, V, D, I, Z, C bool
	// register
	a, x, y, sp uint8
	// program counter
	pc uint16
	// auxilary
	test     bool
	opcode   uint8
	data     uint8
	buffer   uint8
	adh, adl uint8
	action   func()
	// ioc
	bus common.Bus
	// aux flag
	fInstruction bool
	fDMA         bool
	fStop        bool
	stopped      chan bool
	fCross       uint8
	fY           bool
	needInit     bool
	fIRQ         bool
	fNMI         bool
	Tick         chan bool
	innerTick    chan bool
	skip1        bool
	rw           sync.Mutex

	// falling []func()
	// rising  []func()

	decoder map[uint8]func(cpu *CPU)
	stat    map[uint8]int
}

var starTime int64
var timeCounter int64

func (cpu *CPU) Init() *CPU {
	// cpu.rising = []func(){}
	// cpu.falling = []func(){}
	cpu.Tick = make(chan bool, 10)
	cpu.innerTick = make(chan bool)
	cpu.action = cpu.readIn
	cpu.stat = map[uint8]int{}
	cpu.stopped = make(chan bool)
	cpu.decoder = map[uint8]func(cpu *CPU){
		0x00: func(cpu *CPU) { cpu.brk() },
		0x01: func(cpu *CPU) { cpu.indirectX(); cpu.ora() },
		0x02: func(cpu *CPU) { cpu.future() },
		0x03: func(cpu *CPU) { cpu.future() },
		0x04: func(cpu *CPU) { cpu.future() },
		0x05: func(cpu *CPU) { cpu.zeropage(); cpu.ora() },
		0x06: func(cpu *CPU) { cpu.zeropage(); cpu.asl() },
		0x07: func(cpu *CPU) { cpu.future() },
		0x08: func(cpu *CPU) { cpu.implied(); cpu.php() },
		0x09: func(cpu *CPU) { cpu.immediate(); cpu.ora() },
		0x0a: func(cpu *CPU) { cpu.accumulator(cpu.asl) },
		0x0b: func(cpu *CPU) { cpu.future() },
		0x0c: func(cpu *CPU) { cpu.future() },
		0x0d: func(cpu *CPU) { cpu.absolute(); cpu.ora() },
		0x0e: func(cpu *CPU) { cpu.absolute(); cpu.asl() },
		0x0f: func(cpu *CPU) { cpu.future() },
		0x10: func(cpu *CPU) { cpu.relative(cpu.bpl) },
		0x11: func(cpu *CPU) { cpu.modifyAddr(cpu.indirectY, CN); cpu.ora() },
		0x12: func(cpu *CPU) { cpu.future() },
		0x13: func(cpu *CPU) { cpu.future() },
		0x14: func(cpu *CPU) { cpu.future() },
		0x15: func(cpu *CPU) { cpu.zpIndexed(); cpu.ora() },
		0x16: func(cpu *CPU) { cpu.zpIndexed(); cpu.asl() },
		0x17: func(cpu *CPU) { cpu.future() },
		0x18: func(cpu *CPU) { cpu.implied(); cpu.clc() },
		0x19: func(cpu *CPU) { cpu.modifyAddr(cpu.absIndexed, Y); cpu.ora() },
		0x1a: func(cpu *CPU) { cpu.future() },
		0x1b: func(cpu *CPU) { cpu.future() },
		0x1c: func(cpu *CPU) { cpu.future() },
		0x1d: func(cpu *CPU) { cpu.absIndexed(); cpu.ora() },
		0x1e: func(cpu *CPU) { cpu.modifyAddr(cpu.absIndexed, CP); cpu.asl() },
		0x1f: func(cpu *CPU) { cpu.future() },
		0x20: func(cpu *CPU) { cpu.jsr() },
		0x21: func(cpu *CPU) { cpu.indirectX(); cpu.and() },
		0x22: func(cpu *CPU) { cpu.future() },
		0x23: func(cpu *CPU) { cpu.future() },
		0x24: func(cpu *CPU) { cpu.zeropage(); cpu.bit() },
		0x25: func(cpu *CPU) { cpu.zeropage(); cpu.and() },
		0x26: func(cpu *CPU) { cpu.zeropage(); cpu.rol() },
		0x27: func(cpu *CPU) { cpu.future() },
		0x28: func(cpu *CPU) { cpu.implied(); cpu.plp() },
		0x29: func(cpu *CPU) { cpu.immediate(); cpu.and() },
		0x2a: func(cpu *CPU) { cpu.accumulator(cpu.rol) },
		0x2b: func(cpu *CPU) { cpu.future() },
		0x2c: func(cpu *CPU) { cpu.absolute(); cpu.bit() },
		0x2d: func(cpu *CPU) { cpu.absolute(); cpu.and() },
		0x2e: func(cpu *CPU) { cpu.absolute(); cpu.rol() },
		0x2f: func(cpu *CPU) { cpu.future() },
		0x30: func(cpu *CPU) { cpu.relative(cpu.bmi) },
		0x31: func(cpu *CPU) { cpu.modifyAddr(cpu.indirectY, CN); cpu.and() },
		0x32: func(cpu *CPU) { cpu.future() },
		0x33: func(cpu *CPU) { cpu.future() },
		0x34: func(cpu *CPU) { cpu.future() },
		0x35: func(cpu *CPU) { cpu.zpIndexed(); cpu.and() },
		0x36: func(cpu *CPU) { cpu.zpIndexed(); cpu.rol() },
		0x37: func(cpu *CPU) { cpu.future() },
		0x38: func(cpu *CPU) { cpu.implied(); cpu.sec() },
		0x39: func(cpu *CPU) { cpu.modifyAddr(cpu.absIndexed, Y); cpu.and() },
		0x3a: func(cpu *CPU) { cpu.future() },
		0x3b: func(cpu *CPU) { cpu.future() },
		0x3c: func(cpu *CPU) { cpu.future() },
		0x3d: func(cpu *CPU) { cpu.absIndexed(); cpu.and() },
		0x3e: func(cpu *CPU) { cpu.modifyAddr(cpu.absIndexed, CP); cpu.rol() },
		0x3f: func(cpu *CPU) { cpu.future() },
		0x40: func(cpu *CPU) { cpu.rti() },
		0x41: func(cpu *CPU) { cpu.indirectX(); cpu.eor() },
		0x42: func(cpu *CPU) { cpu.future() },
		0x43: func(cpu *CPU) { cpu.future() },
		0x44: func(cpu *CPU) { cpu.future() },
		0x45: func(cpu *CPU) { cpu.zeropage(); cpu.eor() },
		0x46: func(cpu *CPU) { cpu.zeropage(); cpu.lsr() },
		0x47: func(cpu *CPU) { cpu.future() },
		0x48: func(cpu *CPU) { cpu.implied(); cpu.pha() },
		0x49: func(cpu *CPU) { cpu.immediate(); cpu.eor() },
		0x4a: func(cpu *CPU) { cpu.accumulator(cpu.lsr) },
		0x4b: func(cpu *CPU) { cpu.future() },
		0x4c: func(cpu *CPU) { cpu.jump() },
		0x4d: func(cpu *CPU) { cpu.absolute(); cpu.eor() },
		0x4e: func(cpu *CPU) { cpu.absolute(); cpu.lsr() },
		0x4f: func(cpu *CPU) { cpu.future() },
		0x50: func(cpu *CPU) { cpu.relative(cpu.bvc) },
		0x51: func(cpu *CPU) { cpu.indirectY(); cpu.eor() },
		0x52: func(cpu *CPU) { cpu.future() },
		0x53: func(cpu *CPU) { cpu.future() },
		0x54: func(cpu *CPU) { cpu.future() },
		0x55: func(cpu *CPU) { cpu.zpIndexed(); cpu.eor() },
		0x56: func(cpu *CPU) { cpu.zpIndexed(); cpu.lsr() },
		0x57: func(cpu *CPU) { cpu.future() },
		0x58: func(cpu *CPU) { cpu.implied(); cpu.cli() },
		0x59: func(cpu *CPU) { cpu.modifyAddr(cpu.absIndexed, Y); cpu.eor() },
		0x5a: func(cpu *CPU) { cpu.future() },
		0x5b: func(cpu *CPU) { cpu.future() },
		0x5c: func(cpu *CPU) { cpu.future() },
		0x5d: func(cpu *CPU) { cpu.absIndexed(); cpu.eor() },
		0x5e: func(cpu *CPU) { cpu.modifyAddr(cpu.absIndexed, CP); cpu.lsr() },
		0x5f: func(cpu *CPU) { cpu.future() },
		0x60: func(cpu *CPU) { cpu.rts() },
		0x61: func(cpu *CPU) { cpu.indirectX(); cpu.adc() },
		0x62: func(cpu *CPU) { cpu.future() },
		0x63: func(cpu *CPU) { cpu.future() },
		0x64: func(cpu *CPU) { cpu.future() },
		0x65: func(cpu *CPU) { cpu.zeropage(); cpu.adc() },
		0x66: func(cpu *CPU) { cpu.zeropage(); cpu.ror() },
		0x67: func(cpu *CPU) { cpu.future() },
		0x68: func(cpu *CPU) { cpu.implied(); cpu.pla() },
		0x69: func(cpu *CPU) { cpu.immediate(); cpu.adc() },
		0x6a: func(cpu *CPU) { cpu.accumulator(cpu.ror) },
		0x6b: func(cpu *CPU) { cpu.future() },
		0x6c: func(cpu *CPU) { cpu.jumpIndirect() },
		0x6d: func(cpu *CPU) { cpu.absolute(); cpu.adc() },
		0x6e: func(cpu *CPU) { cpu.absolute(); cpu.ror() },
		0x6f: func(cpu *CPU) { cpu.future() },
		0x70: func(cpu *CPU) { cpu.relative(cpu.bvs) },
		0x71: func(cpu *CPU) { cpu.indirectY(); cpu.adc() },
		0x72: func(cpu *CPU) { cpu.future() },
		0x73: func(cpu *CPU) { cpu.future() },
		0x74: func(cpu *CPU) { cpu.future() },
		0x75: func(cpu *CPU) { cpu.zpIndexed(); cpu.adc() },
		0x76: func(cpu *CPU) { cpu.zpIndexed(); cpu.ror() },
		0x77: func(cpu *CPU) { cpu.future() },
		0x78: func(cpu *CPU) { cpu.implied(); cpu.sei() },
		0x79: func(cpu *CPU) { cpu.modifyAddr(cpu.absIndexed, Y); cpu.adc() },
		0x7a: func(cpu *CPU) { cpu.future() },
		0x7b: func(cpu *CPU) { cpu.future() },
		0x7c: func(cpu *CPU) { cpu.future() },
		0x7d: func(cpu *CPU) { cpu.absIndexed(); cpu.adc() },
		0x7e: func(cpu *CPU) { cpu.modifyAddr(cpu.absIndexed, CP); cpu.ror() },
		0x7f: func(cpu *CPU) { cpu.future() },
		0x80: func(cpu *CPU) { cpu.future() },
		0x81: func(cpu *CPU) { cpu.sta(); cpu.modifyAddr(cpu.indirectX, W) },
		0x82: func(cpu *CPU) { cpu.future() },
		0x83: func(cpu *CPU) { cpu.future() },
		0x84: func(cpu *CPU) { cpu.sty(); cpu.modifyAddr(cpu.zeropage, W) },
		0x85: func(cpu *CPU) { cpu.sta(); cpu.modifyAddr(cpu.zeropage, W) },
		0x86: func(cpu *CPU) { cpu.stx(); cpu.modifyAddr(cpu.zeropage, W) },
		0x87: func(cpu *CPU) { cpu.future() },
		0x88: func(cpu *CPU) { cpu.implied(); cpu.dey() },
		0x89: func(cpu *CPU) { cpu.future() },
		0x8a: func(cpu *CPU) { cpu.implied(); cpu.txa() },
		0x8b: func(cpu *CPU) { cpu.future() },
		0x8c: func(cpu *CPU) { cpu.sty(); cpu.modifyAddr(cpu.absolute, W) },
		0x8d: func(cpu *CPU) { cpu.sta(); cpu.modifyAddr(cpu.absolute, W) },
		0x8e: func(cpu *CPU) { cpu.stx(); cpu.modifyAddr(cpu.absolute, W) },
		0x8f: func(cpu *CPU) { cpu.future() },
		0x90: func(cpu *CPU) { cpu.relative(cpu.bcc) },
		0x91: func(cpu *CPU) { cpu.sta(); cpu.modifyAddr(cpu.indirectY, W|CP) },
		0x92: func(cpu *CPU) { cpu.future() },
		0x93: func(cpu *CPU) { cpu.future() },
		0x94: func(cpu *CPU) { cpu.sty(); cpu.modifyAddr(cpu.zpIndexed, W) },
		0x95: func(cpu *CPU) { cpu.sta(); cpu.modifyAddr(cpu.zpIndexed, W) },
		0x96: func(cpu *CPU) { cpu.stx(); cpu.modifyAddr(cpu.zpIndexed, W|Y) },
		0x97: func(cpu *CPU) { cpu.future() },
		0x98: func(cpu *CPU) { cpu.implied(); cpu.tya() },
		0x99: func(cpu *CPU) { cpu.sta(); cpu.modifyAddr(cpu.absIndexed, W|Y|CP) },
		0x9a: func(cpu *CPU) { cpu.implied(); cpu.txs() },
		0x9b: func(cpu *CPU) { cpu.future() },
		0x9c: func(cpu *CPU) { cpu.future() },
		0x9d: func(cpu *CPU) { cpu.sta(); cpu.modifyAddr(cpu.absIndexed, W|CP) },
		0x9e: func(cpu *CPU) { cpu.future() },
		0x9f: func(cpu *CPU) { cpu.future() },
		0xa0: func(cpu *CPU) { cpu.immediate(); cpu.ldy() },
		0xa1: func(cpu *CPU) { cpu.indirectX(); cpu.lda() },
		0xa2: func(cpu *CPU) { cpu.immediate(); cpu.ldx() },
		0xa3: func(cpu *CPU) { cpu.future() },
		0xa4: func(cpu *CPU) { cpu.zeropage(); cpu.ldy() },
		0xa5: func(cpu *CPU) { cpu.zeropage(); cpu.lda() },
		0xa6: func(cpu *CPU) { cpu.zeropage(); cpu.ldx() },
		0xa7: func(cpu *CPU) { cpu.future() },
		0xa8: func(cpu *CPU) { cpu.implied(); cpu.tay() },
		0xa9: func(cpu *CPU) { cpu.immediate(); cpu.lda() },
		0xaa: func(cpu *CPU) { cpu.implied(); cpu.tax() },
		0xab: func(cpu *CPU) { cpu.future() },
		0xac: func(cpu *CPU) { cpu.absolute(); cpu.ldy() },
		0xad: func(cpu *CPU) { cpu.absolute(); cpu.lda() },
		0xae: func(cpu *CPU) { cpu.absolute(); cpu.ldx() },
		0xaf: func(cpu *CPU) { cpu.future() },
		0xb0: func(cpu *CPU) { cpu.relative(cpu.bcs) },
		0xb1: func(cpu *CPU) { cpu.indirectY(); cpu.lda() },
		0xb2: func(cpu *CPU) { cpu.future() },
		0xb3: func(cpu *CPU) { cpu.future() },
		0xb4: func(cpu *CPU) { cpu.zpIndexed(); cpu.ldy() },
		0xb5: func(cpu *CPU) { cpu.zpIndexed(); cpu.lda() },
		0xb6: func(cpu *CPU) { cpu.modifyAddr(cpu.zpIndexed, Y); cpu.ldx() },
		0xb7: func(cpu *CPU) { cpu.future() },
		0xb8: func(cpu *CPU) { cpu.implied(); cpu.clv() },
		0xb9: func(cpu *CPU) { cpu.modifyAddr(cpu.absIndexed, Y); cpu.lda() },
		0xba: func(cpu *CPU) { cpu.implied(); cpu.tsx() },
		0xbb: func(cpu *CPU) { cpu.future() },
		0xbc: func(cpu *CPU) { cpu.absIndexed(); cpu.ldy() },
		0xbd: func(cpu *CPU) { cpu.absIndexed(); cpu.lda() },
		0xbe: func(cpu *CPU) { cpu.modifyAddr(cpu.absIndexed, Y); cpu.ldx() },
		0xbf: func(cpu *CPU) { cpu.future() },
		0xc0: func(cpu *CPU) { cpu.immediate(); cpu.cpy() },
		0xc1: func(cpu *CPU) { cpu.indirectX(); cpu.cmp() },
		0xc2: func(cpu *CPU) { cpu.future() },
		0xc3: func(cpu *CPU) { cpu.future() },
		0xc4: func(cpu *CPU) { cpu.zeropage(); cpu.cpy() },
		0xc5: func(cpu *CPU) { cpu.zeropage(); cpu.cmp() },
		0xc6: func(cpu *CPU) { cpu.zeropage(); cpu.dec() },
		0xc7: func(cpu *CPU) { cpu.future() },
		0xc8: func(cpu *CPU) { cpu.implied(); cpu.iny() },
		0xc9: func(cpu *CPU) { cpu.immediate(); cpu.cmp() },
		0xca: func(cpu *CPU) { cpu.implied(); cpu.dex() },
		0xcb: func(cpu *CPU) { cpu.future() },
		0xcc: func(cpu *CPU) { cpu.absolute(); cpu.cpy() },
		0xcd: func(cpu *CPU) { cpu.absolute(); cpu.cmp() },
		0xce: func(cpu *CPU) { cpu.absolute(); cpu.dec() },
		0xcf: func(cpu *CPU) { cpu.future() },
		0xd0: func(cpu *CPU) { cpu.relative(cpu.bne) },
		0xd1: func(cpu *CPU) { cpu.indirectY(); cpu.cmp() },
		0xd2: func(cpu *CPU) { cpu.future() },
		0xd3: func(cpu *CPU) { cpu.future() },
		0xd4: func(cpu *CPU) { cpu.future() },
		0xd5: func(cpu *CPU) { cpu.zpIndexed(); cpu.cmp() },
		0xd6: func(cpu *CPU) { cpu.zpIndexed(); cpu.dec() },
		0xd7: func(cpu *CPU) { cpu.future() },
		0xd8: func(cpu *CPU) { cpu.implied(); cpu.cld() },
		0xd9: func(cpu *CPU) { cpu.modifyAddr(cpu.absIndexed, Y); cpu.cmp() },
		0xda: func(cpu *CPU) { cpu.future() },
		0xdb: func(cpu *CPU) { cpu.future() },
		0xdc: func(cpu *CPU) { cpu.future() },
		0xdd: func(cpu *CPU) { cpu.absIndexed(); cpu.cmp() },
		0xde: func(cpu *CPU) { cpu.modifyAddr(cpu.absIndexed, CP); cpu.dec() },
		0xdf: func(cpu *CPU) { cpu.future() },
		0xe0: func(cpu *CPU) { cpu.immediate(); cpu.cpx() },
		0xe1: func(cpu *CPU) { cpu.indirectX(); cpu.sbc() },
		0xe2: func(cpu *CPU) { cpu.future() },
		0xe3: func(cpu *CPU) { cpu.future() },
		0xe4: func(cpu *CPU) { cpu.zeropage(); cpu.cpx() },
		0xe5: func(cpu *CPU) { cpu.zeropage(); cpu.sbc() },
		0xe6: func(cpu *CPU) { cpu.zeropage(); cpu.inc() },
		0xe7: func(cpu *CPU) { cpu.future() },
		0xe8: func(cpu *CPU) { cpu.implied(); cpu.inx() },
		0xe9: func(cpu *CPU) { cpu.immediate(); cpu.sbc() },
		0xea: func(cpu *CPU) { cpu.implied(); cpu.nop() },
		0xeb: func(cpu *CPU) { cpu.future() },
		0xec: func(cpu *CPU) { cpu.absolute(); cpu.cpx() },
		0xed: func(cpu *CPU) { cpu.absolute(); cpu.sbc() },
		0xee: func(cpu *CPU) { cpu.absolute(); cpu.inc() },
		0xef: func(cpu *CPU) { cpu.future() },
		0xf0: func(cpu *CPU) { cpu.relative(cpu.beq) },
		0xf1: func(cpu *CPU) { cpu.indirectY(); cpu.sbc() },
		0xf2: func(cpu *CPU) { cpu.future() },
		0xf3: func(cpu *CPU) { cpu.future() },
		0xf4: func(cpu *CPU) { cpu.future() },
		0xf5: func(cpu *CPU) { cpu.zpIndexed(); cpu.sbc() },
		0xf6: func(cpu *CPU) { cpu.zpIndexed(); cpu.inc() },
		0xf7: func(cpu *CPU) { cpu.future() },
		0xf8: func(cpu *CPU) { cpu.implied(); cpu.sed() },
		0xf9: func(cpu *CPU) { cpu.modifyAddr(cpu.absIndexed, Y); cpu.sbc() },
		0xfa: func(cpu *CPU) { cpu.future() },
		0xfb: func(cpu *CPU) { cpu.future() },
		0xfc: func(cpu *CPU) { cpu.future() },
		0xfd: func(cpu *CPU) { cpu.absIndexed(); cpu.sbc() },
		0xfe: func(cpu *CPU) { cpu.modifyAddr(cpu.absIndexed, CP); cpu.inc() },
		0xff: func(cpu *CPU) { cpu.future() },
	}
	return cpu
}

func (cpu *CPU) Start() {
	go cpu.instruction()
	go cpu.onTick()
}

func (cpu *CPU) SetBus(bus common.Bus) {
	cpu.bus = bus
}

func (cpu *CPU) onTick() {
	for {
		<-cpu.Tick
		if !cpu.fDMA {
			cpu.innerTick <- true
		}
	}
}

var KEY = false

func (cpu *CPU) read(addr uint16) uint8 {
	// for _, target := range cpu.rising {
	// 	target()
	// }
	val := cpu.bus.Read(addr)
	// if Echo {
	// 	printf("read 0x%x from 0x%x\n", val, addr)
	// }
	// if addr == 0x4016 && KEY {
	// 	fmt.Printf("0x%x\n", val)
	// }

	return val
}

func (cpu *CPU) write(addr uint16, val uint8) {
	// if addr == 0x00 {
	// 	printf("0x%x: write 0x%x to 0x%x\n", cpu.pc, val, addr)
	// }
	// if addr == 0x4016 {
	// 	if val%2 == 0 {
	// 		KEY = true
	// 		fmt.Println("##########")
	// 	} else {
	// 		KEY = false
	// 	}
	// }

	// if addr == 0x2000 {
	// 	if (val & 0x80) > 0 {
	// 		fmt.Println("ENABLE NMI")
	// 	} else {
	// 		fmt.Println("DISABLE NMI")
	// 	}
	// }
	//
	// if addr == 0xe000 {
	// 	fmt.Printf("DISABLE IRQ in 0x%x\n", cpu.pc)
	// }
	// if addr == 0xe001 {
	// 	fmt.Printf("ENABLE IRQ in 0x%x\n", cpu.pc)
	// }

	// if Echo {
	// 	printf("write 0x%x to 0x%x\n", val, addr)
	// }
	// for accumulator-mode
	if addr == 0xffff {
		cpu.a = val
		return
	}
	// dma
	if addr == 0x4014 {
		cpu.fDMA = true
		base := uint16(val) << 8
		for i := base; i <= base+0xff; i++ {
			cpu.bus.Write(0x2004, cpu.bus.Read(i))
			// for _, target := range cpu.falling {
			// 	target()
			// }
		}
		cpu.fDMA = false
		return
	}
	cpu.bus.Write(addr, val)
	// for _, target := range cpu.falling {
	// 	go target()
	// }

}

func (cpu *CPU) pull() uint8 {
	cpu.sp++
	val := cpu.read(makeUint16(0x01, cpu.sp))
	// fmt.Printf("pull 0x%x from 0x%x\n", val, cpu.sp)
	return val
}

func (cpu *CPU) push(val uint8) {
	cpu.write(makeUint16(0x01, cpu.sp), val)
	// fmt.Printf("push 0x%x to 0x%x\n", val, cpu.sp)
	cpu.sp--
}

// nv‑-BdIZc

// see https://wiki.nesdev.com/w/index.php/CPU_status_flag_behavior
func (cpu *CPU) GetP() uint8 {
	p := uint8(0)
	if cpu.C {
		p |= 1
	}
	if cpu.Z {
		p |= 1 << 1
	}
	if cpu.I {
		p |= 1 << 2
	}
	if cpu.D {
		p |= 1 << 3
	}
	p |= 1 << 4
	p |= 1 << 5
	if cpu.V {
		p |= 1 << 6
	}
	if cpu.N {
		p |= 1 << 7
	}
	return p
}

func (cpu *CPU) SetP(p uint8) {
	cpu.C = false
	cpu.Z = false
	cpu.I = false
	cpu.D = false
	cpu.V = false
	cpu.N = false
	if (p & 0x01) > 0 {
		cpu.C = true
	}
	if (p & 0x02) > 0 {
		cpu.Z = true
	}
	if (p & 0x04) > 0 {
		cpu.I = true
	}
	if (p & 0x08) > 0 {
		cpu.D = true
	}
	if (p & 0x40) > 0 {
		cpu.V = true
	}
	if (p & 0x80) > 0 {
		cpu.N = true
	}
}

func (cpu *CPU) instruction() {
	cpu.sp = 0xff
	cpu.skip1 = false
	cpu.waitTick()
	adl := cpu.read(0xfffc)
	cpu.waitTick()
	adh := cpu.read(0xfffd)
	cpu.pc = makeUint16(adh, adl)
	// cpu.pc = 0xc000
	for {
		if cpu.fNMI {
			cpu.nmi()
			continue
		}
		if cpu.fIRQ && !cpu.I {
			cpu.irq()
			continue
		}
		if !cpu.skip1 {
			cpu.waitTick()
		}

		println("new instruction")
		cpu.opcode = cpu.read(cpu.pc)
		cpu.pc++

		cpu.waitTick()
		cpu.data = cpu.read(cpu.pc)
		cpu.pc++

		cpu.skip1 = true

		// 这里好像比 cpu.decode() 然后在函数里面引用cpu.opcode好一点阿
		// next := cpu.decode(cpu.opcode)
		cpu.stat[cpu.opcode]++
		printf("0x%x [0x%x]: ", cpu.pc-2, cpu.opcode)
		next := cpu.decoder[cpu.opcode]
		next(cpu)
		if cpu.fStop {
			cpu.stopped <- true
			return
		}
	}
}

func (cpu *CPU) waitTick() {
	if cpu.test == true {
		return
	}
	<-cpu.innerTick
}

func (cpu *CPU) addr() uint16 {
	return makeUint16(cpu.adh, cpu.adl)
}

// cpu.action
func (cpu *CPU) writeTo() {
	printf(" 0x%x -> 0x%x\n", cpu.buffer, cpu.addr())
	cpu.write(cpu.addr(), cpu.buffer)
}

func (cpu *CPU) readIn() {

	cpu.data = cpu.read(cpu.addr())
	printf(" 0x%x <- 0x%x ", cpu.data, cpu.addr())
}

func (cpu *CPU) pch() uint8 {
	return uint8(cpu.pc >> 8)
}

func (cpu *CPU) pcl() uint8 {
	return uint8(cpu.pc)
}

func (cpu *CPU) getIndex() uint8 {
	if cpu.fY {
		return cpu.y
	}
	return cpu.x
}

const Y int = 0x01
const W int = 0x02
const CP int = 0x04
const CN int = 0x08

func (cpu *CPU) modifyAddr(f func(), flag int) {
	if (flag & W) > 0 {
		cpu.action = cpu.writeTo
	}
	if (flag & Y) > 0 {
		cpu.fY = true
	}
	if (flag & CP) > 0 {
		cpu.fCross = 1
	} else if (flag & CN) > 0 {
		cpu.fCross = Uint8(-1)
	}
	f()
	cpu.fCross = 0
	cpu.fY = false
	cpu.action = cpu.readIn
}

// func (cpu *CPU) AddRising(f func()) {
// 	cpu.rising = append(cpu.rising, f)
// }
//
// func (cpu *CPU) AddFalling(f func()) {
// 	cpu.falling = append(cpu.falling, f)
// }

func (cpu *CPU) IRQ() {
	if !cpu.I {
		cpu.fIRQ = true
	}
}

func (cpu *CPU) NMI() {
	cpu.fNMI = true
}

func (cpu *CPU) reset() {
	panic("TODO")
}
