package controller

// JOYPAD1 = $4016
// JOYPAD2 = $4017
// bit:   	 7     6     5     4     3     2     1     0
// button:	 A     B  Select Start  Up   Down  Left  Right

type Controller struct {
	status  uint8
	counter uint8
	fReload bool
}

func (c *Controller) Init() *Controller {
	c.counter = 7
	return c
}

func (c *Controller) Read(addr uint16) uint8 {
	// ignore $4017
	if addr == 0x01 {
		return 0
	}
	val := (c.status & (1 << c.counter)) >> (c.counter)
	if c.fReload {
		c.counter = 7
	} else {
		c.counter = (c.counter - 1) % 8
	}
	return val
}

func (c *Controller) Write(addr uint16, val uint8) {
	// val > 0 也可能是disable reload, 因为看的是最后一位，也就是奇偶
	if val%2 == 1 {
		c.fReload = true
	} else {
		c.fReload = false
	}
}

const (
	Right = 1 << iota
	Left
	Down
	Up
	Start
	Select
	B
	A
)

func (c *Controller) SetButton(button uint8) {
	// fmt.Printf("Set 0b%b\n", button)
	c.status |= button
	// fmt.Printf("STATUS NOW 0b%b\n", c.status)
}

func (c *Controller) ClearButton(button uint8) {
	// fmt.Printf("Clear 0b%b\n", button)
	c.status &= (^button)
	// fmt.Printf("STATUS NOW 0b%b\n", c.status)
}
