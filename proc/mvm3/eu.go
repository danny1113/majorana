package mvm3

import (
	"fmt"

	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type executeUnit struct {
	branchUnit      *simpleBranchUnit
	processing      bool
	remainingCycles int
	runner          risc.InstructionRunner
	pc              int32
}

func newExecuteUnit(branchUnit *simpleBranchUnit) *executeUnit {
	return &executeUnit{branchUnit: branchUnit}
}

func (eu *executeUnit) cycle(ctx *risc.Context, app risc.Application, inBus *comp.SimpleBus[risc.InstructionRunnerPc], outBus *comp.SimpleBus[comp.ExecutionContext]) (bool, int32, bool, error) {
	if !eu.processing {
		runner, exists := inBus.Get()
		if !exists {
			return false, 0, false, nil
		}
		eu.runner = runner.Runner
		eu.pc = runner.Pc
		eu.remainingCycles = risc.CyclesPerInstruction(runner.Runner.InstructionType())
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

	// To avoid writeback hazard, if the pipeline contains read registers not
	// written yet, we wait for it
	if ctx.ContainWrittenRegisters(runner.ReadRegisters()) {
		eu.remainingCycles = 1
		return false, 0, false, nil
	}

	if ctx.Debug {
		fmt.Printf("\tEU: Executing instruction %d\n", eu.pc/4)
	}
	execution, err := runner.Run(ctx, app.Labels, eu.pc)
	if err != nil {
		return false, 0, false, err
	}
	if execution.Return {
		return false, 0, true, err
	}

	outBus.Add(comp.ExecutionContext{
		Execution:       execution,
		InstructionType: runner.InstructionType(),
		WriteRegisters:  runner.WriteRegisters(),
	})
	ctx.AddWriteRegisters(runner.WriteRegisters())
	eu.runner = nil
	eu.processing = false

	if execution.PcChange && eu.branchUnit.shouldFlushPipeline(execution.NextPc) {
		return true, execution.NextPc, false, nil
	}

	return false, 0, false, nil
}

func (eu *executeUnit) isEmpty() bool {
	return !eu.processing
}
