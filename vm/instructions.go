package vm

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type (
	Opcode     int
	OprandType uint
	Register   int
	Label      string
	Value      int
)

type Oprand interface {
	OprandType() OprandType
	Equal(o Oprand) bool
}

const (
	OpEmpty   Opcode = -1 // Only used internally for empty lines, not a real instruction
	OpInvalid Opcode = iota
	OpNOP
	OpMOV
	OpSWP
	OpSAV
	OpADD
	OpSUB
	OpNEG
	OpJMP
	OpJEZ
	OpJNZ
	OpJGZ
	OpJLZ
	OpJRO
	OpHCF // HCF is hidden instruction, not documented in the manual, use to get archivement.

	OprandRegister OprandType = 1
	OprandLabel    OprandType = 2
	OprandValue    OprandType = 4

	RegisterNA      Register = -1
	RegisterInvalid Register = iota
	RegisterAcc
	RegisterBak
	RegisterNil
	RegisterLeft
	RegisterRight
	RegisterUp
	RegisterDown
	RegisterAny
	RegisterLast

	InvalidOpCode       = OpInvalid
	InvalidOpcodeName   = "INVALID"
	InvalidRegister     = RegisterInvalid
	InvalidRegisterName = "INVALID"
)

var (
	opCodeNames = map[Opcode]string{
		OpEmpty: "#EMPTY",
		OpNOP:   "NOP",
		OpMOV:   "MOV",
		OpSWP:   "SWP",
		OpSAV:   "SAV",
		OpADD:   "ADD",
		OpSUB:   "SUB",
		OpNEG:   "NEG",
		OpJMP:   "JMP",
		OpJEZ:   "JEZ",
		OpJNZ:   "JNZ",
		OpJGZ:   "JGZ",
		OpJLZ:   "JLZ",
		OpJRO:   "JRO",
		OpHCF:   "HCF",
	}
	opcodeAcceptOprands = map[Opcode][]OprandType{
		OpMOV: {OprandRegister | OprandValue, OprandRegister | OprandValue},
		OpADD: {OprandRegister | OprandValue},
		OpSUB: {OprandRegister | OprandValue},
		OpJMP: {OprandLabel},
		OpJEZ: {OprandLabel},
		OpJNZ: {OprandLabel},
		OpJGZ: {OprandLabel},
		OpJLZ: {OprandLabel},
		OpJRO: {OprandRegister | OprandValue},
	}

	registerNames = map[Register]string{
		RegisterNA:    "N/A",
		RegisterAcc:   "ACC",
		RegisterBak:   "BAK",
		RegisterNil:   "NIL",
		RegisterLeft:  "LEFT",
		RegisterRight: "RIGHT",
		RegisterUp:    "UP",
		RegisterDown:  "DOWN",
		RegisterAny:   "ANY",
	}

	opCodeValues   = map[string]Opcode{}
	registerValues = map[string]Register{}
)

func (t OprandType) Include(other OprandType) bool {
	return (t & other) == other
}

func NewOpcode(name string) Opcode {
	if code, ok := opCodeValues[name]; ok {
		return code
	}

	return InvalidOpCode
}

func (c Opcode) String() string {
	if name, ok := opCodeNames[c]; ok {
		return name
	}

	return InvalidOpcodeName
}

func (c Opcode) AcceptOprands() []OprandType {
	if oprands, ok := opcodeAcceptOprands[c]; ok {
		return slices.Clone(oprands)
	}

	return nil
}

func NewRegister(name string) Register {
	if reg, ok := registerValues[name]; ok {
		return reg
	}

	return InvalidRegister
}

func (r Register) OprandType() OprandType {
	return OprandRegister
}

func (r Register) String() string {
	if name, ok := registerNames[r]; ok {
		return name
	}

	return InvalidRegisterName
}

func (r Register) Equal(o Oprand) bool {
	if oReg, ok := o.(Register); ok {
		return r == oReg
	}

	return false
}

func NewLabel(name string) Label {
	return Label(strings.ToUpper(name))
}

func (l Label) OprandType() OprandType {
	return OprandLabel
}

func (l Label) String() string {
	return string(l)
}

func (l Label) Equal(o Oprand) bool {
	if oLabel, ok := o.(Label); ok {
		return l == oLabel
	}

	return false
}

func ParseValue(s string) (Value, error) {
	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}

	return Value(value), nil
}

func (l Value) OprandType() OprandType {
	return OprandValue
}

func (l Value) String() string {
	return strconv.Itoa(int(l))
}

func (l Value) Equal(o Oprand) bool {
	if oValue, ok := o.(Value); ok {
		return l == oValue
	}

	return false
}

func (l Value) InStandardRange() bool {
	return l >= -999 && l <= 999
}

type Context struct {
	Raw   []rune
	Start int
	End   int
	Line  int
}

func newContext(content []rune, start int, end int, line int) *Context {
	c := &Context{
		Raw:   content,
		Start: start,
		End:   end,
		Line:  line,
	}

	return c
}

func NewContext(content []rune) *Context {
	return newContext(content, 0, len(content), -1)
}

func (c *Context) Mark(start int, end int) *Context {
	return newContext(c.Raw, start, end, c.Line)
}

func (c *Context) MarkAll() *Context {
	return newContext(c.Raw, 0, len(c.Raw), c.Line)
}

func (c *Context) Message(message string) string {
	leadSpace := strings.Repeat(" ", c.Start)
	lines := []string{
		string(c.Raw),
		leadSpace + strings.Repeat("^", c.End-c.Start),
		leadSpace + message,
	}
	return strings.Join(lines, "\n")
}

func (c *Context) Error(message string, args ...any) *SyntaxError {
	err := &SyntaxError{
		ctx:     *c,
		message: fmt.Sprintf(message, args...),
	}

	return err
}

type Instruction struct {
	Label      Label
	LabelCtx   *Context
	Opcode     Opcode
	OpCodeCtx  *Context
	Oprand1    Oprand
	Oprand1Ctx *Context
	Oprand2    Oprand
	Oprand2Ctx *Context
	Breakpoint bool
	Comment    string
}

var InvalidInstruction = Instruction{
	Opcode: InvalidOpCode,
}

func (i *Instruction) Equals(o Instruction) bool {
	if i.Breakpoint != o.Breakpoint {
		return false
	}

	if i.Opcode != o.Opcode || i.Breakpoint != o.Breakpoint || i.Comment != o.Comment {
		return false
	}

	if (i.Oprand1 == nil) != (o.Oprand1 == nil) || (i.Oprand2 == nil) != (o.Oprand2 == nil) {
		return false
	}

	if i.Oprand1 != nil && o.Oprand1 != nil {
		if !i.Oprand1.Equal(o.Oprand1) {
			return false
		}
	}

	if i.Oprand2 != nil && o.Oprand2 != nil {
		if !i.Oprand2.Equal(o.Oprand2) {
			return false
		}
	}

	return true
}

func (i *Instruction) ExpectBreakpoint(r rune) bool {
	if i.Breakpoint {
		return false
	}

	return r == CharBreakpoint
}

func (i *Instruction) Empty() bool {
	return i.Opcode <= OpInvalid
}

func (i *Instruction) SetLabel(label string, ctx *Context) {
	i.Label = NewLabel(label)
	i.LabelCtx = ctx
}

func (i *Instruction) SetOpcode(opcode Opcode, ctx *Context) {
	i.Opcode = opcode
	i.OpCodeCtx = ctx
}

func (i *Instruction) AddOprand(oprand Oprand, ctx *Context) {
	if i.Oprand1 == nil {
		i.Oprand1 = oprand
		i.Oprand1Ctx = ctx

	} else if i.Oprand2 == nil {
		i.Oprand2 = oprand
		i.Oprand2Ctx = ctx
	}
}

type Code []Instruction

func (c Code) Equals(o Code) bool {
	if len(c) != len(o) {
		return false
	}

	for i := range c {
		if !c[i].Equals(o[i]) {
			return false
		}
	}

	return true
}
