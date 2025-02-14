package mvp6_1

import (
	co "github.com/teivah/majorana/common/coroutine"
	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type euReq struct {
	cycle int
	ctx   *risc.Context
	app   risc.Application
}

type euResp struct {
	flush    bool
	from     int32
	pc       int32
	isReturn bool
	err      error
}

type executeUnit struct {
	co.Coroutine[euReq, euResp]
	bu     *btbBranchUnit
	inBus  *comp.BufferedBus[*risc.InstructionRunnerPc]
	outBus *comp.BufferedBus[risc.ExecutionContext]
	mmu    *memoryManagementUnit

	// Pending
	memory []int8
	runner risc.InstructionRunnerPc
}

func newExecuteUnit(bu *btbBranchUnit, inBus *comp.BufferedBus[*risc.InstructionRunnerPc], outBus *comp.BufferedBus[risc.ExecutionContext], mmu *memoryManagementUnit) *executeUnit {
	eu := &executeUnit{
		bu:     bu,
		inBus:  inBus,
		outBus: outBus,
		mmu:    mmu,
	}
	eu.Coroutine = co.New(eu.start)
	return eu
}

func (u *executeUnit) start(r euReq) euResp {
	runner, exists := u.inBus.Get()
	if !exists {
		return euResp{}
	}
	u.runner = *runner
	return u.ExecuteWithCheckpoint(r, u.prepareRun)
}

func (u *executeUnit) prepareRun(r euReq) euResp {
	if !u.outBus.CanAdd() {
		log.Infou(r.ctx, "EU", "can't add")
		return euResp{}
	}

	if u.runner.Receiver != nil {
		var value int32
		select {
		case v := <-u.runner.Receiver:
			log.Infoi(r.ctx, "EU", u.runner.Runner.InstructionType(), u.runner.Pc, "receive forward register value %d", v)
			value = v
		default:
			return euResp{}
		}

		u.runner.Runner.Forward(risc.Forward{Value: value, Register: u.runner.ForwardRegister})
	}

	// Create the branch unit assertions
	u.bu.assert(u.runner)

	log.Infoi(r.ctx, "EU", u.runner.Runner.InstructionType(), u.runner.Pc, "executing")

	addrs := u.runner.Runner.MemoryRead(r.ctx)
	if len(addrs) != 0 {
		if memory, exists := u.mmu.getFromL1D(addrs); exists {
			u.memory = memory
			// As the coroutine is executed the next cycle, if a L1D access takes
			// one cycle, we should be good to go during the next cycle
			remainingCycles := cycleL1DAccess - 1

			u.Checkpoint(func(r euReq) euResp {
				if remainingCycles > 0 {
					remainingCycles--
					return euResp{}
				}
				return u.run(r)
			})
			return euResp{}
		} else {
			remainingCycles := cyclesMemoryAccess - 1

			u.Checkpoint(func(r euReq) euResp {
				if remainingCycles > 0 {
					log.Infoi(r.ctx, "EU", u.runner.Runner.InstructionType(), u.runner.Pc, "pending memory access %d", remainingCycles)
					remainingCycles--
					return euResp{}
				}
				line := u.mmu.fetchCacheLine(addrs[0])
				u.mmu.pushLineToL1D(addrs[0], line)
				m, exists := u.mmu.getFromL1D(addrs)
				if !exists {
					panic("cache line doesn't exist")
				}
				u.memory = m
				return u.run(r)
			})
			return euResp{}
		}
	}
	return u.run(r)
}

func (u *executeUnit) run(r euReq) euResp {
	u.Reset()
	execution, err := u.runner.Runner.Run(r.ctx, r.app.Labels, u.runner.Pc, u.memory)
	if err != nil {
		return euResp{err: err}
	}
	log.Infoi(r.ctx, "EU", u.runner.Runner.InstructionType(), u.runner.Pc, "execution result: %+v", execution)
	if execution.Return {
		return euResp{isReturn: true}
	}

	if execution.MemoryChange && u.mmu.doesExecutionMemoryChangesExistsInL1D(execution) {
		u.mmu.writeExecutionMemoryChangesToL1D(execution)
		r.ctx.DeletePendingRegisters(u.runner.Runner.ReadRegisters(), u.runner.Runner.WriteRegisters())
		return euResp{}
	}

	u.outBus.Add(risc.ExecutionContext{
		Pc:              u.runner.Pc,
		Execution:       execution,
		InstructionType: u.runner.Runner.InstructionType(),
		WriteRegisters:  u.runner.Runner.WriteRegisters(),
		ReadRegisters:   u.runner.Runner.ReadRegisters(),
	}, r.cycle)

	if u.runner.Forwarder == nil {
		if u.runner.Runner.InstructionType().IsUnconditionalBranch() {
			log.Infoi(r.ctx, "EU", u.runner.Runner.InstructionType(), u.runner.Pc,
				"notify jump address resolved from %d to %d", u.runner.Pc/4, execution.NextPc/4)
			u.bu.notifyJumpAddressResolved(u.runner.Pc, execution.NextPc)
		}
		if execution.PcChange && u.bu.shouldFlushPipeline(execution.NextPc) {
			log.Infoi(r.ctx, "EU", u.runner.Runner.InstructionType(), u.runner.Pc, "should be a flush")
			return euResp{flush: true, from: u.runner.Pc, pc: execution.NextPc}
		}
	} else {
		u.runner.Forwarder <- execution.RegisterValue
		log.Infoi(r.ctx, "EU", u.runner.Runner.InstructionType(), u.runner.Pc, "forward register value %d", execution.RegisterValue)
		if u.runner.Runner.InstructionType().IsBranch() {
			panic("shouldn't be a branch")
		}
	}

	return euResp{}
}

func (u *executeUnit) flush() {
	u.Reset()
}

func (u *executeUnit) isEmpty() bool {
	return u.IsStart()
}
