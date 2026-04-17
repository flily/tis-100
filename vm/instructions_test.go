package vm

import (
	"testing"

	"strings"
)

func checkOprandType(t *testing.T, oprand Oprand, expect OprandType) {
	t.Helper()

	if oprand.OprandType() != expect {
		t.Errorf("expect OprandType %d, got %d", expect, oprand.OprandType())
	}
}

func TestOperandTypeInclude(t *testing.T) {
	types := OprandRegister | OprandLiteral

	if !types.Include(OprandRegister) {
		t.Errorf("expect OprandType to include OprandRegister")
	}

	if !types.Include(OprandLiteral) {
		t.Errorf("expect OprandType to include OprandLiteral")
	}

	if types.Include(OprandLabel) {
		t.Errorf("expect OprandType to not include OprandLabel")
	}
}

func TestOpCodeNames(t *testing.T) {
	for code, name := range opCodeNames {
		if code.String() != name {
			t.Errorf("OpCode %d: expect name '%s', got '%s'", code, name, code.String())
		}

		if !strings.HasPrefix(name, "#") && NewOpcode(name) != code {
			t.Errorf("OpCode name '%s': expect code %d, got %d", name, code, NewOpcode(name))
		}
	}
}

func TestInvalidOpCode(t *testing.T) {
	invalidCode := Opcode(999)
	if invalidCode.String() != InvalidOpcodeName {
		t.Errorf("Invalid OpCode: expect name '%s', got '%s'", InvalidOpcodeName, invalidCode.String())
	}

	invalidOpCode := "UNKNOWN"
	if NewOpcode(invalidOpCode) != InvalidOpCode {
		t.Errorf("Invalid OpCode name '%s': expect code %d, got %d", invalidOpCode, InvalidOpCode, NewOpcode(invalidOpCode))
	}
}

func TestRegisterNames(t *testing.T) {
	for reg, name := range registerNames {
		checkOprandType(t, reg, OprandRegister)

		if reg.String() != name {
			t.Errorf("Register %d: expect name '%s', got '%s'", reg, name, reg.String())
		}

		if NewRegister(name) != reg {
			t.Errorf("Register name '%s': expect register %d, got %d", name, reg, NewRegister(name))
		}
	}
}

func TestInvalidRegister(t *testing.T) {
	invalidReg := Register(999)

	checkOprandType(t, invalidReg, OprandRegister)

	if invalidReg.String() != InvalidRegisterName {
		t.Errorf("Invalid Register: expect name '%s', got '%s'", InvalidRegisterName, invalidReg.String())
	}

	invalidRegName := "UNKNOWN"
	if NewRegister(invalidRegName) != InvalidRegister {
		t.Errorf("Invalid Register name '%s': expect register %d, got %d", invalidRegName, InvalidRegister, NewRegister(invalidRegName))
	}
}

func TestRegisterEqual(t *testing.T) {
	cases := []struct {
		reg      Register
		oprand   Oprand
		expected bool
	}{
		{reg: RegisterAcc, oprand: RegisterAcc, expected: true},
		{reg: RegisterAcc, oprand: RegisterBak, expected: false},
		{reg: RegisterAcc, oprand: Literal(42), expected: false},
		{reg: RegisterAcc, oprand: NewLabel("LOOP"), expected: false},
	}

	for _, c := range cases {
		if c.reg.Equal(c.oprand) != c.expected {
			t.Errorf("Register %d Equal Oprand %v: expect %t, got %t", c.reg, c.oprand, c.expected, c.reg.Equal(c.oprand))
		}
	}
}

func TestLabel(t *testing.T) {
	labelName := "LOOP"
	label := NewLabel(labelName)

	checkOprandType(t, label, OprandLabel)

	if label.String() != labelName {
		t.Errorf("Label: expect name '%s', got '%s'", labelName, label.String())
	}
}

func TestLabelEqual(t *testing.T) {
	cases := []struct {
		label    Label
		oprand   Oprand
		expected bool
	}{
		{label: NewLabel("LOOP"), oprand: NewLabel("LOOP"), expected: true},
		{label: NewLabel("LOOP"), oprand: NewLabel("END"), expected: false},
		{label: NewLabel("LOOP"), oprand: RegisterAcc, expected: false},
		{label: NewLabel("LOOP"), oprand: Literal(42), expected: false},
	}

	for _, c := range cases {
		if c.label.Equal(c.oprand) != c.expected {
			t.Errorf("Label '%s' Equal Oprand %v: expect %t, got %t", c.label, c.oprand, c.expected, c.label.Equal(c.oprand))
		}
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
		t.Errorf("Literal: expect value %d, got %d", literalValue, literal)
	}

	if literal.String() != literalStr {
		t.Errorf("Literal: expect string '%s', got '%s'", literalStr, literal.String())
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
		t.Errorf("Negative Literal: expect value %d, got %d", literalValue, literal)
	}

	if literal.String() != literalStr {
		t.Errorf("Negative Literal: expect string '%s', got '%s'", literalStr, literal.String())
	}
}

func TestInvalidLiteral(t *testing.T) {
	invalidLiteralStr := "abc"

	_, err := ParseLiteral(invalidLiteralStr)
	if err == nil {
		t.Fatalf("expect error when parsing invalid literal '%s', but got none", invalidLiteralStr)
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
			t.Errorf("Literal %d: expect InStandardRange() to return %t, got %t",
				c.value, c.expeceted, c.value.InStandardRange())
		}
	}
}

func TestLiteralEqual(t *testing.T) {
	cases := []struct {
		literal  Literal
		oprand   Oprand
		expected bool
	}{
		{literal: Literal(42), oprand: Literal(42), expected: true},
		{literal: Literal(42), oprand: Literal(43), expected: false},
		{literal: Literal(42), oprand: RegisterAcc, expected: false},
		{literal: Literal(42), oprand: NewLabel("LOOP"), expected: false},
	}

	for _, c := range cases {
		if c.literal.Equal(c.oprand) != c.expected {
			t.Errorf("Literal %d Equal Oprand %v: expect %t, got %t", c.literal, c.oprand, c.expected, c.literal.Equal(c.oprand))
		}
	}
}

func TestContextMarkAndMessage(t *testing.T) {
	content := []rune("LOREM IPSUM DOLOR SIT AMET")
	base := NewContext(content)

	ctx := base.Mark(12, 17)
	message := ctx.Message("consectetur adipiscing elit")
	expect := strings.Join([]string{
		"LOREM IPSUM DOLOR SIT AMET",
		"            ^^^^^",
		"            consectetur adipiscing elit",
	}, "\n")

	if message != expect {
		t.Errorf("Context Message: expect:\n%s\ngot:\n%s", expect, message)
	}
}
