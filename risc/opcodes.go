package risc

import "fmt"

type InstructionRunnerPc struct {
	Runner InstructionRunner
	Pc     int32
}

type InstructionRunner interface {
	Run(ctx *Context, labels map[string]int32, pc int32) (Execution, error)
	InstructionType() InstructionType
	ReadRegisters() []RegisterType
	WriteRegisters() []RegisterType
}

type add struct {
	rd  RegisterType
	rs1 RegisterType
	rs2 RegisterType
}

func (a add) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	register, value := IsRegisterChange(a.rd, ctx.Registers[a.rs1]+ctx.Registers[a.rs2])
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (a add) InstructionType() InstructionType {
	return Add
}

func (a add) ReadRegisters() []RegisterType {
	return []RegisterType{a.rs1, a.rs2}
}

func (a add) WriteRegisters() []RegisterType {
	return []RegisterType{a.rd}
}

type addi struct {
	imm int32
	rd  RegisterType
	rs  RegisterType
}

func (a addi) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	register, value := IsRegisterChange(a.rd, ctx.Registers[a.rs]+a.imm)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (a addi) InstructionType() InstructionType {
	return Addi
}

func (a addi) ReadRegisters() []RegisterType {
	return []RegisterType{a.rs}
}

func (a addi) WriteRegisters() []RegisterType {
	return []RegisterType{a.rd}
}

type and struct {
	rd  RegisterType
	rs1 RegisterType
	rs2 RegisterType
}

func (a and) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	register, value := IsRegisterChange(a.rd, ctx.Registers[a.rs1]&ctx.Registers[a.rs2])
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (a and) InstructionType() InstructionType {
	return And
}

func (a and) ReadRegisters() []RegisterType {
	return []RegisterType{a.rs1, a.rs2}
}

func (a and) WriteRegisters() []RegisterType {
	return []RegisterType{a.rd}
}

type andi struct {
	imm int32
	rd  RegisterType
	rs  RegisterType
}

func (a andi) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	register, value := IsRegisterChange(a.rd, ctx.Registers[a.rs]&a.imm)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (a andi) InstructionType() InstructionType {
	return Andi
}

func (a andi) ReadRegisters() []RegisterType {
	return []RegisterType{a.rs}
}

func (a andi) WriteRegisters() []RegisterType {
	return []RegisterType{a.rd}
}

type auipc struct {
	rd  RegisterType
	imm int32
}

func (a auipc) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	register, value := IsRegisterChange(a.rd, pc+(a.imm<<12))
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (a auipc) InstructionType() InstructionType {
	return Auipc
}

func (a auipc) ReadRegisters() []RegisterType {
	return nil
}

func (a auipc) WriteRegisters() []RegisterType {
	return []RegisterType{a.rd}
}

type beq struct {
	rs1   RegisterType
	rs2   RegisterType
	label string
}

func (b beq) Run(ctx *Context, labels map[string]int32, pc int32) (Execution, error) {
	if ctx.Registers[b.rs1] == ctx.Registers[b.rs2] {
		addr, ok := labels[b.label]
		if !ok {
			return Execution{}, fmt.Errorf("label %s does not exist", b.label)
		}
		return Execution{
			NextPc:   addr,
			PcChange: true,
		}, nil
	}
	return Execution{}, nil
}

func (b beq) InstructionType() InstructionType {
	return Beq
}

func (b beq) ReadRegisters() []RegisterType {
	return []RegisterType{b.rs1, b.rs2}
}

func (b beq) WriteRegisters() []RegisterType {
	return nil
}

type bge struct {
	rs1   RegisterType
	rs2   RegisterType
	label string
}

func (b bge) Run(ctx *Context, labels map[string]int32, pc int32) (Execution, error) {
	if ctx.Registers[b.rs1] >= ctx.Registers[b.rs2] {
		addr, ok := labels[b.label]
		if !ok {
			return Execution{}, fmt.Errorf("label %s does not exist", b.label)
		}
		return Execution{
			NextPc:   addr,
			PcChange: true,
		}, nil
	}
	return Execution{}, nil
}

func (b bge) InstructionType() InstructionType {
	return Bge
}

func (b bge) ReadRegisters() []RegisterType {
	return []RegisterType{b.rs1, b.rs2}
}

func (b bge) WriteRegisters() []RegisterType {
	return nil
}

type bgeu struct {
	rs1   RegisterType
	rs2   RegisterType
	label string
}

func (b bgeu) Run(ctx *Context, labels map[string]int32, pc int32) (Execution, error) {
	if ctx.Registers[b.rs1] >= ctx.Registers[b.rs2] {
		addr, ok := labels[b.label]
		if !ok {
			return Execution{}, fmt.Errorf("label %s does not exist", b.label)
		}
		return Execution{
			NextPc:   addr,
			PcChange: true,
		}, nil
	}
	return Execution{}, nil
}

func (b bgeu) InstructionType() InstructionType {
	return Bgeu
}

func (b bgeu) ReadRegisters() []RegisterType {
	return []RegisterType{b.rs1, b.rs2}
}

func (b bgeu) WriteRegisters() []RegisterType {
	return nil
}

type blt struct {
	rs1   RegisterType
	rs2   RegisterType
	label string
}

func (b blt) Run(ctx *Context, labels map[string]int32, pc int32) (Execution, error) {
	if ctx.Registers[b.rs1] < ctx.Registers[b.rs2] {
		addr, ok := labels[b.label]
		if !ok {
			return Execution{}, fmt.Errorf("label %s does not exist", b.label)
		}
		return Execution{
			NextPc:   addr,
			PcChange: true,
		}, nil
	}
	return Execution{}, nil
}

func (b blt) InstructionType() InstructionType {
	return Blt
}

func (b blt) ReadRegisters() []RegisterType {
	return []RegisterType{b.rs1, b.rs2}
}

func (b blt) WriteRegisters() []RegisterType {
	return nil
}

type bltu struct {
	rs1   RegisterType
	rs2   RegisterType
	label string
}

func (b bltu) Run(ctx *Context, labels map[string]int32, pc int32) (Execution, error) {
	if ctx.Registers[b.rs1] < ctx.Registers[b.rs2] {
		addr, ok := labels[b.label]
		if !ok {
			return Execution{}, fmt.Errorf("label %s does not exist", b.label)
		}
		return Execution{
			NextPc:   addr,
			PcChange: true,
		}, nil
	}
	return Execution{}, nil
}

func (b bltu) InstructionType() InstructionType {
	return Bltu
}

func (b bltu) ReadRegisters() []RegisterType {
	return []RegisterType{b.rs1, b.rs2}
}

func (b bltu) WriteRegisters() []RegisterType {
	return nil
}

type bne struct {
	rs1   RegisterType
	rs2   RegisterType
	label string
}

func (b bne) Run(ctx *Context, labels map[string]int32, pc int32) (Execution, error) {
	if ctx.Registers[b.rs1] != ctx.Registers[b.rs2] {
		addr, ok := labels[b.label]
		if !ok {
			return Execution{}, fmt.Errorf("label %s does not exist", b.label)
		}
		return Execution{
			NextPc:   addr,
			PcChange: true,
		}, nil
	}
	return Execution{}, nil
}

func (b bne) InstructionType() InstructionType {
	return Bne
}

func (b bne) ReadRegisters() []RegisterType {
	return []RegisterType{b.rs1, b.rs2}
}

func (b bne) WriteRegisters() []RegisterType {
	return nil
}

type div struct {
	rd  RegisterType
	rs1 RegisterType
	rs2 RegisterType
}

func (d div) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	if ctx.Registers[d.rs2] == 0 {
		return Execution{}, fmt.Errorf("division by zero")
	}
	register, value := IsRegisterChange(d.rd, ctx.Registers[d.rs1]/ctx.Registers[d.rs2])
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (d div) InstructionType() InstructionType {
	return Div
}

func (d div) ReadRegisters() []RegisterType {
	return []RegisterType{d.rs1, d.rs2}
}

func (d div) WriteRegisters() []RegisterType {
	return []RegisterType{d.rd}
}

type jal struct {
	label string
	rd    RegisterType
}

func (j jal) Run(ctx *Context, labels map[string]int32, pc int32) (Execution, error) {
	addr, ok := labels[j.label]
	if !ok {
		return Execution{}, fmt.Errorf("label %s does not exist", j.label)
	}
	ctx.Registers[Ra] = pc
	register, value := IsRegisterChange(j.rd, pc+4)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
		NextPc:         addr,
		PcChange:       true,
	}, nil
}

func (j jal) InstructionType() InstructionType {
	return Jal
}

func (j jal) ReadRegisters() []RegisterType {
	return nil
}

func (j jal) WriteRegisters() []RegisterType {
	return []RegisterType{j.rd}
}

type jalr struct {
	rd  RegisterType
	rs  RegisterType
	imm int32
}

func (j jalr) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	register, value := IsRegisterChange(j.rd, pc+4)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
		NextPc:         ctx.Registers[j.rs] + j.imm,
		PcChange:       true,
	}, nil
}

func (j jalr) InstructionType() InstructionType {
	return Jalr
}

func (j jalr) ReadRegisters() []RegisterType {
	return []RegisterType{j.rs}
}

func (j jalr) WriteRegisters() []RegisterType {
	return []RegisterType{j.rd}
}

type lui struct {
	rd  RegisterType
	imm int32
}

func (l lui) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	register, value := IsRegisterChange(l.rd, l.imm<<12)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (l lui) InstructionType() InstructionType {
	return Lui
}

func (l lui) ReadRegisters() []RegisterType {
	return nil
}

func (l lui) WriteRegisters() []RegisterType {
	return []RegisterType{l.rd}
}

type lb struct {
	rs2    RegisterType
	offset int32
	rs1    RegisterType
}

func (l lb) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	idx := ctx.Registers[l.rs1] + l.offset
	n := ctx.Memory[idx]

	register, value := IsRegisterChange(l.rs2, int32(n))
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (l lb) InstructionType() InstructionType {
	return Lb
}

func (l lb) ReadRegisters() []RegisterType {
	return []RegisterType{l.rs1, l.rs2}
}

func (l lb) WriteRegisters() []RegisterType {
	return nil
}

type lh struct {
	rs2    RegisterType
	offset int32
	rs1    RegisterType
}

func (l lh) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	idx := ctx.Registers[l.rs1] + l.offset
	i1 := ctx.Memory[idx]
	idx++
	i2 := ctx.Memory[idx]

	n := i32FromBytes(i1, i2, 0, 0)
	register, value := IsRegisterChange(l.rs2, n)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (l lh) InstructionType() InstructionType {
	return Lh
}

func (l lh) ReadRegisters() []RegisterType {
	return []RegisterType{l.rs1, l.rs2}
}

func (l lh) WriteRegisters() []RegisterType {
	return nil
}

type li struct {
	rd  RegisterType
	imm int32
}

func (l li) Run(ctx *Context, _ map[string]int32, _ int32) (Execution, error) {
	register, value := IsRegisterChange(l.rd, l.imm)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (l li) InstructionType() InstructionType {
	return Li
}

func (l li) ReadRegisters() []RegisterType {
	return nil
}

func (l li) WriteRegisters() []RegisterType {
	return []RegisterType{l.rd}
}

type lw struct {
	rd     RegisterType
	offset int32
	rs     RegisterType
}

func (l lw) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	idx := ctx.Registers[l.rs] + l.offset
	i1 := ctx.Memory[idx]
	idx++
	i2 := ctx.Memory[idx]
	idx++
	i3 := ctx.Memory[idx]
	idx++
	i4 := ctx.Memory[idx]

	n := i32FromBytes(i1, i2, i3, i4)
	register, value := IsRegisterChange(l.rd, n)
	if ctx.Debug {
		fmt.Printf("\t\tRun: Lw %s %d\n", register, value)
	}
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (l lw) InstructionType() InstructionType {
	return Lw
}

func (l lw) ReadRegisters() []RegisterType {
	return []RegisterType{l.rs}
}

func (l lw) WriteRegisters() []RegisterType {
	return []RegisterType{l.rd}
}

type nop struct{}

func (n nop) Run(_ *Context, _ map[string]int32, pc int32) (Execution, error) {
	return Execution{}, nil
}

func (n nop) InstructionType() InstructionType {
	return Nop
}

func (n nop) ReadRegisters() []RegisterType {
	return nil
}

func (n nop) WriteRegisters() []RegisterType {
	return nil
}

type mul struct {
	rd  RegisterType
	rs1 RegisterType
	rs2 RegisterType
}

func (m mul) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	register, value := IsRegisterChange(m.rd, ctx.Registers[m.rs1]*ctx.Registers[m.rs2])
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (m mul) InstructionType() InstructionType {
	return Mul
}

func (m mul) ReadRegisters() []RegisterType {
	return []RegisterType{m.rs1, m.rs2}
}

func (m mul) WriteRegisters() []RegisterType {
	return []RegisterType{m.rd}
}

type mv struct {
	rd RegisterType
	rs RegisterType
}

func (m mv) Run(ctx *Context, _ map[string]int32, _ int32) (Execution, error) {
	register, value := IsRegisterChange(m.rd, ctx.Registers[m.rs])
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (m mv) InstructionType() InstructionType {
	return Mv
}

func (m mv) ReadRegisters() []RegisterType {
	return []RegisterType{m.rs}
}

func (m mv) WriteRegisters() []RegisterType {
	return []RegisterType{m.rd}
}

type or struct {
	rd  RegisterType
	rs1 RegisterType
	rs2 RegisterType
}

func (o or) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	register, value := IsRegisterChange(o.rd, ctx.Registers[o.rs1]|ctx.Registers[o.rs2])
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (o or) InstructionType() InstructionType {
	return Or
}

func (o or) ReadRegisters() []RegisterType {
	return []RegisterType{o.rs1, o.rs2}
}

func (o or) WriteRegisters() []RegisterType {
	return []RegisterType{o.rd}
}

type ori struct {
	imm int32
	rd  RegisterType
	rs  RegisterType
}

func (o ori) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	register, value := IsRegisterChange(o.rd, ctx.Registers[o.rs]|o.imm)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (o ori) InstructionType() InstructionType {
	return Ori
}

func (o ori) ReadRegisters() []RegisterType {
	return []RegisterType{o.rs}
}

func (o ori) WriteRegisters() []RegisterType {
	return []RegisterType{o.rd}
}

type rem struct {
	rd  RegisterType
	rs1 RegisterType
	rs2 RegisterType
}

func (r rem) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	if ctx.Debug {
		fmt.Printf("\t\tRun: Rem %d %d\n", ctx.Registers[r.rs1], ctx.Registers[r.rs2])
	}
	register, value := IsRegisterChange(r.rd, ctx.Registers[r.rs1]%ctx.Registers[r.rs2])
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (r rem) InstructionType() InstructionType {
	return Rem
}

func (r rem) ReadRegisters() []RegisterType {
	return []RegisterType{r.rs1, r.rs2}
}

func (r rem) WriteRegisters() []RegisterType {
	return []RegisterType{r.rd}
}

type ret struct{}

func (r ret) Run(_ *Context, _ map[string]int32, _ int32) (Execution, error) {
	return Execution{Return: true}, nil
}

func (r ret) InstructionType() InstructionType {
	return Ret
}

func (r ret) ReadRegisters() []RegisterType {
	return nil
}

func (r ret) WriteRegisters() []RegisterType {
	return nil
}

type sb struct {
	rs2    RegisterType
	offset int32
	rs1    RegisterType
}

func (s sb) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	idx := ctx.Registers[s.rs1] + s.offset
	n := ctx.Registers[s.rs2]
	return Execution{
		MemoryChange:  true,
		MemoryChanges: map[int32]int8{idx: int8(n)},
	}, nil
}

func (s sb) InstructionType() InstructionType {
	return Sb
}

func (s sb) ReadRegisters() []RegisterType {
	return []RegisterType{s.rs1, s.rs2}
}

func (s sb) WriteRegisters() []RegisterType {
	return nil
}

type sh struct {
	rs2    RegisterType
	offset int32
	rs1    RegisterType
}

func (s sh) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	idx := ctx.Registers[s.rs1] + s.offset
	n := ctx.Registers[s.rs2]
	bytes := BytesFromLowBits(n)
	return Execution{
		MemoryChange: true,
		MemoryChanges: map[int32]int8{
			idx:     bytes[0],
			idx + 1: bytes[1],
		},
	}, nil
}

func (s sh) InstructionType() InstructionType {
	return Sh
}

func (s sh) ReadRegisters() []RegisterType {
	return []RegisterType{s.rs1, s.rs2}
}

func (s sh) WriteRegisters() []RegisterType {
	return nil
}

type sll struct {
	rd  RegisterType
	rs1 RegisterType
	rs2 RegisterType
}

func (s sll) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	register, value := IsRegisterChange(s.rd, ctx.Registers[s.rs1]<<uint(ctx.Registers[s.rs2]))
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (s sll) InstructionType() InstructionType {
	return Sll
}

func (s sll) ReadRegisters() []RegisterType {
	return []RegisterType{s.rs1, s.rs2}
}

func (s sll) WriteRegisters() []RegisterType {
	return []RegisterType{s.rd}
}

type slli struct {
	rd  RegisterType
	rs  RegisterType
	imm int32
}

func (s slli) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	register, value := IsRegisterChange(s.rd, ctx.Registers[s.rs]<<uint(s.imm))
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (s slli) InstructionType() InstructionType {
	return Slli
}

func (s slli) ReadRegisters() []RegisterType {
	return []RegisterType{s.rs}
}

func (s slli) WriteRegisters() []RegisterType {
	return []RegisterType{s.rd}
}

type slt struct {
	rd  RegisterType
	rs1 RegisterType
	rs2 RegisterType
}

func (s slt) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	var register RegisterType
	var value int32
	if ctx.Registers[s.rs1] < ctx.Registers[s.rs2] {
		register, value = IsRegisterChange(s.rd, 1)
	} else {
		register, value = IsRegisterChange(s.rd, 0)
	}
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (s slt) InstructionType() InstructionType {
	return Slt
}

func (s slt) ReadRegisters() []RegisterType {
	return []RegisterType{s.rs1, s.rs2}
}

func (s slt) WriteRegisters() []RegisterType {
	return []RegisterType{s.rd}
}

type sltu struct {
	rd  RegisterType
	rs1 RegisterType
	rs2 RegisterType
}

func (s sltu) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	var register RegisterType
	var value int32
	if ctx.Registers[s.rs1] < ctx.Registers[s.rs2] {
		register, value = IsRegisterChange(s.rd, 1)
	} else {
		register, value = IsRegisterChange(s.rd, 0)
	}
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (s sltu) InstructionType() InstructionType {
	return Sltu
}

func (s sltu) ReadRegisters() []RegisterType {
	return []RegisterType{s.rs1, s.rs2}
}

func (s sltu) WriteRegisters() []RegisterType {
	return []RegisterType{s.rd}
}

type slti struct {
	rd  RegisterType
	rs  RegisterType
	imm int32
}

func (s slti) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	var register RegisterType
	var value int32
	if ctx.Registers[s.rs] < s.imm {
		register, value = IsRegisterChange(s.rd, 1)
	} else {
		register, value = IsRegisterChange(s.rd, 0)
	}
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (s slti) InstructionType() InstructionType {
	return Slti
}

func (s slti) ReadRegisters() []RegisterType {
	return []RegisterType{s.rs}
}

func (s slti) WriteRegisters() []RegisterType {
	return []RegisterType{s.rd}
}

type sra struct {
	rd  RegisterType
	rs1 RegisterType
	rs2 RegisterType
}

func (s sra) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	register, value := IsRegisterChange(s.rd, ctx.Registers[s.rs1]>>ctx.Registers[s.rs2])
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (s sra) InstructionType() InstructionType {
	return Sra
}

func (s sra) ReadRegisters() []RegisterType {
	return []RegisterType{s.rs1, s.rs2}
}

func (s sra) WriteRegisters() []RegisterType {
	return []RegisterType{s.rd}
}

type srai struct {
	rd  RegisterType
	rs  RegisterType
	imm int32
}

func (s srai) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	register, value := IsRegisterChange(s.rd, ctx.Registers[s.rs]>>s.imm)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (s srai) InstructionType() InstructionType {
	return Srai
}

func (s srai) ReadRegisters() []RegisterType {
	return []RegisterType{s.rs}
}

func (s srai) WriteRegisters() []RegisterType {
	return []RegisterType{s.rd}
}

type srl struct {
	rd  RegisterType
	rs1 RegisterType
	rs2 RegisterType
}

func (s srl) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	register, value := IsRegisterChange(s.rd, ctx.Registers[s.rs1]>>ctx.Registers[s.rs2])
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (s srl) InstructionType() InstructionType {
	return Srl
}

func (s srl) ReadRegisters() []RegisterType {
	return []RegisterType{s.rs1, s.rs2}
}

func (s srl) WriteRegisters() []RegisterType {
	return []RegisterType{s.rd}
}

type srli struct {
	rd  RegisterType
	rs  RegisterType
	imm int32
}

func (s srli) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	register, value := IsRegisterChange(s.rd, ctx.Registers[s.rs]>>s.imm)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (s srli) InstructionType() InstructionType {
	return Srli
}

func (s srli) ReadRegisters() []RegisterType {
	return []RegisterType{s.rs}
}

func (s srli) WriteRegisters() []RegisterType {
	return []RegisterType{s.rd}
}

type sub struct {
	rd  RegisterType
	rs1 RegisterType
	rs2 RegisterType
}

func (s sub) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	register, value := IsRegisterChange(s.rd, ctx.Registers[s.rs1]-ctx.Registers[s.rs2])
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (s sub) InstructionType() InstructionType {
	return Sub
}

func (s sub) ReadRegisters() []RegisterType {
	return []RegisterType{s.rs1, s.rs2}
}

func (s sub) WriteRegisters() []RegisterType {
	return []RegisterType{s.rd}
}

type sw struct {
	rs2    RegisterType
	offset int32
	rs1    RegisterType
}

func (s sw) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	idx := ctx.Registers[s.rs1] + s.offset
	n := ctx.Registers[s.rs2]
	bytes := BytesFromLowBits(n)
	if ctx.Debug {
		fmt.Printf("\t\tRun: Sw %d to %d\n", idx, n)
	}
	return Execution{
		MemoryChange: true,
		MemoryChanges: map[int32]int8{
			idx:     bytes[0],
			idx + 1: bytes[1],
			idx + 2: bytes[2],
			idx + 3: bytes[3],
		},
	}, nil
}

func (s sw) InstructionType() InstructionType {
	return Sw
}

func (s sw) ReadRegisters() []RegisterType {
	return []RegisterType{s.rs1, s.rs2}
}

func (s sw) WriteRegisters() []RegisterType {
	return nil
}

type xor struct {
	rd  RegisterType
	rs1 RegisterType
	rs2 RegisterType
}

func (x xor) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	register, value := IsRegisterChange(x.rd, ctx.Registers[x.rs1]^ctx.Registers[x.rs2])
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (x xor) InstructionType() InstructionType {
	return Xor
}

func (x xor) ReadRegisters() []RegisterType {
	return []RegisterType{x.rs1, x.rs2}
}

func (x xor) WriteRegisters() []RegisterType {
	return []RegisterType{x.rd}
}

type xori struct {
	imm int32
	rd  RegisterType
	rs  RegisterType
}

func (x xori) Run(ctx *Context, _ map[string]int32, pc int32) (Execution, error) {
	register, value := IsRegisterChange(x.rd, ctx.Registers[x.rs]^x.imm)
	return Execution{
		RegisterChange: true,
		Register:       register,
		RegisterValue:  value,
	}, nil
}

func (x xori) InstructionType() InstructionType {
	return Xori
}

func (x xori) ReadRegisters() []RegisterType {
	return []RegisterType{x.rs}
}

func (x xori) WriteRegisters() []RegisterType {
	return []RegisterType{x.rd}
}
