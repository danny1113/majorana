package mvp4

import (
	"fmt"

	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type executeUnit struct {
	branchUnit        *simpleBranchUnit
	processing        bool
	pendingMemoryRead bool
	addrs             []int32
	memory            []int8
	remainingCycles   int
	runner            risc.InstructionRunnerPc
	mmu               *memoryManagementUnit
}

func newExecuteUnit(branchUnit *simpleBranchUnit, mmu *memoryManagementUnit) *executeUnit {
	return &executeUnit{branchUnit: branchUnit, mmu: mmu}
}

func (eu *executeUnit) cycle(ctx *risc.Context, app risc.Application, inBus *comp.SimpleBus[risc.InstructionRunnerPc], outBus *comp.SimpleBus[risc.ExecutionContext]) (bool, int32, bool, error) {
	if eu.pendingMemoryRead {
		eu.remainingCycles--
		if eu.remainingCycles != 0 {
			return false, 0, false, nil
		}
		eu.pendingMemoryRead = false
		defer func() {
			eu.runner = risc.InstructionRunnerPc{}
		}()
		var memory []int8
		if eu.memory != nil {
			memory = eu.memory
		} else {
			line := eu.mmu.fetchCacheLine(eu.addrs[0])
			eu.mmu.pushLineToL1D(eu.addrs[0], line)
			m, exists := eu.mmu.getFromL1D(eu.addrs)
			if !exists {
				panic("cache line doesn't exist")
			}
			memory = m
		}
		eu.memory = nil
		return eu.run(ctx, app, outBus, memory)
	}

	if !eu.processing {
		runner, exists := inBus.Get()
		if !exists {
			return false, 0, false, nil
		}
		eu.runner = runner
		eu.remainingCycles = runner.Runner.InstructionType().Cycles()
		eu.processing = true
	}

	eu.remainingCycles--
	if eu.remainingCycles != 0 {
		return false, 0, false, nil
	}

	if !outBus.CanAdd() {
		eu.remainingCycles = 1
		return false, 0, false, nil
	}

	runner := eu.runner
	// Create the branch unit assertions
	eu.branchUnit.assert(runner)

	// To avoid writeback hazard, if the pipeline contains read registers not
	// written yet, we wait for it
	if ctx.IsWriteDataHazard(runner.Runner.ReadRegisters()) {
		eu.remainingCycles = 1
		return false, 0, false, nil
	}

	if ctx.Debug {
		fmt.Printf("\tEU: Executing instruction %d\n", eu.runner.Pc/4)
	}

	addrs := runner.Runner.MemoryRead(ctx)
	if len(addrs) != 0 {
		if m, exists := eu.mmu.getFromL1D(addrs); exists {
			eu.memory = m
			eu.pendingMemoryRead = true
			eu.remainingCycles = cyclesL1Access
		} else {
			eu.addrs = addrs
			eu.pendingMemoryRead = true
			eu.remainingCycles = cyclesMemoryAccess
		}
		return false, 0, false, nil
	}

	defer func() {
		eu.runner = risc.InstructionRunnerPc{}
	}()
	return eu.run(ctx, app, outBus, nil)
}

func (eu *executeUnit) run(ctx *risc.Context, app risc.Application, outBus *comp.SimpleBus[risc.ExecutionContext], memory []int8) (bool, int32, bool, error) {
	execution, err := eu.runner.Runner.Run(ctx, app.Labels, eu.runner.Pc, memory)
	if err != nil {
		return false, 0, false, err
	}
	if execution.Return {
		return false, 0, true, err
	}

	eu.processing = false
	if execution.MemoryChange && eu.mmu.doesExecutionMemoryChangesExistsInL1D(execution) {
		eu.mmu.writeExecutionMemoryChangesToL1D(execution)
		return false, 0, false, nil
	}

	outBus.Add(risc.ExecutionContext{
		Execution:       execution,
		InstructionType: eu.runner.Runner.InstructionType(),
		WriteRegisters:  eu.runner.Runner.WriteRegisters(),
	})
	ctx.AddPendingWriteRegisters(eu.runner.Runner.WriteRegisters())

	if execution.PcChange && eu.branchUnit.shouldFlushPipeline(execution.NextPc) {
		return true, execution.NextPc, false, nil
	}

	return false, 0, false, nil
}

func (eu *executeUnit) isEmpty() bool {
	return !eu.processing
}
