package cpu

import (
	"testing"

	"git.letus.rocks/funsun/peridot/common"
)

type testBus struct {
	code map[uint16]uint8
}

func NewTestBus(code map[uint16]uint8) *testBus {
	t := &testBus{}
	t.code = code
	return t
}

func (t *testBus) Write(addr uint16, val uint8) {
	t.code[addr] = val
}

func (t *testBus) Read(addr uint16) uint8 {
	v, ok := t.code[addr]
	if !ok {
		t.code[addr] = 0
		return 0
	}
	return v
}

func prepareCPU(codes []uint8) (*CPU, *testBus) {
	common.Echo = true
	c := new(CPU).Init()
	c.test = true
	codeMap := map[uint16]uint8{}
	for i, code := range codes {
		codeMap[0xe000+uint16(i)] = code
	}
	codeMap[0xe000+uint16(len(codes))] = 0x02
	codeMap[0xfffc] = 0x00
	codeMap[0xfffd] = 0xe0
	tb := NewTestBus(codeMap)
	c.SetBus(tb)
	return c, tb
}

func runCPU(c *CPU) {
	c.Start()
	<-c.stopped
}

const (
	LDA_IMME  = 0xa9
	LDA_ZP    = 0xa5
	LDA_ZP_X  = 0xb5
	LDA_ABS   = 0xad
	LDA_ABS_X = 0xbd
	LDA_ABS_Y = 0xb9
	LDA_IND_X = 0xa1
	LDA_IND_Y = 0xb1
	LSR_AC    = 0x4a
	SEC       = 0x38
	ROR_AC    = 0x6a
)

func TestCPU(t *testing.T) {
	c, _ := prepareCPU([]uint8{
		LDA_IMME,
		0x03,
	})
	c.Start()
	<-c.stopped
	if c.a != 0x03 {
		t.Fail()
	}
}

func TestLSR_AC(t *testing.T) {
	c, _ := prepareCPU([]uint8{LDA_IMME, 0x03, LSR_AC})
	runCPU(c)
	if !c.C || c.a != 0x01 {
		t.Fail()
	}
}

func TestROR_AC(t *testing.T) {
	c, _ := prepareCPU([]uint8{SEC, LDA_IMME, 0x03, ROR_AC})
	runCPU(c)
	if !c.C || c.a != 0x81 {
		t.Fail()
	}
	c, _ = prepareCPU([]uint8{SEC, LDA_IMME, 0x02, ROR_AC})
	runCPU(c)
	if c.C || c.a != 0x81 {
		t.Log(c.C, c.a)
		t.Fail()
	}
}
