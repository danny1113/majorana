package mvm5

import (
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

const (
	cyclesMemoryAccess            = 50
	flushCycles                   = 1
	l1ICacheLineSizeInBytes int32 = 64
)

type CPU struct {
	ctx          *risc.Context
	fetchUnit    *fetchUnit
	decodeBus    *comp.BufferedBus[int32]
	decodeUnit   *decodeUnit
	controlBus   *comp.BufferedBus[risc.InstructionRunnerPc]
	controlUnit  *controlUnit
	executeBus   *comp.BufferedBus[risc.InstructionRunnerPc]
	executeUnits []*executeUnit
	writeBus     *comp.BufferedBus[comp.ExecutionContext]
	writeUnit    *writeUnit
	branchUnit   *btbBranchUnit
}

func NewCPU(debug bool, memoryBytes int) *CPU {
	decodeBus := comp.NewBufferedBus[int32](2, 2)
	controlBus := comp.NewBufferedBus[risc.InstructionRunnerPc](2, 2)
	executeBus := comp.NewBufferedBus[risc.InstructionRunnerPc](2, 2)
	writeBus := comp.NewBufferedBus[comp.ExecutionContext](2, 2)

	fu := newFetchUnit(l1ICacheLineSizeInBytes, decodeBus)
	du := newDecodeUnit(decodeBus, controlBus)
	bu := newBTBBranchUnit(4, fu, du)
	return &CPU{
		ctx:         risc.NewContext(debug, memoryBytes),
		fetchUnit:   fu,
		decodeBus:   decodeBus,
		decodeUnit:  du,
		controlBus:  controlBus,
		controlUnit: newControlUnit(controlBus, executeBus),
		executeBus:  executeBus,
		executeUnits: []*executeUnit{
			newExecuteUnit(bu, executeBus, writeBus),
			newExecuteUnit(bu, executeBus, writeBus),
		},
		writeBus:   writeBus,
		writeUnit:  newWriteUnit(writeBus),
		branchUnit: bu,
	}
}

func (m *CPU) Context() *risc.Context {
	return m.ctx
}

func (m *CPU) Run(app risc.Application) (int, error) {
	cycle := 0
	for {
		cycle += 1
		log(m.ctx, "Cycle %d", cycle)
		if cycle > 1000 {
			return 0, nil
		}
		m.decodeBus.Connect(cycle)
		m.controlBus.Connect(cycle)
		m.executeBus.Connect(cycle)
		m.writeBus.Connect(cycle)

		// Fetch
		m.fetchUnit.cycle(cycle, app, m.ctx)

		// Decode
		m.decodeUnit.cycle(cycle, app, m.ctx)

		// Control
		m.controlUnit.cycle(cycle, m.ctx)

		// Execute
		var (
			flush bool
			pc    int32
			ret   bool
		)
		for _, executeUnit := range m.executeUnits {
			f, p, r, err := executeUnit.cycle(cycle, m.ctx, app)
			if err != nil {
				return 0, err
			}
			flush = flush || f
			pc = max(pc, p)
			ret = ret || r
		}

		// Write back
		m.writeUnit.cycle(m.ctx)
		log(m.ctx, "\tRegisters: %v", m.ctx.Registers)

		if ret {
			return cycle, nil
		}
		if flush {
			log(m.ctx, "\t️⚠️ Flush to %d", pc/4)
			m.flush(pc)
			cycle += flushCycles
			continue
		}

		if m.isEmpty() {
			if m.ctx.Registers[risc.Ra] != 0 {
				m.ctx.Registers[risc.Ra] = 0
				m.fetchUnit.reset(m.ctx.Registers[risc.Ra], false)
				continue
			}
			break
		}
	}
	return cycle, nil
}

func (m *CPU) flush(pc int32) {
	m.fetchUnit.flush(pc)
	m.decodeUnit.flush()
	m.controlUnit.flush()
	for _, executeUnit := range m.executeUnits {
		executeUnit.flush()
	}
	m.decodeBus.Clean()
	m.controlBus.Clean()
	m.executeBus.Clean()
	m.writeBus.Clean()
	m.ctx.Flush()
}

func (m *CPU) isEmpty() bool {
	empty := m.fetchUnit.isEmpty() &&
		m.decodeUnit.isEmpty() &&
		m.controlUnit.isEmpty() &&
		m.writeUnit.isEmpty() &&
		m.decodeBus.IsEmpty() &&
		m.controlBus.IsEmpty() &&
		m.executeBus.IsEmpty() &&
		m.writeBus.IsEmpty()
	if !empty {
		//if m.ctx.Debug {
		//	fmt.Println("fu:", m.fetchUnit.isEmpty(), "du:", m.decodeUnit.isEmpty(), "cu:", m.controlUnit.isEmpty(), "wu:", m.writeUnit.isEmpty(),
		//		"db:", m.decodeBus.IsEmpty(), "cb:", m.controlBus.IsEmpty(), "eb:", m.executeBus.IsEmpty(), "wb:", m.writeBus.IsEmpty())
		//}
		return false
	}
	for _, executeUnit := range m.executeUnits {
		if !executeUnit.isEmpty() {
			return false
		}
	}
	return true
}
