package main

import (
	"fmt"

	"git.letus.rocks/funsun/peridot/cartridge"
	"git.letus.rocks/funsun/peridot/ines"
)

type Disassembler struct {
	position   int
	base       int
	start, end uint16
	code       []uint8
}

func (d *Disassembler) Init(rpg []uint8, start, end uint16, base int) *Disassembler {
	d.code = rpg
	d.base = base
	d.position = base
	d.start = start
	d.end = end
	return d
}

func (d *Disassembler) Run() {
	for {
		d.printOne()
		if uint16(d.position-d.base)+d.start == d.end {
			return
		}
	}
}

func (d *Disassembler) print(s string) {
	fmt.Printf("[%4x] %s\n", uint16(d.position-1-d.base)+d.start, s)
}

func (d *Disassembler) next() uint8 {
	val := d.code[d.position]
	d.position++
	return val
}

func hex8(num uint8) string {
	return fmt.Sprintf("0x%2x", num)
}

func hex16(num uint16) string {
	return fmt.Sprintf("0x%4x", num)
}

func Int8(num uint8) string {
	res := int8(0)
	if num > 127 {
		res = -int8(255-num) - 1
	}
	res = int8(num)
	return fmt.Sprintf("%d", res)
}

func (d *Disassembler) hex8Next() {
	d.print(hex8(d.next()))
}

func (d *Disassembler) printOne() {
	opcode := d.next()
	switch opcode {
	case 0x00:
		d.print("BRK")
	case 0x01:
		d.print("ORA-IND-X")
		d.print(hex8(d.next()))
	case 0x05:
		d.print("ORA-zp")
		d.print(hex8(d.next()))
	case 0x06:
		d.print("asl-zp")
		d.print(hex8(d.next()))
	case 0x08:
		d.print("PHP")
	case 0x09:
		d.print("ORA-imme")
		d.print("#" + hex8(d.next()))
	case 0x0a:
		d.print("ASL-accu")
	case 0x0d:
		d.print("ORA-abs")
		d.print(hex8(d.next()))
		d.print(hex8(d.next()))
	case 0x0e:
		d.print("ASL-abs")
		d.print(hex8(d.next()))
		d.print(hex8(d.next()))
	case 0x10:
		d.print("BPL")
		d.print(Int8(d.next()))
	case 0x11:
		d.print("ORA-IND-Y")
		d.print(hex8(d.next()))
	case 0x15:
		d.print("ORA-ZP-X")
		d.print(hex8(d.next()))
	// case 0x16:
	case 0x18:
		d.print("CLC")
	// case 0x19:
	// case 0x1d:
	// case 0x1e:
	case 0x20:
		d.print("JSR")
		d.print(hex8(d.next()))
		d.print(hex8(d.next()))
	// case 0x21:
	case 0x24:
		d.print("BIT-zp")
		d.hex8Next()
	// case 0x25:
	case 0x26:
		d.print("ROL-zp")
		d.hex8Next()
	case 0x28:
		d.print("PLP")
	case 0x29:
		d.print("AND-IMME")
		d.hex8Next()
	// case 0x2a:
	// case 0x2c:
	// case 0x2d:
	// case 0x2e:
	case 0x30:
		d.print("BMI")
		d.print(Int8(d.next()))
	// case 0x31:
	// case 0x35:
	// case 0x36:
	case 0x38:
		d.print("SEC")
	// case 0x39:
	case 0x3d:
		d.print("AND-ABS-X")
		d.hex8Next()
		d.hex8Next()
	// case 0x3e:
	// case 0x40:
	// case 0x41:
	// case 0x45:
	// case 0x46:
	case 0x48:
		d.print("PHA")
	case 0x49:
		d.print("EOR-IMME")
		d.hex8Next()
	case 0x4a:
		d.print("LSR-accu")
	case 0x4c:
		d.print("JUMP")
		d.hex8Next()
		d.hex8Next()
	// case 0x4d:
	// case 0x4e:
	case 0x50:
		d.print("BVC")
		d.print(Int8(d.next()))
	// case 0x51:
	// case 0x55:
	// case 0x56:
	// case 0x58:
	// case 0x59:
	// case 0x5d:
	// case 0x5e:
	case 0x60:
		d.print("RTS")
	// case 0x61:
	case 0x65:
		d.print("ADC-zp")
		d.hex8Next()
	// case 0x66:
	case 0x68:
		d.print("PLA")
	case 0x69:
		d.print("ADC-IMME")
		d.hex8Next()
	// case 0x6a:
	case 0x6c:
		d.print("JUMP-IND")
		d.hex8Next()
		d.hex8Next()
	// case 0x6d:
	// case 0x6e:
	// case 0x70:
	// case 0x71:
	// case 0x75:
	// case 0x76:
	// case 0x78:
	// case 0x79:
	// case 0x7d:
	// case 0x7e:
	// case 0x81:
	case 0x84:
		d.print("STY-zp")
		d.hex8Next()
	case 0x85:
		d.print("STA-zp")
		d.hex8Next()
	// case 0x86:
	// case 0x88:
	case 0x8a:
		d.print("TXA")
	case 0x8c:
		d.print("STY-ABS")
		d.print(hex8(d.next()))
		d.print(hex8(d.next()))
	case 0x8d:
		d.print("STA-ABS")
		d.print(hex8(d.next()))
		d.print(hex8(d.next()))
	case 0x8e:
		d.print("STX-ABS")
		d.hex8Next()
		d.hex8Next()
	case 0x90:
		d.print("BCC")
		d.print(hex8(d.next()))
	// case 0x91:
	// case 0x94:
	// case 0x95:
	// case 0x96:
	case 0x98:
		d.print("TYA")
	case 0x99:
		d.print("STA-ABS-Y")
		d.hex8Next()
		d.hex8Next()
	// case 0x9a:
	case 0x9d:
		d.print("STA-ABS-X")
		d.hex8Next()
		d.hex8Next()
	case 0xa0:
		d.print("LDY-IMME")
		d.hex8Next()
	// case 0xa1:
	case 0xa2:
		d.print("LDX-IMME")
		d.hex8Next()
	case 0xa4:
		d.print("LDY-zp")
		d.hex8Next()
	case 0xa5:
		d.print("LDA-zp")
		d.print(hex8(d.next()))
	case 0xa6:
		d.print("LDA-zp")
		d.print(hex8(d.next()))
	case 0xa8:
		d.print("TAY")
	case 0xa9:
		d.print("LDA-IMME")
		d.hex8Next()
	case 0xaa:
		d.print("TAX")
	// case 0xab:
	case 0xac:
		d.print("LDY-ABS")
		d.hex8Next()
		d.hex8Next()
	case 0xad:
		d.print("LDA-ABS")
		d.hex8Next()
		d.hex8Next()
	case 0xae:
		d.print("LDX-ABS")
		d.hex8Next()
		d.hex8Next()
	case 0xb0:
		d.print("BCS")
		d.print(hex8(d.next()))
	case 0xb1:
		d.print("LDA-IND-Y")
		d.hex8Next()
	// case 0xb4:
	// case 0xb5:
	// case 0xb6:
	// case 0xb8:
	case 0xb9:
		d.print("LDA-ABS-Y")
		d.hex8Next()
		d.hex8Next()
	// case 0xba:
	// case 0xbc:
	case 0xbd:
		d.print("LDA-ABS-X")
		d.hex8Next()
		d.hex8Next()
	// case 0xbe:
	case 0xc0:
		d.print("CPY-IMME")
		d.hex8Next()
	// case 0xc1:
	// case 0xc4:
	// case 0xc5:
	// case 0xc6:
	case 0xc8:
		d.print("INY")
	case 0xc9:
		d.print("CMP-IMME")
		d.hex8Next()
	// case 0xca:
	// case 0xcc:
	// case 0xcd:
	case 0xce:
		d.print("DEC-ABS")
		d.hex8Next()
		d.hex8Next()
	case 0xd0:
		d.print("BNE")
		d.print(Int8(d.next()))
	// case 0xd1:
	// case 0xd5:
	// case 0xd6:
	case 0xd8:
		d.print("CLD")
	// case 0xd9:
	// case 0xdd:
	// case 0xde:
	case 0xe0:
		d.print("CPX-IMME")
		d.hex8Next()
	// case 0xe1:
	// case 0xe4:
	case 0xe5:
		d.print("SBC-zp")
		d.hex8Next()
	case 0xe6:
		d.print("INC-zp")
		d.hex8Next()
	case 0xe8:
		d.print("INX")
	case 0xe9:
		d.print("SBC-IMME")
		d.hex8Next()
	case 0xea:
		d.print("NOP")
	// case 0xec:
	// case 0xed:
	case 0xee:
		d.print("INC-ABS")
		d.hex8Next()
		d.hex8Next()
	case 0xf0:
		d.print("BEQ")
		d.print(Int8(d.next()))
	// case 0xf1:
	// case 0xf5:
	// case 0xf6:
	// case 0xf8:
	// case 0xf9:
	// case 0xfd:
	// case 0xfe:
	default:
		panic("future: " + hex8(opcode))
	}
}

func main() {
	mmc3 := ines.ReadFile("../../test.nes")
	rpg := mmc3.(*cartridge.MMC3).GetRPG()
	fmt.Println("fffe:ffff ", hex8(rpg[len(rpg)-2]), hex8(rpg[len(rpg)-1]))
	fmt.Println("fffc:fffd ", hex8(rpg[len(rpg)-4]), hex8(rpg[len(rpg)-3]))
	fmt.Println("fffa:fffb ", hex8(rpg[len(rpg)-6]), hex8(rpg[len(rpg)-5]))
	// d := new(Disassembler).Init(rpg, 0xfa67, 0xfffa, len(rpg)-(0xffff-0xfa67+1))
	// d := new(Disassembler).Init(rpg, 0xfe19, 0xfffa, len(rpg)-(0xffff-0xfe19+1))
	d := new(Disassembler).Init(rpg, 0xfdcc, 0xfffa, len(rpg)-(0xffff-0xfdcc+1))

	d.Run()
}
