package vm

import (
	"testing"

	"strings"
)

func checkParseInstructionSuccess(t *testing.T, code string, expected Instruction) {
	t.Helper()

	got, err := ParseInstruction([]rune(code))
	if err != nil {
		t.Fatalf("ParseInstruction() error:\n%s", err)
	}

	if !got.Equals(expected) {
		t.Errorf("ParseInstruction() = %v, expect %v", got, expected)
	}
}

func checkParseInstructionError(t *testing.T, code string, errMessage []string) {
	t.Helper()

	got, err := ParseInstruction([]rune(code))
	if got.Opcode != OpInvalid || err == nil {
		t.Errorf("ParseInstruction() expected nil and error, got %v", got)
		t.Errorf("%s", err)
	}

	gotMessage := err.Error()
	expected := strings.Join(errMessage, "\n")
	if gotMessage != expected {
		t.Errorf("wrong error message, got:\n%s\nexpect:\n%s", gotMessage, expected)

	}
}

func TestParseInstructionWithEmptyContent(t *testing.T) {
	code := ""
	exp := Instruction{
		Opcode: OpEmpty,
	}

	checkParseInstructionSuccess(t, code, exp)
}

func TestParseInstructionWithOnlySpaces(t *testing.T) {
	code := "    "
	exp := Instruction{
		Opcode: OpEmpty,
	}

	checkParseInstructionSuccess(t, code, exp)
}

func TestParseInstructWithNoOprands(t *testing.T) {
	code := "NOP"
	exp := Instruction{
		Opcode: OpNOP,
	}

	checkParseInstructionSuccess(t, code, exp)
}

func TestParseInstructionWithOneRegisterOperand(t *testing.T) {
	code := "ADD ACC"
	exp := Instruction{
		Opcode:   OpADD,
		Oprands1: RegisterAcc,
	}

	checkParseInstructionSuccess(t, code, exp)
}

func TestParseInstructionWithOneLiteralOprand(t *testing.T) {
	code := "SUB 42"
	exp := Instruction{
		Opcode:   OpSUB,
		Oprands1: Literal(42),
	}

	checkParseInstructionSuccess(t, code, exp)
}

func TestParseInstructionWithOneLabelOprand(t *testing.T) {
	code := "JMP LOOP"
	exp := Instruction{
		Opcode:   OpJMP,
		Oprands1: Label("LOOP"),
	}

	checkParseInstructionSuccess(t, code, exp)
}

func TestParseInstructionWithTwoOprands(t *testing.T) {
	code := "MOV ACC, LEFT"
	exp := Instruction{
		Opcode:   OpMOV,
		Oprands1: RegisterAcc,
		Oprands2: RegisterLeft,
	}

	checkParseInstructionSuccess(t, code, exp)
}

func TestParseInstructionWithTwoOprandsNoComma(t *testing.T) {
	code := "MOV ACC LEFT"
	exp := Instruction{
		Opcode:   OpMOV,
		Oprands1: RegisterAcc,
		Oprands2: RegisterLeft,
	}

	checkParseInstructionSuccess(t, code, exp)
}

func TestParseInstructionWithCommas(t *testing.T) {
	codes := []string{
		"MOV,ACC,LEFT",
		",MOV,ACC,LEFT",
		"MOV,ACC,LEFT,",
		",MOV,ACC,LEFT,",
		",,,MOV,,,,ACC,,,,LEFT,,,",
	}

	exp := Instruction{
		Opcode:   OpMOV,
		Oprands1: RegisterAcc,
		Oprands2: RegisterLeft,
	}

	for _, code := range codes {
		checkParseInstructionSuccess(t, code, exp)
	}
}

func TestParseInstructionWithOnlyLabel(t *testing.T) {
	code := "LOOP:"
	exp := Instruction{
		Label:  "LOOP",
		Opcode: OpEmpty,
	}

	checkParseInstructionSuccess(t, code, exp)
}

func TestParseInstructionWithLabelAndInstruction0(t *testing.T) {
	code := "LOOP: NOP"
	exp := Instruction{
		Label:  "LOOP",
		Opcode: OpNOP,
	}

	checkParseInstructionSuccess(t, code, exp)
}

func TestParseInstructionWithLabelAndInstruction1(t *testing.T) {
	code := "START: ADD ACC"
	exp := Instruction{
		Label:    "START",
		Opcode:   OpADD,
		Oprands1: RegisterAcc,
	}

	checkParseInstructionSuccess(t, code, exp)
}

func TestParseInstructionWithLabelAndInstruction2(t *testing.T) {
	code := "LOOP: MOV ACC, LEFT"
	exp := Instruction{
		Label:    "LOOP",
		Opcode:   OpMOV,
		Oprands1: RegisterAcc,
		Oprands2: RegisterLeft,
	}

	checkParseInstructionSuccess(t, code, exp)
}

func TestParseInstructionErrorWithInvalidOpcode1(t *testing.T) {
	code := "LOREM IPSUM"
	errMessage := []string{
		"LOREM IPSUM",
		"^^^^^",
		`INVALID OPCODE "LOREM"`,
	}

	checkParseInstructionError(t, code, errMessage)
}

func TestParseInstructionErrorWithInvalidOpcode2(t *testing.T) {
	code := "    LOREM IPSUM"
	errMessage := []string{
		"    LOREM IPSUM",
		"    ^^^^^",
		`    INVALID OPCODE "LOREM"`,
	}

	checkParseInstructionError(t, code, errMessage)
}

func TestParseInstructionErrorWithTooManyOperands0(t *testing.T) {
	code := "NOP ACC"
	errMessage := []string{
		"NOP ACC",
		"    ^^^",
		"    TOO MANY OPERANDS",
	}

	checkParseInstructionError(t, code, errMessage)
}

func TestParseInstructionErrorWithTooManyOperands1(t *testing.T) {
	code := "ADD ACC, LEFT, RIGHT"
	errMessage := []string{
		"ADD ACC, LEFT, RIGHT",
		"         ^^^^^^^^^^^",
		"         TOO MANY OPERANDS",
	}

	checkParseInstructionError(t, code, errMessage)
}

func TestParseInstructionErrorWithTooManyOperands2(t *testing.T) {
	code := "MOV ACC, LEFT, RIGHT, UP"
	errMessage := []string{
		"MOV ACC, LEFT, RIGHT, UP",
		"               ^^^^^^^^^",
		"               TOO MANY OPERANDS",
	}

	checkParseInstructionError(t, code, errMessage)
}

func TestParseInstructionErrorWithTooManyOperandsAndInvalidOpcode(t *testing.T) {
	code := "MV ACC, LEFT, RIGHT, UP"
	errMessage := []string{
		"MV ACC, LEFT, RIGHT, UP",
		"^^",
		`INVALID OPCODE "MV"`,
	}

	checkParseInstructionError(t, code, errMessage)
}

func TestParseInstructionErrorWithTooManyOperandsAndInvalidOprand(t *testing.T) {
	code := "MOV AC, LEFT, RIGHT, UP"
	errMessage := []string{
		"MOV AC, LEFT, RIGHT, UP",
		"              ^^^^^^^^^",
		`              TOO MANY OPERANDS`,
	}

	checkParseInstructionError(t, code, errMessage)
}

func TestParseInstructionErrorWithMissingOperand1(t *testing.T) {
	code := "ADD"
	errMessage := []string{
		"ADD",
		"^^^",
		"MISSING OPERAND",
	}

	checkParseInstructionError(t, code, errMessage)
}

func TestParseInstructionErrorWithMissingOperand1AndTrailingSpace(t *testing.T) {
	code := "ADD  "
	errMessage := []string{
		"ADD  ",
		"^^^^^",
		"MISSING OPERAND",
	}

	checkParseInstructionError(t, code, errMessage)
}

func TestParseInstructionErrorWithMissingOperand2(t *testing.T) {
	code := "MOV ACC"
	errMessage := []string{
		"MOV ACC",
		"^^^^^^^",
		"MISSING OPERAND",
	}

	checkParseInstructionError(t, code, errMessage)
}
