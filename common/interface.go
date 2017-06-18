package common

import (
	"fmt"
	"image"
)

type Bus interface {
	Read(uint16) uint8
	Write(uint16, uint8)
}

type Ticker interface {
	OnTick()
}

type Router interface {
	AddMapping(uint16, uint16, Bus, bool)
}

type Cartridge interface {
	SetCPURouter(Router)
	SetPPURouter(Router)
	IRQ()
	SetIRQ(func())
}

type ComplexBus interface {
	PPURead(uint16) uint8
	PPUWrite(uint16, uint8)
	CPURead(uint16) uint8
	CPUWrite(uint16, uint8)
}

type Screen interface {
	AddFrameBuffer(image.Image)
}

func Hex(a int) string {
	return fmt.Sprintf("%x", a)
}

var Echo = false

var Terminate = make(chan bool)
