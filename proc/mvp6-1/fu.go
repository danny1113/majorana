package mvp6_1

import (
	co "github.com/teivah/majorana/common/coroutine"
	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type fuReq struct {
	cycle int
	app   risc.Application
	ctx   *risc.Context
}

type fetchUnit struct {
	co.Coroutine[fuReq, error]
	pc             int32
	toCleanPending bool
	outBus         *comp.BufferedBus[int32]
	complete       bool
	mmu            *memoryManagementUnit
}

func newFetchUnit(mmu *memoryManagementUnit, outBus *comp.BufferedBus[int32]) *fetchUnit {
	fu := &fetchUnit{
		mmu:    mmu,
		outBus: outBus,
	}
	fu.Coroutine = co.New(fu.start)
	fu.Coroutine.Pre(func(r fuReq) {
		if fu.toCleanPending {
			// The fetch unit may have sent to the bus wrong instruction, we make sure
			// this is not the case by cleaning it
			log.Infou(r.ctx, "FU", "cleaning output bus")
			fu.outBus.Clean()
			fu.toCleanPending = false
		}
	})
	return fu
}

func (u *fetchUnit) start(r fuReq) error {
	for i := 0; i < u.outBus.OutLength(); i++ {
		if !u.outBus.CanAdd() {
			log.Infou(r.ctx, "FU", "can't add")
			return nil
		}

		if _, exists := u.mmu.getFromL1I([]int32{u.pc}); !exists {
			remainingCycles := cyclesMemoryAccess - 1
			u.Checkpoint(func(r fuReq) error {
				if remainingCycles != 0 {
					log.Infou(r.ctx, "FU", "pending memory access")
					remainingCycles--
					return nil
				}
				u.Reset()
				u.mmu.pushLineToL1I(u.pc, make([]int8, l1ICacheLineSize))

				currentPc := u.pc
				u.pc += 4
				if u.pc/4 >= int32(len(r.app.Instructions)) {
					u.Checkpoint(func(fuReq) error { return nil })
					u.complete = true
				}
				log.Infou(r.ctx, "FU", "pushing new element from pc %d", currentPc/4)
				u.outBus.Add(currentPc, r.cycle)
				return nil
			})
			return nil
		}

		currentPc := u.pc
		u.pc += 4
		if u.pc/4 >= int32(len(r.app.Instructions)) {
			u.Checkpoint(func(fuReq) error { return nil })
			u.complete = true
		}
		log.Infou(r.ctx, "FU", "pushing new element from pc %d", currentPc/4)
		u.outBus.Add(currentPc, r.cycle)
	}
	return nil
}

func (u *fetchUnit) reset(pc int32, cleanPending bool) {
	u.Reset()
	u.pc = pc
	u.toCleanPending = cleanPending
}

func (u *fetchUnit) flush(pc int32) {
	u.Reset()
	u.complete = false
	u.pc = pc
}

func (u *fetchUnit) isEmpty() bool {
	return u.complete
}
