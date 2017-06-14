package motherboard

import (
	"fmt"
	"time"

	"git.letus.rocks/funsun/peridot/common"
)

type busMapper struct {
	Start, Offset uint16
	Target        common.Bus
	Remap         bool
}

type Router struct {
	mappers []*busMapper
}

func (r *Router) Init() *Router {
	r.mappers = []*busMapper{}
	return r
}

func (r *Router) Read(addr uint16) uint8 {
	mapper := r.findTarget(addr)
	if mapper.Remap {
		return mapper.Target.Read(addr - mapper.Start)
	}
	return mapper.Target.Read(addr)
}

func (r *Router) Write(addr uint16, val uint8) {
	mapper := r.findTarget(addr)
	if mapper.Remap {
		mapper.Target.Write(addr-mapper.Start, val)
		return
	}
	mapper.Target.Write(addr, val)
}

func (r *Router) AddMapping(start, offset uint16, target common.Bus, remap bool) {
	r.mappers = append(r.mappers, &busMapper{start, offset, target, remap})
}

func (r *Router) findTarget(addr uint16) *busMapper {
	for _, mapper := range r.mappers {
		if (mapper.Start <= addr) && (int(addr) < int(mapper.Start)+int(mapper.Offset)) {
			return mapper
		}
	}
	for _, mapper := range r.mappers {
		fmt.Println(mapper)
	}
	panic(fmt.Sprintf("cannot find addr for: 0x%x", addr))
}

type MotherBoard struct {
	CPUBus, PPUBus *Router
	cpu, ppu       chan bool
}

func (mb *MotherBoard) Init() *MotherBoard {
	mb.CPUBus = new(Router).Init()
	mb.PPUBus = new(Router).Init()
	return mb
}

func (mb *MotherBoard) AddCPU(ticker chan bool) {
	mb.cpu = ticker
}

func (mb *MotherBoard) AddPPU(ticker chan bool) {
	mb.ppu = ticker
}

func (mb *MotherBoard) Start() {
	ch := time.Tick(1 * time.Millisecond)
	var accu int64
	var timeCounter int64
	for {
		<-ch
		s := time.Now().UnixNano()
		for i := 0; i < 20040; i++ {
			if i%12 == 0 {
				mb.cpu <- true
			}
			if i%4 == 0 {
				mb.ppu <- true
			}
			// mb.tickAll()
		}
		t := time.Now().UnixNano() - s
		timeCounter++
		accu += t
		if timeCounter%10 == 0 {
			// fmt.Println(accu / 200000)
			accu = 0
		}

	}
}

// func (mb *MotherBoard) tickAll() {
// 	for _, ticker := range mb.tickers {
// 		ticker <- true
// 	}
// }
