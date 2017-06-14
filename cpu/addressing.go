package cpu

import "fmt"

func (cpu *CPU) future() {
	println("")
	fmt.Printf("error opcde is 0x%x, addr is 0x%x\n", cpu.opcode, cpu.pc-2)
	for k, v := range cpu.stat {
		printf("opcode 0x%x count %d\n", k, v)
	}
	panic("opcode not usable!")
}

func (cpu *CPU) implied() {
	cpu.pc--
}

func (cpu *CPU) immediate() {
	printf("immediate 0x%x -> ", cpu.data)
}

func (cpu *CPU) zeropage() {
	printf("zp: ")
	cpu.waitTick()
	cpu.adh = 0x00
	cpu.adl = cpu.data
	cpu.action()
}

func (cpu *CPU) zpIndexed() {
	printf("zpIndexed ")
	cpu.waitTick()
	cpu.adh = 0x00
	cpu.adl = cpu.data + cpu.getIndex()
	cpu.action()
}

func (cpu *CPU) absolute() {
	printf("absolute: ")
	cpu.adl = cpu.data
	cpu.waitTick()
	cpu.adh = cpu.read(cpu.pc)
	cpu.pc++
	cpu.waitTick()
	cpu.action()
}

// cpu.action和cpu.cross 通过decrator更改
func (cpu *CPU) absIndexed() {
	printf("absIndexed: ")
	cpu.adl = cpu.data
	cpu.waitTick()
	cpu.adh = cpu.read(cpu.pc)
	cpu.pc++
	c := uint8(0)
	cpu.adl, c = addUint8(cpu.adl, cpu.getIndex(), c)

	if (c + cpu.fCross) > 0 {
		cpu.waitTick()
		cpu.adh, c = addUint8(cpu.adh, 0, c)
	}
	cpu.waitTick()
	cpu.action()
}

func (cpu *CPU) indirectX() {
	printf("indirectX: ")
	bal := cpu.data
	cpu.waitTick()
	bal = bal + cpu.x
	cpu.waitTick()
	cpu.adl = cpu.read(makeUint16(0x00, bal))
	bal++
	cpu.waitTick()
	cpu.adh = cpu.read(makeUint16(0x00, bal))
	cpu.waitTick()
	cpu.action()
}

func (cpu *CPU) indirectY() {
	printf("indirectY: ")
	ial := cpu.data
	cpu.waitTick()
	cpu.adl = cpu.read(makeUint16(0x00, ial))
	ial++
	cpu.waitTick()
	cpu.adh = cpu.read(makeUint16(0x00, ial))
	c := uint8(0)
	cpu.adl, c = addUint8(cpu.adl, cpu.y, c)

	// must cross, according c, never cross
	if (c + cpu.fCross) > 0 {
		cpu.waitTick()
		cpu.adh, c = addUint8(cpu.adh, 0, c)
	}
	cpu.waitTick()
	cpu.action()
}

func (cpu *CPU) relative(checker func() bool) {
	offset := cpu.data
	cpu.waitTick()
	if !checker() {
		println(" failed")
		return
	}

	cpu.waitTick()
	// 在waitick之前会被另一个instruction覆盖
	realOffset := int16(Int8(offset))
	res := int16(cpu.pcl()) + realOffset
	cross := int16(0)
	if res < 0 {
		cross = -1
	} else if res > 255 {
		cross = 1
	}
	// cross := res / 256 这样不行，res是负数一样是0
	cpu.adl = uint8(res % 256)
	cpu.adh = cpu.pch()
	if cross != 0 {
		cpu.waitTick()
		cpu.adh = uint8(int16(cpu.adh) + cross)
	}
	cpu.pc = makeUint16(cpu.adh, cpu.adl)
	printf(" off %d to 0x%x\n", Int8(offset), cpu.pc)
	cpu.skip1 = false
}

func (cpu *CPU) accumulator(f func()) {
	cpu.pc--
	cpu.data = cpu.a
	cpu.adh = 0xff
	cpu.adl = 0xff
	printf("A -> ")
	f()
	cpu.skip1 = true
}
