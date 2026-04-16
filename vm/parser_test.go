package vm

import (
	"testing"
)

func checkParseInstruction(t *testing.T, code string, expected Instruction) {
	t.Helper()

	got, err := ParseInstruction([]rune(code))
	if err != nil {
		t.Fatalf("ParseInstruction() error:\n%s", err)
	}

	if !got.Equals(expected) {
		t.Errorf("ParseInstruction() = %v, want %v", got, expected)
	}
}

func TestParseInstructWithNoOprands(t *testing.T) {
	code := "NOP"
	exp := Instruction{
		Opcode: OpNOP,
	}

	checkParseInstruction(t, code, exp)
}

func TestParseInstructionWithOneRegisterOperand(t *testing.T) {
	code := "ADD ACC"
	exp := Instruction{
		Opcode:   OpADD,
		Oprands1: RegisterAcc,
	}

	checkParseInstruction(t, code, exp)
}

func TestParseInstructionWithOneLiteralOprand(t *testing.T) {
	code := "SUB 42"
	exp := Instruction{
		Opcode:   OpSUB,
		Oprands1: Literal(42),
	}

	checkParseInstruction(t, code, exp)
}

func TestParseInstructionWithOneLabelOprand(t *testing.T) {
	code := "JMP LOOP"
	exp := Instruction{
		Opcode:   OpJMP,
		Oprands1: Label("LOOP"),
	}

	checkParseInstruction(t, code, exp)
}

func TestParseInstructionWithTwoOprands(t *testing.T) {
	code := "MOV ACC, LEFT"
	exp := Instruction{
		Opcode:   OpMOV,
		Oprands1: RegisterAcc,
		Oprands2: RegisterLeft,
	}

	checkParseInstruction(t, code, exp)
}
