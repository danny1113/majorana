package mvm4

import (
	"github.com/teivah/majorana/risc"
)

type btbBranchUnit struct {
	btb         *branchTargetBuffer
	fu          *fetchUnit
	du          *decodeUnit
	toCheck     bool
	expectation int32
}

func newBTBBranchUnit(btbSize int, fu *fetchUnit, du *decodeUnit) *btbBranchUnit {
	return &btbBranchUnit{
		btb: newBranchTargetBuffer(btbSize),
		fu:  fu,
		du:  du,
	}
}

func (bu *btbBranchUnit) assert(runner risc.InstructionRunnerPc) {
	instructionType := runner.Runner.InstructionType()
	if risc.IsUnconditionalBranch(instructionType) {
		nextPc, exists := bu.btb.get(runner.Pc)
		if !exists {
			// Unknown branch, it will lead to a pipeline flush
			bu.toCheck = true
			bu.expectation = -1
		} else {
			// Known branch, no need to check
			bu.toCheck = false
			bu.fu.reset(nextPc, true)
		}
	} else if risc.IsConditionalBranch(instructionType) {
		// Assuming next instruction
		bu.toCheck = true
		bu.expectation = runner.Pc + 4
	} else {
		bu.toCheck = false
	}
}

func (bu *btbBranchUnit) shouldFlushPipeline(pc int32) bool {
	if !bu.toCheck {
		return false
	}
	bu.toCheck = false

	// If the expectation doesn't correspond to the current pc, we made a wrong
	// assumption; therefore, we should flush
	return bu.expectation != pc
}

func (bu *btbBranchUnit) notifyJumpAddressResolved(pc, pcTo int32) {
	bu.btb.add(pc, pcTo)
	bu.fu.reset(pcTo, true)
	bu.du.notifyBranchResolved()
}
