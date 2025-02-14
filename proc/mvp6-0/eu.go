package mvp6_0

import (
	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type executeUnit struct {
	bu     *btbBranchUnit
	inBus  *comp.BufferedBus[*risc.InstructionRunnerPc]
	outBus *comp.BufferedBus[risc.ExecutionContext]
	mmu    *memoryManagementUnit

	// Pending
	coroutine func(cycle int, ctx *risc.Context, app risc.Application) (bool, int32, int32, bool, error)
	memory    []int8
	runner    risc.InstructionRunnerPc
}

func newExecuteUnit(bu *btbBranchUnit, inBus *comp.BufferedBus[*risc.InstructionRunnerPc], outBus *comp.BufferedBus[risc.ExecutionContext], mmu *memoryManagementUnit) *executeUnit {
	return &executeUnit{
		bu:     bu,
		inBus:  inBus,
		outBus: outBus,
		mmu:    mmu,
	}
}

func (u *executeUnit) cycle(cycle int, ctx *risc.Context, app risc.Application) (bool, int32, int32, bool, error) {
	if u.coroutine != nil {
		return u.coroutine(cycle, ctx, app)
	}

	runner, exists := u.inBus.Get()
	if !exists {
		return false, 0, 0, false, nil
	}
	u.runner = *runner
	u.coroutine = u.coPrepareRun
	return u.coPrepareRun(cycle, ctx, app)
}

func (u *executeUnit) coPrepareRun(cycle int, ctx *risc.Context, app risc.Application) (bool, int32, int32, bool, error) {
	if !u.outBus.CanAdd() {
		log.Infou(ctx, "EU", "can't add")
		return false, 0, 0, false, nil
	}

	// Create the branch unit assertions
	u.bu.assert(u.runner)

	log.Infoi(ctx, "EU", u.runner.Runner.InstructionType(), u.runner.Pc, "executing")

	addrs := u.runner.Runner.MemoryRead(ctx)
	if len(addrs) != 0 {
		if memory, exists := u.mmu.getFromL1D(addrs); exists {
			u.memory = memory
			// As the coroutine is executed the next cycle, if a L1D access takes
			// one cycle, we should be good to go during the next cycle
			remainingCycles := cycleL1DAccess - 1
			u.coroutine = func(cycle int, ctx *risc.Context, app risc.Application) (bool, int32, int32, bool, error) {
				if remainingCycles > 0 {
					remainingCycles--
					return false, 0, 0, false, nil
				}
				return u.coRun(cycle, ctx, app)
			}
			return false, 0, 0, false, nil
		} else {
			remainingCycles := cyclesMemoryAccess - 1

			u.coroutine = func(cycle int, ctx *risc.Context, app risc.Application) (bool, int32, int32, bool, error) {
				if remainingCycles > 0 {
					remainingCycles--
					return false, 0, 0, false, nil
				}
				line := u.mmu.fetchCacheLine(addrs[0])
				u.mmu.pushLineToL1D(addrs[0], line)
				m, exists := u.mmu.getFromL1D(addrs)
				if !exists {
					panic("cache line doesn't exist")
				}
				u.memory = m
				return u.coRun(cycle, ctx, app)
			}
			return false, 0, 0, false, nil
		}
	}
	return u.coRun(cycle, ctx, app)
}

func (u *executeUnit) coRun(cycle int, ctx *risc.Context, app risc.Application) (bool, int32, int32, bool, error) {
	u.coroutine = nil
	execution, err := u.runner.Runner.Run(ctx, app.Labels, u.runner.Pc, u.memory)
	if err != nil {
		return false, 0, 0, false, err
	}
	if execution.Return {
		return false, 0, 0, true, nil
	}

	if execution.MemoryChange && u.mmu.doesExecutionMemoryChangesExistsInL1D(execution) {
		u.mmu.writeExecutionMemoryChangesToL1D(execution)
		ctx.DeletePendingRegisters(u.runner.Runner.ReadRegisters(), u.runner.Runner.WriteRegisters())
		return false, 0, 0, false, nil
	}

	u.outBus.Add(risc.ExecutionContext{
		Pc:              u.runner.Pc,
		Execution:       execution,
		InstructionType: u.runner.Runner.InstructionType(),
		WriteRegisters:  u.runner.Runner.WriteRegisters(),
		ReadRegisters:   u.runner.Runner.ReadRegisters(),
	}, cycle)

	if u.runner.Runner.InstructionType().IsUnconditionalBranch() {
		log.Infoi(ctx, "EU", u.runner.Runner.InstructionType(), u.runner.Pc,
			"notify jump address resolved from %d to %d", u.runner.Pc/4, execution.NextPc/4)
		u.bu.notifyJumpAddressResolved(u.runner.Pc, execution.NextPc)
	}
	if execution.PcChange && u.bu.shouldFlushPipeline(execution.NextPc) {
		log.Infoi(ctx, "EU", u.runner.Runner.InstructionType(), u.runner.Pc,
			"should be a flush")
		return true, u.runner.Pc, execution.NextPc, false, nil
	}

	return false, 0, 0, false, nil
}

func (u *executeUnit) flush() {
	u.coroutine = nil
}

func (u *executeUnit) isEmpty() bool {
	return u.coroutine == nil
}
