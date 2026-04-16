package vm

import (
	"testing"
)

func checkOprandType(t *testing.T, oprand Oprand, expected OprandType) {
	t.Helper()

	if oprand.OprandType() != expected {
		t.Errorf("Expected OprandType %d, got %d", expected, oprand.OprandType())
	}
}

func TestOperandTypeInclude(t *testing.T) {
	types := OprandRegister | OprandLiteral

	if !types.Include(OprandRegister) {
		t.Errorf("Expected OprandType to include OprandRegister")
	}

	if !types.Include(OprandLiteral) {
		t.Errorf("Expected OprandType to include OprandLiteral")
	}

	if types.Include(OprandLabel) {
		t.Errorf("Expected OprandType to not include OprandLabel")
	}
}

func TestOpCodeNames(t *testing.T) {
	for code, name := range opCodeNames {
		if code.String() != name {
			t.Errorf("OpCode %d: expected name '%s', got '%s'", code, name, code.String())
		}

		if NewOpcode(name) != code {
			t.Errorf("OpCode name '%s': expected code %d, got %d", name, code, NewOpcode(name))
		}
	}
}

func TestInvalidOpCode(t *testing.T) {
	invalidCode := Opcode(999)
	if invalidCode.String() != InvalidOpcodeName {
		t.Errorf("Invalid OpCode: expected name '%s', got '%s'", InvalidOpcodeName, invalidCode.String())
	}

	invalidOpCode := "UNKNOWN"
	if NewOpcode(invalidOpCode) != InvalidOpCode {
		t.Errorf("Invalid OpCode name '%s': expected code %d, got %d", invalidOpCode, InvalidOpCode, NewOpcode(invalidOpCode))
	}
}

func TestRegisterNames(t *testing.T) {
	for reg, name := range registerNames {
		checkOprandType(t, reg, OprandRegister)

		if reg.String() != name {
			t.Errorf("Register %d: expected name '%s', got '%s'", reg, name, reg.String())
		}

		if NewRegister(name) != reg {
			t.Errorf("Register name '%s': expected register %d, got %d", name, reg, NewRegister(name))
		}
	}
}

func TestInvalidRegister(t *testing.T) {
	invalidReg := Register(999)

	checkOprandType(t, invalidReg, OprandRegister)

	if invalidReg.String() != InvalidRegisterName {
		t.Errorf("Invalid Register: expected name '%s', got '%s'", InvalidRegisterName, invalidReg.String())
	}

	invalidRegName := "UNKNOWN"
	if NewRegister(invalidRegName) != InvalidRegister {
		t.Errorf("Invalid Register name '%s': expected register %d, got %d", invalidRegName, InvalidRegister, NewRegister(invalidRegName))
	}
}

func TestLabel(t *testing.T) {
	labelName := "LOOP"
	label := NewLabel(labelName)

	checkOprandType(t, label, OprandLabel)

	if label.String() != labelName {
		t.Errorf("Label: expected name '%s', got '%s'", labelName, label.String())
	}
}

func TestLiteral(t *testing.T) {
	literalValue := 42
	literalStr := "42"

	literal, err := ParseLiteral(literalStr)
	if err != nil {
		t.Fatalf("Failed to parse literal: %v", err)
	}

	checkOprandType(t, literal, OprandLiteral)

	if int(literal) != literalValue {
		t.Errorf("Literal: expected value %d, got %d", literalValue, literal)
	}

	if literal.String() != literalStr {
		t.Errorf("Literal: expected string '%s', got '%s'", literalStr, literal.String())
	}
}

func TestNegativeLiteral(t *testing.T) {
	literalValue := -42
	literalStr := "-42"

	literal, err := ParseLiteral(literalStr)
	if err != nil {
		t.Fatalf("Failed to parse negative literal: %v", err)
	}

	checkOprandType(t, literal, OprandLiteral)

	if int(literal) != literalValue {
		t.Errorf("Negative Literal: expected value %d, got %d", literalValue, literal)
	}

	if literal.String() != literalStr {
		t.Errorf("Negative Literal: expected string '%s', got '%s'", literalStr, literal.String())
	}
}

func TestInvalidLiteral(t *testing.T) {
	invalidLiteralStr := "abc"

	_, err := ParseLiteral(invalidLiteralStr)
	if err == nil {
		t.Fatalf("Expected error when parsing invalid literal '%s', but got none", invalidLiteralStr)
	}
}

func TestLiteralRange(t *testing.T) {
	cases := []struct {
		value     Literal
		expeceted bool
	}{
		{value: 0, expeceted: true},
		{value: 42, expeceted: true},
		{value: -42, expeceted: true},
		{value: 999, expeceted: true},
		{value: -999, expeceted: true},
		{value: 1000, expeceted: false},
		{value: -1000, expeceted: false},
		{value: 12345, expeceted: false},
		{value: -12345, expeceted: false},
	}

	for _, c := range cases {
		if c.value.InStandardRange() != c.expeceted {
			t.Errorf("Literal %d: expected InStandardRange() to return %t, got %t", c.value, c.expeceted, c.value.InStandardRange())
		}
	}
}
