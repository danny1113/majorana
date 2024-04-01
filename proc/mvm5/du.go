package mvm5

import (
	"fmt"

	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type decodeUnit struct {
	pendingBranchResolution bool
	inBus                   *comp.BufferedBus[int32]
	outBus                  *comp.BufferedBus[risc.InstructionRunnerPc]
}

func newDecodeUnit(inBus *comp.BufferedBus[int32], outBus *comp.BufferedBus[risc.InstructionRunnerPc]) *decodeUnit {
	return &decodeUnit{inBus: inBus, outBus: outBus}
}

func (u *decodeUnit) cycle(cycle int, app risc.Application, ctx *risc.Context) {
	if u.pendingBranchResolution {
		return
	}

	for i := 0; i < u.outBus.InLength(); i++ {
		if !u.outBus.CanAdd() {
			return
		}
		pc, exists := u.inBus.Get()
		if !exists {
			return
		}
		if ctx.Debug {
			fmt.Printf("\tDU: Decoding instruction %d\n", pc/4)
		}
		runner := app.Instructions[pc/4]
		if risc.IsJump(runner.InstructionType()) {
			u.pendingBranchResolution = true
		}
		u.outBus.Add(risc.InstructionRunnerPc{
			Runner: runner,
			Pc:     pc,
		}, cycle)
	}
}

func (u *decodeUnit) notifyBranchResolved() {
	u.pendingBranchResolution = false
}

func (u *decodeUnit) flush() {
	u.pendingBranchResolution = false
}

func (u *decodeUnit) isEmpty() bool {
	// As the decode unit takes only one cycle, it is considered as empty by default
	return true
}
