package mvp6_1

import (
	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/common/obs"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

const (
	pendingLength = 10
)

type controlUnit struct {
	inBus                        *comp.BufferedBus[risc.InstructionRunnerPc]
	outBus                       *comp.BufferedBus[*risc.InstructionRunnerPc]
	pendings                     *comp.Queue[risc.InstructionRunnerPc]
	pushedRunnersInPreviousCycle map[*risc.InstructionRunnerPc]bool
	pushedRunnersInCurrentCycle  map[*risc.InstructionRunnerPc]bool
	skippedInCurrentCycle        []risc.InstructionRunnerPc

	pushed            *obs.Gauge
	pending           *obs.Gauge
	pendingRead       *obs.Gauge
	blocked           *obs.Gauge
	forwarding        int
	total             int
	cantAdd           int
	blockedBranch     int
	blockedDataHazard int
	routeSecond       bool
}

func newControlUnit(inBus *comp.BufferedBus[risc.InstructionRunnerPc], outBus *comp.BufferedBus[*risc.InstructionRunnerPc]) *controlUnit {
	return &controlUnit{
		inBus:                        inBus,
		outBus:                       outBus,
		pendings:                     comp.NewQueue[risc.InstructionRunnerPc](pendingLength),
		pushed:                       &obs.Gauge{},
		pending:                      &obs.Gauge{},
		pendingRead:                  &obs.Gauge{},
		blocked:                      &obs.Gauge{},
		pushedRunnersInCurrentCycle:  make(map[*risc.InstructionRunnerPc]bool),
		pushedRunnersInPreviousCycle: make(map[*risc.InstructionRunnerPc]bool),
	}
}

func (u *controlUnit) cycle(cycle int, ctx *risc.Context) {
	pushedCount := 0
	u.pushedRunnersInCurrentCycle = make(map[*risc.InstructionRunnerPc]bool)
	defer func() {
		u.pushed.Push(pushedCount)
		u.pending.Push(u.pendings.Length())
		u.pushedRunnersInPreviousCycle = u.pushedRunnersInCurrentCycle
	}()
	u.skippedInCurrentCycle = nil
	u.pendingRead.Push(u.inBus.PendingRead())
	if u.inBus.CanGet() {
		u.blocked.Push(1)
	} else {
		u.blocked.Push(0)
	}
	u.total++

	if !u.outBus.CanAdd() {
		u.cantAdd++
		log.Infou(ctx, "CU", "can't add")
		return
	}

	//remaining := u.outBus.RemainingToAdd()
	for elem := range u.pendings.Iterator() {
		runner := u.pendings.Value(elem)

		push, stop := u.handleRunner(ctx, cycle, pushedCount, &runner)
		if push {
			u.pendings.Remove(elem)
			pushedCount++
		} else {
			u.skippedInCurrentCycle = append(u.skippedInCurrentCycle, runner)
		}
		if stop {
			return
		}
	}

	for !u.pendings.IsFull() {
		runner, exists := u.inBus.Get()
		if !exists {
			return
		}

		push, stop := u.handleRunner(ctx, cycle, pushedCount, &runner)
		if push {
			pushedCount++
		} else {
			u.pendings.Push(runner)
			u.skippedInCurrentCycle = append(u.skippedInCurrentCycle, runner)
		}
		if stop {
			return
		}
	}
}

func (u *controlUnit) handleRunner(ctx *risc.Context, cycle int, pushedCount int, runner *risc.InstructionRunnerPc) (push, stop bool) {
	if pushedCount > 0 && runner.Runner.InstructionType().IsBranch() {
		u.blockedBranch++
		return false, true
	}

	if u.isDataHazardWithSkippedRunners(runner) {
		log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "hazard with skipped runner")
		return false, false
	}

	hazards, hazardTypes := ctx.IsDataHazard3(runner.Runner)
	if len(hazards) == 0 {
		pushed := u.pushRunner(ctx, cycle, runner)
		if !pushed {
			return false, true
		}
		u.pushedRunnersInCurrentCycle[runner] = true
		return true, false
	}

	if u.isDataHazardWithSkippedRunners(runner) {
		log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "data hazard with skipped runners")
		return false, false
	}

	if should, previousRunner, register := u.shouldUseForwarding(runner, hazards, hazardTypes); should {
		ch := make(chan int32, 1)
		previousRunner.Forwarder = ch
		previousRunner.ForwardRegister = register
		runner.Receiver = ch
		runner.ForwardRegister = register

		pushed := u.pushRunner(ctx, cycle, runner)
		if !pushed {
			return false, true
		}
		u.pushedRunnersInCurrentCycle[runner] = true
		log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "forward runner on %s (source %d)", register, previousRunner.Pc/4)
		u.forwarding++
		// TODO Return?
		return true, true
	}

	log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "data hazard: reason=%+v, types=%+v", hazards, hazardTypes)
	u.blockedDataHazard++

	return false, true
}

func (u *controlUnit) isDataHazardWithSkippedRunners(runner *risc.InstructionRunnerPc) bool {
	for _, skippedRunner := range u.skippedInCurrentCycle {
		for _, register := range runner.Runner.ReadRegisters() {
			if register == risc.Zero {
				continue
			}
			for _, skippedRegister := range skippedRunner.Runner.WriteRegisters() {
				if skippedRegister == risc.Zero {
					continue
				}
				if register == skippedRegister {
					// Read after write
					return true
				}
			}
		}

		for _, register := range runner.Runner.WriteRegisters() {
			if register == risc.Zero {
				continue
			}
			for _, skippedRegister := range skippedRunner.Runner.WriteRegisters() {
				if skippedRegister == risc.Zero {
					continue
				}
				if register == skippedRegister {
					// Write after write
					return true
				}
			}
			for _, skippedRegister := range skippedRunner.Runner.ReadRegisters() {
				if skippedRegister == risc.Zero {
					continue
				}
				if register == skippedRegister {
					// Write after read
					return true
				}
			}
		}
	}

	return false
}

func (u *controlUnit) shouldUseForwarding(runner *risc.InstructionRunnerPc, hazards []risc.Hazard, hazardTypes map[risc.HazardType]bool) (bool, *risc.InstructionRunnerPc, risc.RegisterType) {
	if len(hazardTypes) > 1 || !hazardTypes[risc.ReadAfterWrite] || len(hazards) > 1 {
		return false, nil, risc.Zero
	}

	// Can we use forwarding with an instruction pushed in the previous cycle
	for previousRunner := range u.pushedRunnersInPreviousCycle {
		for _, writeRegister := range previousRunner.Runner.WriteRegisters() {
			for _, readRegister := range runner.Runner.ReadRegisters() {
				if readRegister == risc.Zero {
					continue
				}
				if readRegister == writeRegister {
					return true, previousRunner, readRegister
				}
			}
		}
	}
	return false, nil, risc.Zero
}

func (u *controlUnit) pushRunner(ctx *risc.Context, cycle int, runner *risc.InstructionRunnerPc) bool {
	if !u.outBus.CanAdd() {
		return false
	}

	u.outBus.Add(runner, cycle)
	ctx.AddPendingRegisters(runner.Runner)
	log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "pushing runner")
	return true
}

func (u *controlUnit) flush() {
	u.pendings = comp.NewQueue[risc.InstructionRunnerPc](pendingLength)
	u.pushedRunnersInPreviousCycle = nil
}

func (u *controlUnit) isEmpty() bool {
	return u.pendings.Length() == 0
}
