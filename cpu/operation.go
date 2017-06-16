package cpu

func (cpu *CPU) clc() {
	println("CLC")
	cpu.waitTick()
	cpu.C = false
}

func (cpu *CPU) cld() {
	println("CLD")
	cpu.waitTick()
	cpu.D = false
}

func (cpu *CPU) cli() {
	println("CLI")
	cpu.waitTick()
	cpu.I = false
}

func (cpu *CPU) clv() {
	println("CLV")
	cpu.waitTick()
	cpu.V = false
}

func (cpu *CPU) sec() {
	println("SEC")
	cpu.waitTick()
	cpu.C = true
}

func (cpu *CPU) sed() {
	println("SED")
	cpu.waitTick()
	cpu.D = true
}

func (cpu *CPU) sei() {
	println("SEI")
	cpu.waitTick()
	cpu.I = true
}

func (cpu *CPU) tax() {
	println("TAX")
	cpu.waitTick()
	// 这样写让tax和tay只差一个symbol
	cpu.x = cpu.a
	loadFlag(cpu, cpu.a)
}

func (cpu *CPU) tay() {
	println("TAY")
	cpu.waitTick()
	cpu.y = cpu.a
	loadFlag(cpu, cpu.a)
}

func (cpu *CPU) tya() {
	println("TYA")
	cpu.waitTick()
	cpu.a = cpu.y
	loadFlag(cpu, cpu.y)
}

func (cpu *CPU) tsx() {
	println("TSX")
	cpu.waitTick()
	cpu.x = cpu.sp
	loadFlag(cpu, cpu.sp)
}

func (cpu *CPU) txa() {
	println("TXA")
	cpu.waitTick()
	cpu.a = cpu.x
	loadFlag(cpu, cpu.x)
}

func (cpu *CPU) txs() {
	println("TXS")
	cpu.waitTick()
	// fmt.Printf("TXS 0x%v to 0x%v", cpu.x, cpu.sp)
	cpu.sp = cpu.x
}

func (cpu *CPU) inx() {
	println("INX")
	cpu.waitTick()
	cpu.x += 1
	loadFlag(cpu, cpu.x)
}

func (cpu *CPU) iny() {
	println("INY")
	cpu.waitTick()
	cpu.y += 1
	loadFlag(cpu, cpu.y)
}

func (cpu *CPU) dex() {
	println("DEX")
	cpu.waitTick()
	cpu.x -= 1
	loadFlag(cpu, cpu.x)
}

func (cpu *CPU) dey() {
	println("DEY")
	cpu.waitTick()
	cpu.y -= 1
	loadFlag(cpu, cpu.y)
}

func (cpu *CPU) nop() {
	println("NOP")
	cpu.waitTick()
}

func (cpu *CPU) brk() {
	println("BRK")
	cpu.waitTick()
	cpu.pc += 2
	cpu.push(cpu.pch())
	cpu.waitTick()
	cpu.push(cpu.pcl())
	cpu.waitTick()
	// thereis no b-flag , see https://wiki.nesdev.com/w/index.php/CPU_status_flag_behavior
	// cpu.B = true
	cpu.push(cpu.GetP())
	cpu.waitTick()
	adl := cpu.read(0xfffe)
	cpu.waitTick()
	adh := cpu.read(0xffff)
	cpu.pc = makeUint16(adh, adl)
	printf("BRK -> 0x%x\n", cpu.pc)
	cpu.skip1 = false
}

func (cpu *CPU) pha() {
	println("PHA")
	cpu.waitTick()
	cpu.push(cpu.a)
	cpu.skip1 = false
}

func (cpu *CPU) pla() {
	println("PLA")
	cpu.waitTick()
	cpu.waitTick()
	m := cpu.pull()
	cpu.waitTick()
	cpu.a = m
	loadFlag(cpu, cpu.a)

}

func (cpu *CPU) php() {
	println("PHP")
	cpu.waitTick()
	cpu.push(cpu.GetP())
	cpu.skip1 = false
}

func (cpu *CPU) plp() {
	println("PLP")
	cpu.waitTick()
	cpu.waitTick()
	m := cpu.pull()
	cpu.waitTick()
	cpu.SetP(m)
}

func (cpu *CPU) rts() {
	pc := cpu.pc

	cpu.waitTick()
	cpu.waitTick()
	pcl := cpu.pull()
	cpu.waitTick()
	pch := cpu.pull()
	cpu.pc = makeUint16(pch, pcl)
	cpu.waitTick()
	cpu.pc++
	// fmt.Printf("RTS from 0x%x to 0x%x, stack is 0x%x\n", pc, cpu.pc, cpu.sp)
	printf("RTS from 0x%x to 0x%x\n", pc, cpu.pc)
	cpu.skip1 = false
}

func (cpu *CPU) rti() {
	// pc := cpu.pc
	cpu.waitTick()
	cpu.waitTick()
	cpu.SetP(cpu.pull())
	cpu.waitTick()
	pcl := cpu.pull()
	cpu.waitTick()
	pch := cpu.pull()
	cpu.pc = makeUint16(pch, pcl)
	// fmt.Printf("RTI from 0x%x to 0x%x\n", pc, cpu.pc)
	// fmt.Printf("0x%x 0x%x\n", cpu.pc, cpu.sp)
	// for i := cpu.sp; i != 0; i++ {
	// 	fmt.Printf("val 0x%x sp 0x%x\n", cpu.read(makeUint16(0x01, cpu.sp+i)), cpu.sp+i)
	// }

	Echo = false
	cpu.skip1 = false
}

func (cpu *CPU) bcs() bool {
	printf("BCS")
	return cpu.C
}

func (cpu *CPU) bcc() bool {
	printf("BCC")
	return !cpu.C
}

func (cpu *CPU) beq() bool {
	printf("BEQ")
	return cpu.Z
}

func (cpu *CPU) bne() bool {
	printf("BNE")
	return !cpu.Z
}

func (cpu *CPU) bmi() bool {
	printf("BMI")
	return cpu.N
}

func (cpu *CPU) bpl() bool {
	printf("BPL")
	return !cpu.N
}

func (cpu *CPU) bvs() bool {
	printf("BVS")
	return cpu.V
}

func (cpu *CPU) bvc() bool {
	printf("BVC")
	return !cpu.V
}

func (cpu *CPU) inc() {
	println("INC")
	cpu.waitTick()
	cpu.data++
	loadFlag(cpu, cpu.data)
	cpu.waitTick()
	cpu.write(cpu.addr(), cpu.data)
	cpu.skip1 = false
}

func (cpu *CPU) dec() {
	println("DEC")
	cpu.waitTick()
	cpu.data--
	loadFlag(cpu, cpu.data)
	cpu.waitTick()
	cpu.write(cpu.addr(), cpu.data)
	cpu.skip1 = false
}

func (cpu *CPU) lsr() {
	println("LSR")
	cpu.waitTick()
	cpu.C = (cpu.data & 0x01) == 0x01
	cpu.data = cpu.data >> 1
	loadFlag(cpu, cpu.data)
	cpu.waitTick()
	cpu.write(cpu.addr(), cpu.data)
	cpu.skip1 = false
}

func (cpu *CPU) asl() {
	println("ASL")
	cpu.waitTick()
	cpu.C = (cpu.data & 0x80) == 0x80
	cpu.data = cpu.data << 1
	loadFlag(cpu, cpu.data)
	cpu.waitTick()
	cpu.write(cpu.addr(), cpu.data)
	cpu.skip1 = false
}

func (cpu *CPU) rol() {
	println("ROL")
	cpu.waitTick()
	cbit := boolToUint8(cpu.C)
	cpu.C = (cpu.data & 0x80) == 0x80
	cpu.data = cpu.data<<1 + uint8(cbit)
	loadFlag(cpu, cpu.data)
	cpu.waitTick()
	cpu.write(cpu.addr(), cpu.data)
	cpu.skip1 = false
}

func (cpu *CPU) ror() {
	println("ROR")
	cpu.waitTick()
	cbit := boolToUint8(cpu.C)
	cpu.C = (cpu.data & 0x01) == 0x01
	cpu.data = (cpu.data >> 1) + (uint8(cbit) << 7)
	loadFlag(cpu, cpu.data)
	cpu.waitTick()
	cpu.write(cpu.addr(), cpu.data)
	cpu.skip1 = false
}

func (cpu *CPU) lda() {
	println("LDA")
	cpu.waitTick()
	cpu.a = cpu.data
	loadFlag(cpu, cpu.a)
}

func (cpu *CPU) ldx() {
	println("LDX")
	cpu.waitTick()
	cpu.x = cpu.data
	loadFlag(cpu, cpu.x)
}

func (cpu *CPU) ldy() {
	println("LDY")
	cpu.waitTick()
	cpu.y = cpu.data
	loadFlag(cpu, cpu.y)
}

func (cpu *CPU) and() {
	println("AND")
	cpu.waitTick()
	cpu.a = cpu.data & cpu.a
	loadFlag(cpu, cpu.a)
}

func (cpu *CPU) ora() {
	println("ORA")
	cpu.waitTick()
	cpu.a = cpu.data | cpu.a
	loadFlag(cpu, cpu.a)
}

func (cpu *CPU) cmp() {
	println("CMP")
	cpu.waitTick()
	res, c, _ := subUint8V(cpu.a, cpu.data, 1)
	// printf("cmp:  0x%x - 0x%x = 0x%x, c is %v", cpu.a, cpu.data, res, c)
	cpu.C = c > 0
	loadFlag(cpu, res)
}

func (cpu *CPU) cpx() {
	println("CPX")
	cpu.waitTick()
	res, c, _ := subUint8V(cpu.x, cpu.data, 1)
	cpu.C = c > 0
	loadFlag(cpu, res)
}

func (cpu *CPU) cpy() {
	println("CPY")
	cpu.waitTick()
	res, c, _ := subUint8V(cpu.y, cpu.data, 1)
	cpu.C = c > 0
	loadFlag(cpu, res)
}

// 存的是jsr第三个字节的地址，所以返回的时候要+1的到下一个指令的地址
func (cpu *CPU) jsr() {
	pc := cpu.pc
	adl := cpu.data
	cpu.waitTick()
	cpu.waitTick()
	cpu.push(cpu.pch())
	cpu.waitTick()
	cpu.push(cpu.pcl())
	cpu.waitTick()
	adh := cpu.read(cpu.pc)
	cpu.pc++
	cpu.pc = makeUint16(adh, adl)
	// fmt.Printf("JSR from 0x%x to 0x%x, sp is 0x%x\n", pc, cpu.pc, cpu.sp)
	printf("JSR from 0x%x to 0x%x\n", pc, cpu.pc)

	cpu.skip1 = false
}

//
func (cpu *CPU) jump() {
	pc := cpu.pc
	adl := cpu.data
	cpu.waitTick()
	adh := cpu.read(cpu.pc)
	cpu.pc = makeUint16(adh, adl)
	printf("JUMP from 0x%x to 0x%x\n", pc, cpu.pc)
	cpu.skip1 = false
}

func (cpu *CPU) jumpIndirect() {
	pc := cpu.pc
	bal := cpu.data
	cpu.waitTick()
	bah := cpu.read(cpu.pc)
	cpu.waitTick()
	adl := cpu.read(makeUint16(bah, bal))
	cpu.waitTick()
	adh := cpu.read(makeUint16(bah, bal+1))
	cpu.pc = makeUint16(adh, adl)
	printf("JUMPIND with 0x%x from 0x%x to 0x%x\n", makeUint16(bah, bal), pc, cpu.pc)
	// fmt.Printf("JUMPIND with 0x%x from 0x%x to 0x%x\n", makeUint16(bah, bal), pc, cpu.pc)
	cpu.skip1 = false
}

func (cpu *CPU) adc() {
	println("ADC")
	cpu.waitTick()
	res, c, v := addUint8V(cpu.a, cpu.data, boolToUint8(cpu.C))
	// printf("adc: prec is %v, 0x%x + 0x%x = 0x%x, c is %v, v is %v\n", cpu.C, cpu.a, cpu.data, res, c, v)
	cpu.a = res
	cpu.C = c > 0
	cpu.V = v > 0
	loadFlag(cpu, cpu.a)
}

func (cpu *CPU) sbc() {
	println("SBC")
	cpu.waitTick()
	res, c, v := subUint8V(cpu.a, cpu.data, boolToUint8(cpu.C))
	// printf("sbc: prec is %v,  0x%x - 0x%x = 0x%x, c is %v, v is %v\n", cpu.C, cpu.a, cpu.data, res, c, v)
	cpu.a = res
	cpu.C = c > 0
	cpu.V = v > 0
	loadFlag(cpu, cpu.a)
}

func (cpu *CPU) bit() {
	println("BIT")
	cpu.waitTick()
	cpu.Z = (cpu.data & cpu.a) == 0
	cpu.N = ((cpu.data & 0x80) >> 7) > 0
	cpu.V = ((cpu.data & 0x40) >> 6) > 0
}

func (cpu *CPU) eor() {
	println("EOR")
	cpu.waitTick()
	cpu.a = (cpu.a | cpu.data) & ^(cpu.a & cpu.data)
	loadFlag(cpu, cpu.a)
}

func (cpu *CPU) sta() {
	printf("STA ")
	cpu.buffer = cpu.a
}

func (cpu *CPU) stx() {
	printf("STX ")
	cpu.buffer = cpu.x
}

func (cpu *CPU) sty() {
	printf("STY ")
	cpu.buffer = cpu.y
}

var irqCounter = 0

func (cpu *CPU) irq() {
	// Echo = true
	// if irqCounter < 100 {
	// 	irqCounter++
	// 	Echo = true
	// }
	// common.Echo = false
	// fmt.Println("IRQ##############################")
	// println("IRQ")
	// fmt.Println("IRQ")
	// fmt.Printf("0x%x 0x%x\n", cpu.pc, cpu.sp)
	cpu.fIRQ = false
	cpu.waitTick()
	cpu.waitTick()
	cpu.push(cpu.pch())
	cpu.waitTick()
	cpu.push(cpu.pcl())
	cpu.waitTick()
	cpu.push(cpu.GetP() & 0xef)
	cpu.waitTick()
	adl := cpu.read(0xfffe)
	cpu.waitTick()
	adh := cpu.read(0xffff)
	cpu.pc = makeUint16(adh, adl)
	cpu.I = true
	cpu.skip1 = false
}

func (cpu *CPU) nmi() {
	// common.Echo = true
	// println("NMI!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	// fmt.Println("NMI")
	// fmt.Printf("0x%x 0x%x\n", cpu.pc, cpu.sp)
	// for i := cpu.sp; i != 0; i++ {
	// 	fmt.Printf("val 0x%x sp 0x%x\n", cpu.read(makeUint16(0x01, cpu.sp+i)), cpu.sp+i)
	// }
	cpu.fNMI = false
	cpu.waitTick()
	cpu.waitTick()
	cpu.push(cpu.pch())
	cpu.waitTick()
	cpu.push(cpu.pcl())
	cpu.waitTick()
	cpu.push(cpu.GetP() & 0xef)
	cpu.waitTick()
	adl := cpu.read(0xfffa)
	cpu.waitTick()
	adh := cpu.read(0xfffb)
	cpu.pc = makeUint16(adh, adl)
	cpu.I = true
	cpu.skip1 = false
}
