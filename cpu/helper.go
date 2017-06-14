package cpu

import (
	"fmt"
	"git.letus.rocks/funsun/peridot/common"
)

func makeUint16(h, l uint8) uint16 {
	return uint16(h)<<8 + uint16(l)
}

func addUint8(a, b, carry uint8) (res, c uint8) {
	res, c, _ = addUint8V(a, b, carry)
	return
}

func addUint8V(a, b, carry uint8) (res, c, v uint8) {
	if uint16(a)+uint16(b)+uint16(carry) > 0x00ff {
		c = 1
	}
	res = uint8(a + b + carry)
	signA := int8(a) >= 0
	signB := int8(b) >= 0
	signRes := int8(res) >= 0
	if (signA && signB && !signRes) || (!signA && !signB && !signRes) {
		v = 1
	}
	return
}

func subUint8V(a, b, carry uint8) (res, c, v uint8) {
	res, c, _ = addUint8V(a, ^b, carry)
	signA := int8(a) >= 0
	signB := int8(b) >= 0
	signRes := int8(res) >= 0
	// + - > - || - + > +
	if (signA && !signB && !signRes) || (!signA && signB && signRes) {
		v = 1
	}
	return
}

func Uint8(x int8) uint8 {
	if x >= 0 {
		return uint8(x)
	}
	return uint8(255 + int16(x))
}

func Int8(x uint8) int8 {
	if x > 127 {
		return -int8(255-x) - 1
	}
	return int8(x)
}

func loadFlag(cpu *CPU, val uint8) {
	cpu.Z = (val == 0)
	cpu.N = (val & 0x80) > 0
}

func boolToUint8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

var Echo = false

func println(a ...interface{}) {
	if common.Echo {
		fmt.Println(a...)
	}
}

func printf(format string, a ...interface{}) {
	if common.Echo {
		fmt.Printf(format, a...)
	}
}
