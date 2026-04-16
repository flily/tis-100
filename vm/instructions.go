package vm

import (
	"strconv"
	"strings"
)

type (
	OpCode     int
	OprandType int
	Register   int
	Label      string
	Literal    int
)

type Oprand interface {
	OprandType() OprandType
}

const (
	OpEmpty   OpCode = -2 // Only used internally for empty lines, not a real instruction
	OpLabel   OpCode = -1 // Only used internally for labels, not a real instruction
	OpInvalid OpCode = iota
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

	OprandRegister OprandType = iota
	OprandLabel
	OprandLiteral

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
	opCodeNames = map[OpCode]string{
		OpNOP: "NOP",
		OpMOV: "MOV",
		OpSWP: "SWP",
		OpSAV: "SAV",
		OpADD: "ADD",
		OpSUB: "SUB",
		OpNEG: "NEG",
		OpJMP: "JMP",
		OpJEZ: "JEZ",
		OpJNZ: "JNZ",
		OpJGZ: "JGZ",
		OpJLZ: "JLZ",
		OpJRO: "JRO",
	}
	registerNames = map[Register]string{
		RegisterAcc:   "ACC",
		RegisterBak:   "BAK",
		RegisterNil:   "NIL",
		RegisterLeft:  "LEFT",
		RegisterRight: "RIGHT",
		RegisterUp:    "UP",
		RegisterDown:  "DOWN",
		RegisterAny:   "ANY",
	}

	opCodeValues   = map[string]OpCode{}
	registerValues = map[string]Register{}
)

func NewOpCode(name string) OpCode {
	if code, ok := opCodeValues[name]; ok {
		return code
	}

	return InvalidOpCode
}

func (c OpCode) String() string {
	if name, ok := opCodeNames[c]; ok {
		return name
	}

	return InvalidOpcodeName
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

func NewLabel(name string) Label {
	return Label(strings.ToUpper(name))
}

func (l Label) OprandType() OprandType {
	return OprandLabel
}

func (l Label) String() string {
	return string(l)
}

func ParseLiteral(s string) (Literal, error) {
	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}

	return Literal(value), nil
}

func (l Literal) OprandType() OprandType {
	return OprandLiteral
}

func (l Literal) String() string {
	return strconv.Itoa(int(l))
}

func (l Literal) InStandardRange() bool {
	return l >= -999 && l <= 999
}
