package mvp5_1

import (
	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type executeUnit struct {
	pending           bool
	remainingCycles   int
	pendingMemoryRead bool
	addrs             []int32
	runner            risc.InstructionRunnerPc
	bu                *btbBranchUnit
	inBus             *comp.BufferedBus[*risc.InstructionRunnerPc]
	outBus            *comp.BufferedBus[risc.ExecutionContext]
}

func newExecuteUnit(bu *btbBranchUnit, inBus *comp.BufferedBus[*risc.InstructionRunnerPc], outBus *comp.BufferedBus[risc.ExecutionContext]) *executeUnit {
	return &executeUnit{bu: bu, inBus: inBus, outBus: outBus}
}

func (u *executeUnit) cycle(cycle int, ctx *risc.Context, app risc.Application) (bool, int32, bool, error) {
	if u.pendingMemoryRead {
		u.remainingCycles--
		if u.remainingCycles != 0 {
			return false, 0, false, nil
		}
		u.pendingMemoryRead = false
		defer func() {
			u.runner = risc.InstructionRunnerPc{}
		}()
		var memory []int8
		for _, addr := range u.addrs {
			memory = append(memory, ctx.Memory[addr])
		}
		return u.run(cycle, ctx, app, memory)
	}

	if !u.pending {
		runner, exists := u.inBus.Get()
		if !exists {
			return false, 0, false, nil
		}
		u.runner = *runner
		u.remainingCycles = runner.Runner.InstructionType().Cycles()
		u.pending = true
	}

	u.remainingCycles--
	if u.remainingCycles != 0 {
		log.Infou(ctx, "EU", "pending remaining cycles %d (pc=%d, ins=%s)", u.remainingCycles, u.runner.Pc/4, u.runner.Runner.InstructionType())
		return false, 0, false, nil
	}

	if !u.outBus.CanAdd() {
		u.remainingCycles = 1
		log.Infou(ctx, "EU", "can't add")
		return false, 0, false, nil
	}

	runner := u.runner

	if runner.Receiver != nil {
		var value int32
		select {
		case v := <-runner.Receiver:
			value = v
		default:
			// Not yet ready
			u.pending = true
			u.remainingCycles = 1
			return false, 0, false, nil
		}

		runner.Runner.Forward(risc.Forward{Value: value, Register: runner.ForwardRegister})
	}

	// Create the branch unit assertions
	u.bu.assert(runner)

	log.Infoi(ctx, "EU", runner.Runner.InstructionType(), runner.Pc, "executing")

	addrs := runner.Runner.MemoryRead(ctx)
	if len(addrs) != 0 {
		u.addrs = addrs
		u.pendingMemoryRead = true
		u.remainingCycles = cyclesMemoryAccess
		return false, 0, false, nil
	}

	defer func() {
		u.runner = risc.InstructionRunnerPc{}
	}()
	return u.run(cycle, ctx, app, nil)
}

func (u *executeUnit) run(cycle int, ctx *risc.Context, app risc.Application, memory []int8) (bool, int32, bool, error) {
	execution, err := u.runner.Runner.Run(ctx, app.Labels, u.runner.Pc, memory)
	if err != nil {
		return false, 0, false, err
	}
	if execution.Return {
		return false, 0, true, nil
	}

	u.outBus.Add(risc.ExecutionContext{
		Execution:       execution,
		InstructionType: u.runner.Runner.InstructionType(),
		WriteRegisters:  u.runner.Runner.WriteRegisters(),
		ReadRegisters:   u.runner.Runner.ReadRegisters(),
	}, cycle)
	u.pending = false

	if u.runner.Forwarder == nil {
		if u.runner.Runner.InstructionType().IsUnconditionalBranch() {
			log.Infoi(ctx, "EU", u.runner.Runner.InstructionType(), u.runner.Pc,
				"notify jump address resolved from %d to %d", u.runner.Pc/4, execution.NextPc/4)
			u.bu.notifyJumpAddressResolved(u.runner.Pc, execution.NextPc)
		}
		if execution.PcChange && u.bu.shouldFlushPipeline(execution.NextPc) {
			log.Infoi(ctx, "EU", u.runner.Runner.InstructionType(), u.runner.Pc,
				"should be a flush")
			return true, execution.NextPc, false, nil
		}
	} else {
		u.runner.Forwarder <- execution.RegisterValue
		if u.runner.Runner.InstructionType().IsBranch() {
			panic("shouldn't be a branch")
		}
	}

	return false, 0, false, nil
}

func (u *executeUnit) flush() {
	u.pending = false
	u.remainingCycles = 0
}

func (u *executeUnit) isEmpty() bool {
	return !u.pending
}
