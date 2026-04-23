package vm

import (
	"testing"

	"strings"
)

func checkInstructionContexts(t *testing.T, ins Instruction) {
	t.Helper()

	if ins.Label != "" && ins.LabelCtx == nil {
		t.Errorf("Instruction with label '%s' should have LabelCtx, but got nil", ins.Label)
	}

	if !ins.Empty() && ins.OpCodeCtx == nil {
		t.Errorf("Instruction with opcode %d should have OpCodeCtx, but got nil", ins.Opcode)
	}

	if ins.Oprand1 != nil && ins.Oprand1Ctx == nil {
		t.Errorf("Instruction with Oprand1 should have Oprand1Ctx, but got nil")
	}

	if ins.Oprand2 != nil && ins.Oprand2Ctx == nil {
		t.Errorf("Instruction with Oprand2 should have Oprand2Ctx, but got nil")
	}
}

func checkParseInstructionSuccess(t *testing.T, code string, expected Instruction) Instruction {
	t.Helper()

	got, err := ParseInstruction([]rune(code))
	if err != nil {
		t.Fatalf("ParseInstruction() error:\n%s", err)
	}

	if !got.Equals(expected) {
		t.Errorf("Got wrong instruction on: %s", code)
		t.Errorf("ParseInstruction() = %v, expect %v", got, expected)
	}

	checkInstructionContexts(t, got)

	return got
}

func checkParseInstructionError(t *testing.T, code string, errMessage []string) {
	t.Helper()

	got, err := ParseInstruction([]rune(code))
	if got.Opcode != OpInvalid || err == nil {
		t.Errorf("ParseInstruction() expected nil and error, got %v", got)
		t.Fatalf("%s", err)
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
		Opcode:  OpADD,
		Oprand1: RegisterAcc,
	}

	checkParseInstructionSuccess(t, code, exp)
}

func TestParseInstructionWithOneValueOprand(t *testing.T) {
	code := "SUB 42"
	exp := Instruction{
		Opcode:  OpSUB,
		Oprand1: Value(42),
	}

	checkParseInstructionSuccess(t, code, exp)
}

func TestParseInstructionWithOneLabelOprand(t *testing.T) {
	code := "JMP LOOP"
	exp := Instruction{
		Opcode:  OpJMP,
		Oprand1: Label("LOOP"),
	}

	checkParseInstructionSuccess(t, code, exp)
}

func TestParseInstructionWithTwoOprands(t *testing.T) {
	code := "MOV ACC, LEFT"
	exp := Instruction{
		Opcode:  OpMOV,
		Oprand1: RegisterAcc,
		Oprand2: RegisterLeft,
	}

	checkParseInstructionSuccess(t, code, exp)
}

func TestParseInstructionWithTwoOprandsNoComma(t *testing.T) {
	code := "MOV ACC LEFT"
	exp := Instruction{
		Opcode:  OpMOV,
		Oprand1: RegisterAcc,
		Oprand2: RegisterLeft,
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
		Opcode:  OpMOV,
		Oprand1: RegisterAcc,
		Oprand2: RegisterLeft,
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
		Label:   "START",
		Opcode:  OpADD,
		Oprand1: RegisterAcc,
	}

	checkParseInstructionSuccess(t, code, exp)
}

func TestParseInstructionWithLabelAndInstruction2(t *testing.T) {
	code := "LOOP: MOV ACC, LEFT"
	exp := Instruction{
		Label:   "LOOP",
		Opcode:  OpMOV,
		Oprand1: RegisterAcc,
		Oprand2: RegisterLeft,
	}

	checkParseInstructionSuccess(t, code, exp)
}

func TestParseInstructionWithBreakpointInLabel(t *testing.T) {
	codes := []string{
		"!LOOP:",
		"LOOP:!",
	}

	exp := Instruction{
		Breakpoint: true,
		Label:      "LOOP",
		Opcode:     OpEmpty,
	}

	for _, code := range codes {
		checkParseInstructionSuccess(t, code, exp)
	}
}

func TestParseInstructionWithBreakpointInInstruction0(t *testing.T) {
	codes := []string{
		"NOP!",
		"!NOP",
	}

	exp := Instruction{
		Breakpoint: true,
		Opcode:     OpNOP,
	}

	for _, code := range codes {
		checkParseInstructionSuccess(t, code, exp)
	}
}

func TestParseInstructionWithBreakpointInInstruction1Register(t *testing.T) {
	codes := []string{
		"!ADD ACC",
		"ADD! ACC",
		"ADD !ACC",
		"ADD ACC!",
	}

	exp := Instruction{
		Breakpoint: true,
		Opcode:     OpADD,
		Oprand1:    RegisterAcc,
	}

	for _, code := range codes {
		checkParseInstructionSuccess(t, code, exp)
	}
}

func TestParseInstructionWithBreakpointInInstruction1Value(t *testing.T) {
	codes := []string{
		"!SUB 42",
		"SUB! 42",
		"SUB !42",
		"SUB 42!",
	}

	exp := Instruction{
		Breakpoint: true,
		Opcode:     OpSUB,
		Oprand1:    Value(42),
	}

	for _, code := range codes {
		checkParseInstructionSuccess(t, code, exp)
	}
}

func TestParseInstructionWithBreakpointInInstruction1Label(t *testing.T) {
	codes := []string{
		"!JMP LOOP",
		"JMP! LOOP",
		"JMP !LOOP",
		"JMP LOOP!",
	}

	exp := Instruction{
		Breakpoint: true,
		Opcode:     OpJMP,
		Oprand1:    Label("LOOP"),
	}

	for _, code := range codes {
		checkParseInstructionSuccess(t, code, exp)
	}
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

func TestParseInstructionErrorWithInvalidExpression1(t *testing.T) {
	code := "ADD AC"
	errMessage := []string{
		"ADD AC",
		"    ^^",
		"    INVALID EXPRESSION \"AC\"",
	}

	checkParseInstructionError(t, code, errMessage)
}

func TestParseInstructionErrorWithInvalidExpression2(t *testing.T) {
	code := "MOV ACC, 4O"
	errMessage := []string{
		"MOV ACC, 4O",
		"         ^^",
		"         INVALID EXPRESSION \"4O\"",
	}

	checkParseInstructionError(t, code, errMessage)
}

func TestParseInstructionErrorWithBreakpointInLastOfLabel(t *testing.T) {
	code := "LOOP!:"
	errMessage := []string{
		"LOOP!:",
		"^^^^",
		`INVALID OPCODE "LOOP"`,
	}

	checkParseInstructionError(t, code, errMessage)
}

func TestParseInstructionErrorWithBreakpointInOprand1(t *testing.T) {
	code := "ADD A!CC"
	errMessage := []string{
		"ADD A!CC",
		"      ^^",
		"      TOO MANY OPERANDS",
	}

	checkParseInstructionError(t, code, errMessage)
}

func TestParseInstructionErrorWithBreakpointInOprand2(t *testing.T) {
	code := "MOV ACC, L!EFT"
	errMessage := []string{
		"MOV ACC, L!EFT",
		"           ^^^",
		"           TOO MANY OPERANDS",
	}

	checkParseInstructionError(t, code, errMessage)
}

func TestParseInstructionErrorWithMultipleBreakpointsOnLabel1(t *testing.T) {
	code := "!LOOP:!"
	errMessage := []string{
		"!LOOP:!",
		"      ^",
		`      INVALID OPCODE "!"`,
	}

	checkParseInstructionError(t, code, errMessage)
}

func TestParseInstructionErrorWithMultipleBreakpointsOnLabel2(t *testing.T) {
	code := "LOOP:!!"
	errMessage := []string{
		"LOOP:!!",
		"      ^",
		`      INVALID OPCODE "!"`,
	}

	checkParseInstructionError(t, code, errMessage)
}

func TestParseInstructionErrorWithMultipleBreakpoints0(t *testing.T) {
	code := "!NOP!"
	errMessage := []string{
		"!NOP!",
		" ^^^^",
		` INVALID OPCODE "NOP!"`,
	}

	checkParseInstructionError(t, code, errMessage)
}

func TestParseInstructionErrorWithMultipleBreakpoints1(t *testing.T) {
	code := "ADD !ACC!"
	errMessage := []string{
		"ADD !ACC!",
		"     ^^^^",
		"     INVALID EXPRESSION \"ACC!\"",
	}

	checkParseInstructionError(t, code, errMessage)
}

func TestParseInstructionErrorWithMultipleBreakpoints2(t *testing.T) {
	code := "MOV !ACC, LEFT!"
	errMessage := []string{
		"MOV !ACC, LEFT!",
		"          ^^^^^",
		"          INVALID EXPRESSION \"LEFT!\"",
	}

	checkParseInstructionError(t, code, errMessage)
}

func TestInstructionContexts(t *testing.T) {
	code := "LOOP: MOV ACC, LEFT"
	exp := Instruction{
		Label:   "LOOP",
		Opcode:  OpMOV,
		Oprand1: RegisterAcc,
		Oprand2: RegisterLeft,
	}

	ins := checkParseInstructionSuccess(t, code, exp)

	tag := strings.Join([]string{
		"LOOP: MOV ACC, LEFT",
		"^^^^",
		"HERE",
	}, "\n")
	if ins.LabelCtx == nil || ins.LabelCtx.Message("HERE") != tag {
		t.Errorf("Label context is wrong, got:\n%s\nexpect:\n%s", ins.LabelCtx.Message("HERE"), tag)
	}

	opcode := strings.Join([]string{
		"LOOP: MOV ACC, LEFT",
		"      ^^^",
		"      HERE",
	}, "\n")

	if ins.OpCodeCtx == nil || ins.OpCodeCtx.Message("HERE") != opcode {
		t.Errorf("Opcode context is wrong, got:\n%s\nexpect:\n%s", ins.OpCodeCtx.Message("HERE"), opcode)
	}

	oprand1 := strings.Join([]string{
		"LOOP: MOV ACC, LEFT",
		"          ^^^",
		"          HERE",
	}, "\n")

	if ins.Oprand1Ctx == nil || ins.Oprand1Ctx.Message("HERE") != oprand1 {
		t.Errorf("Oprand1 context is wrong, got:\n%s\nexpect:\n%s", ins.Oprand1Ctx.Message("HERE"), oprand1)
	}

	oprand2 := strings.Join([]string{
		"LOOP: MOV ACC, LEFT",
		"               ^^^^",
		"               HERE",
	}, "\n")

	if ins.Oprand2Ctx == nil || ins.Oprand2Ctx.Message("HERE") != oprand2 {
		t.Errorf("Oprand2 context is wrong, got:\n%s\nexpect:\n%s", ins.Oprand2Ctx.Message("HERE"), oprand2)
	}
}

func TestParseCodeSuccess(t *testing.T) {
	code := strings.Join([]string{
		"MOV UP, DOWN",
		"ADD ACC",
	}, "\n")

	exp := Code{
		{
			Opcode:  OpMOV,
			Oprand1: RegisterUp,
			Oprand2: RegisterDown,
		},
		{
			Opcode:  OpADD,
			Oprand1: RegisterAcc,
		},
	}

	got, n, err := ParseCode(code)
	if n != -1 || err != nil {
		t.Fatalf("ParseCode() error at line %d:\n%s", n, err)
	}

	if !exp.Equals(got) {
		t.Errorf("ParseCode() = %v, expect %v", got, exp)
	}
}

func TestParseCodeError(t *testing.T) {
	code := strings.Join([]string{
		"MOV UP, DOWN",
		"ADD A!CC",
		"SUB 42",
	}, "\n")

	errMessage := []string{
		"ADD A!CC",
		"      ^^",
		"      TOO MANY OPERANDS",
	}

	got, n, err := ParseCode(code)
	if n != 1 || err == nil {
		t.Fatalf("ParseCode() expected error at line 1, got line %d and error:\n%s", n, err)
	}

	if got != nil {
		t.Fatalf("ParseCode() expected nil, got %v", got)
	}

	gotMessage := err.Error()
	expected := strings.Join(errMessage, "\n")
	if gotMessage != expected {
		t.Errorf("wrong error message, got:\n%s\nexpect:\n%s", gotMessage, expected)
	}
}
