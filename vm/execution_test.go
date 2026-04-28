package vm

import (
	"testing"

	"slices"
	"strings"
)

type exeNodeTestCase struct {
	code       string
	acc        Value
	bak        Value
	ip         int
	mode       ExecutionMode
	ports      []IOPort
	portsValue []Value
}

type exeNodeTestCases []exeNodeTestCase

func (cases exeNodeTestCases) Run(t *testing.T, ip int, acc Value, bak Value) {
	t.Helper()

	node := NewExecutionNode()

	for _, c := range cases {
		if i, err := node.LoadCode(c.code); err != nil {
			t.Fatalf("LoadCode failed at instruction %d: %v", i, err)
		}

		node.Acc = acc
		node.Backup = bak

		if c.ports != nil {
			node.LoadPorts(c.ports)
		}

		err, _ := node.Step()
		if err != nil {
			t.Fatalf("Step failed: %v", err)
		}

		if node.Acc != c.acc {
			t.Errorf("Instruction failed: expect Acc %d, got %d", c.acc, node.Acc)
		}

		if node.Backup != c.bak {
			t.Errorf("Instruction failed: expect Backup %d, got %d", c.bak, node.Backup)
		}

		if node.IP != c.ip {
			t.Errorf("Instruction failed: expect IP %d, got %d", c.ip, node.IP)
		}

		if node.Mode != c.mode {
			t.Errorf("Instruction failed: expect Mode %v, got %v", c.mode, node.Mode)
		}

		if c.ports != nil {
			got := node.Snapshot()
			if !slices.Equal(got, c.portsValue) {
				t.Errorf("Instruction failed: ports not as expected")
				t.Errorf("expect: %v", c.portsValue)
				t.Errorf("got: %v", got)
			}
		}
	}
}

func noPortsData() []IOPort {
	ports := []IOPort{
		NewValuePort(),
		NewValuePort(),
		NewValuePort(),
		NewValuePort(),
	}

	return ports
}

func portsWithData(v1 Value, v2 Value, v3 Value, v4 Value) []IOPort {
	ports := []IOPort{
		NewValuePortWithValue(v1),
		NewValuePortWithValue(v2),
		NewValuePortWithValue(v3),
		NewValuePortWithValue(v4),
	}

	return ports
}

func codeLines(lines ...string) string {
	return strings.Join(lines, "\n")
}

func TestNodePortsRead(t *testing.T) {
	ports := NewNodePorts()
	ports.Ports = [4]IOPort{
		NewValuePortWithValue(21),
		NewValuePortWithValue(22),
		NewValuePortWithValue(23),
		NewValuePortWithValue(24),
	}

	result := make([]Value, 0, 10)
	for range 4 {
		v, ok := ports.ReadAny()
		if !ok {
			t.Errorf("ReadAny should return true when ports are not empty")
		}

		result = append(result, v)
	}

	ports.WriteDone()

	expected := []Value{22, 23, 21, 24}
	if !slices.Equal(result, expected) {
		t.Errorf("ReadAny should read ports in correct order")
		t.Errorf("expect: %v", expected)
		t.Errorf("got: %v", result)
	}
}

func TestNodePortsWrite(t *testing.T) {
	ports := NewNodePorts()
	ports.Ports = [4]IOPort{
		NewValuePort(),
		NewValuePort(),
		NewValuePort(),
		NewValuePort(),
	}

	valuesToWrite := []Value{31, 32, 33, 34}
	for _, v := range valuesToWrite {
		ok := ports.WriteAny(v)
		if !ok {
			t.Errorf("WriteAny should return true when ports are not full")
		}
	}
	ports.WriteDone()

	expected := []Value{31, 32, 33, 34}
	got := make([]Value, 0, 4)
	for _, port := range ports.Ports {
		v, ok := port.Read()
		if !ok {
			t.Errorf("Port should have value after write")
		}

		got = append(got, v)
	}

	if !slices.Equal(got, expected) {
		t.Errorf("WriteAny should write ports in correct order")
		t.Errorf("expect: %v", expected)
		t.Errorf("got: %v", got)
	}
}

func TestExecutionNodeLoadCode(t *testing.T) {
	code := strings.Join([]string{
		"MOV UP, ACC",
		"ADD 42",
		"MOV ACC, DOWN",
	}, "\n")

	node := NewExecutionNode()
	if i, err := node.LoadCode(code); err != nil {
		t.Errorf("LoadCode failed at instruction %d: %v", i, err)
	}

	expectedCodes := Code{
		{
			Opcode:  OpMOV,
			Oprand1: RegisterUp,
			Oprand2: RegisterAcc,
		},
		{
			Opcode:  OpADD,
			Oprand1: Value(42),
		},
		{
			Opcode:  OpMOV,
			Oprand1: RegisterAcc,
			Oprand2: RegisterDown,
		},
	}

	if !expectedCodes.Equals(node.Codes) {
		t.Errorf("Load code failure")
		t.Errorf("expect: %v", expectedCodes)
		t.Errorf("got: %v", node.Codes)
	}

	if node.IP != 0 {
		t.Errorf("Load code failure: expect IP 0, got %d", node.IP)
	}

	if node.Acc != 0 {
		t.Errorf("Load code failure: expect Acc 0, got %d", node.Acc)
	}

	if node.Last != RegisterNA {
		t.Errorf("Load code failure: expect Last RegisterNA, got %v", node.Last)
	}

	if node.Mode != ModeIdle {
		t.Errorf("Load code failure: expect ModeIdle, got %v", node.Mode)
	}
}

func TestExecutionNodeLoadLabels(t *testing.T) {
	code := strings.Join([]string{
		"NOP",       // 0
		"LAB1:",     // 1
		"LAB2: NOP", // 2
		"NOP",       // 3
		"",          // 4
		"LAB3:",     // 5
		"NOP",       // 6
	}, "\n")

	node := NewExecutionNode()
	if i, err := node.LoadCode(code); err != nil {
		t.Errorf("LoadCode failed at instruction %d: %v", i, err)
	}

	expectedLabels := map[Label]int{
		"LAB1": 1,
		"LAB2": 2,
		"LAB3": 5,
	}

	gotLabels := node.GetLabels()
	if len(gotLabels) != len(expectedLabels) {
		t.Errorf("LoadLabels failure: expect %d labels, got %d", len(expectedLabels), len(gotLabels))
	}

	for label, idx := range expectedLabels {
		if node.GetLabel(label) != idx {
			t.Errorf("LoadLabels failure: expect label '%s' at instruction %d, got %d", label, idx, node.GetLabel(label))
		}
	}

	if idx := node.GetLabel("LOREM"); idx != -1 {
		t.Errorf("LoadLabels failure: expect label 'LOREM' not found, got index %d", idx)
	}
}

func TestExecutionNodeStep(t *testing.T) {
	code := strings.Join([]string{
		"NOP",
		"NOP",
		"NOP",
	}, "\n")

	node := NewExecutionNode()
	if i, err := node.LoadCode(code); err != nil {
		t.Errorf("LoadCode failed at instruction %d: %v", i, err)
	}

	err, looped := node.Step()
	if err != nil {
		t.Errorf("Step failed: %v", err)
	}

	if looped {
		t.Errorf("Step should not loop when there are more instructions")
	}

	if node.IP != 1 {
		t.Errorf("Step failure: expect IP 1, got %d", node.IP)
	}

	err, looped = node.Step()
	if err != nil {
		t.Errorf("Step failed: %v", err)
	}

	if looped {
		t.Errorf("Step should not loop when there are more instructions")
	}

	if node.IP != 2 {
		t.Errorf("Step failure: expect IP 2, got %d", node.IP)
	}

	err, looped = node.Step()
	if err != nil {
		t.Errorf("Step failed: %v", err)
	}

	if !looped {
		t.Errorf("Step should loop when reaching the end of instructions")
	}

	if node.IP != 0 {
		t.Errorf("Step failure: expect IP 0 after looping, got %d", node.IP)
	}
}

func TestExecutionNodeOpNop(t *testing.T) {
	defaultAcc := Value(13)
	exeNodeTestCases{
		{
			code: "NOP",
			acc:  defaultAcc,
			bak:  0,
			mode: ModeIdle,
		},
	}.Run(t, 0, defaultAcc, 0)
}

func TestExecutionNodeOpMovBasicIO(t *testing.T) {
	defaultAcc := Value(13)

	exeNodeTestCases{
		{
			code: "MOV 42, ACC",
			acc:  42,
			bak:  0,
			mode: ModeIdle,
		},
		{
			code:       "MOV ACC, UP",
			acc:        13,
			bak:        0,
			mode:       ModeIdle,
			ports:      noPortsData(),
			portsValue: []Value{13, 0, 0, 0},
		},
		{
			code:  "MOV ACC, LEFT",
			acc:   13,
			bak:   0,
			mode:  ModeIdle,
			ports: noPortsData(),

			portsValue: []Value{0, 13, 0, 0},
		},
		{
			code:       "MOV ACC, RIGHT",
			acc:        13,
			bak:        0,
			mode:       ModeIdle,
			ports:      noPortsData(),
			portsValue: []Value{0, 0, 13, 0},
		},
		{
			code:       "MOV ACC, DOWN",
			acc:        13,
			bak:        0,
			mode:       ModeIdle,
			ports:      noPortsData(),
			portsValue: []Value{0, 0, 0, 13},
		},
		{
			code:       "MOV UP, ACC",
			acc:        3,
			bak:        0,
			mode:       ModeIdle,
			ports:      portsWithData(3, 5, 7, 11),
			portsValue: []Value{3, 5, 7, 11},
		},
		{
			code:       "MOV LEFT, ACC",
			acc:        5,
			bak:        0,
			mode:       ModeIdle,
			ports:      portsWithData(3, 5, 7, 11),
			portsValue: []Value{3, 5, 7, 11},
		},
		{
			code:       "MOV RIGHT, ACC",
			acc:        7,
			bak:        0,
			mode:       ModeIdle,
			ports:      portsWithData(3, 5, 7, 11),
			portsValue: []Value{3, 5, 7, 11},
		},
		{
			code:       "MOV DOWN, ACC",
			acc:        11,
			bak:        0,
			mode:       ModeIdle,
			ports:      portsWithData(3, 5, 7, 11),
			portsValue: []Value{3, 5, 7, 11},
		},
	}.Run(t, 0, defaultAcc, 0)
}

func TestExecutionNodeOpMovBlocking(t *testing.T) {
	defaultAcc := Value(13)

	exeNodeTestCases{
		{
			code:       "MOV UP, ACC",
			acc:        13,
			bak:        0,
			mode:       ModeRead,
			ports:      noPortsData(),
			portsValue: []Value{0, 0, 0, 0},
		},
		{
			code:       "MOV LEFT, ACC",
			acc:        13,
			bak:        0,
			mode:       ModeRead,
			ports:      noPortsData(),
			portsValue: []Value{0, 0, 0, 0},
		},
		{
			code:       "MOV RIGHT, ACC",
			acc:        13,
			bak:        0,
			mode:       ModeRead,
			ports:      noPortsData(),
			portsValue: []Value{0, 0, 0, 0},
		},
		{
			code:       "MOV DOWN, ACC",
			acc:        13,
			bak:        0,
			mode:       ModeRead,
			ports:      noPortsData(),
			portsValue: []Value{0, 0, 0, 0},
		},
		{
			code:       "MOV ACC, UP",
			acc:        13,
			bak:        0,
			mode:       ModeWrite,
			ports:      portsWithData(3, 5, 7, 11),
			portsValue: []Value{3, 5, 7, 11},
		},
		{
			code:       "MOV ACC, LEFT",
			acc:        13,
			bak:        0,
			mode:       ModeWrite,
			ports:      portsWithData(3, 5, 7, 11),
			portsValue: []Value{3, 5, 7, 11},
		},
		{
			code:       "MOV ACC, RIGHT",
			acc:        13,
			bak:        0,
			mode:       ModeWrite,
			ports:      portsWithData(3, 5, 7, 11),
			portsValue: []Value{3, 5, 7, 11},
		},
		{
			code:       "MOV ACC, DOWN",
			acc:        13,
			bak:        0,
			mode:       ModeWrite,
			ports:      portsWithData(3, 5, 7, 11),
			portsValue: []Value{3, 5, 7, 11},
		},
	}.Run(t, 0, defaultAcc, 0)
}

func TestExecutionNodeOpSwp(t *testing.T) {
	defaultAcc := Value(13)
	defaultBak := Value(17)

	exeNodeTestCases{
		{
			code: "SWP",
			acc:  defaultBak,
			bak:  defaultAcc,
			mode: ModeIdle,
		},
	}.Run(t, 0, defaultAcc, defaultBak)
}

func TestExecutionNodeOpSav(t *testing.T) {
	defaultAcc := Value(13)
	defaultBak := Value(17)

	exeNodeTestCases{
		{
			code: "SAV",
			acc:  defaultAcc,
			bak:  defaultAcc,
			mode: ModeIdle,
		},
	}.Run(t, 0, defaultAcc, defaultBak)
}

func TestExecutionNodeOpAdd(t *testing.T) {
	defaultAcc := Value(13)

	exeNodeTestCases{
		{
			code: "ADD 42",
			acc:  defaultAcc + 42,
			bak:  0,
			mode: ModeIdle,
		},
		{
			code: "ADD 990",
			acc:  ValueMax,
			bak:  0,
			mode: ModeIdle,
		},
		{
			code: "ADD -42",
			acc:  defaultAcc - 42,
			bak:  0,
			mode: ModeIdle,
		},
	}.Run(t, 0, defaultAcc, 0)
}

func TestExecutionNodeOpSub(t *testing.T) {
	defaultAcc := Value(-13)

	exeNodeTestCases{
		{
			code: "SUB 42",
			acc:  defaultAcc - 42,
			bak:  0,
			mode: ModeIdle,
		},
		{
			code: "SUB 990",
			acc:  ValueMin,
			bak:  0,
			mode: ModeIdle,
		},
		{
			code: "SUB -42",
			acc:  defaultAcc + 42,
			bak:  0,
			mode: ModeIdle,
		},
	}.Run(t, 0, defaultAcc, 0)
}

func TestExecutionNodeOpNeg(t *testing.T) {
	defaultAcc := Value(13)

	exeNodeTestCases{
		{
			code: "NEG",
			acc:  -defaultAcc,
			bak:  0,
			mode: ModeIdle,
		},
	}.Run(t, 0, defaultAcc, 0)
}

func TestExecutionNodeOpJmp(t *testing.T) {
	defaultAcc := Value(13)

	exeNodeTestCases{
		{
			code: codeLines(
				"JMP LAB2",
				"LAB1: NOP",
				"LAB2: NOP",
				"LAB3: NOP",
			),
			acc:  defaultAcc,
			bak:  0,
			ip:   2,
			mode: ModeIdle,
		},
	}.Run(t, 0, defaultAcc, 0)
}

func TestExecutionNodeOpJez(t *testing.T) {
	exeNodeTestCases{
		{
			code: codeLines(
				"JEZ LAB2",
				"LAB1: NOP",
				"LAB2: NOP",
				"LAB3: NOP",
			),
			acc:  0,
			bak:  0,
			ip:   2,
			mode: ModeIdle,
		},
	}.Run(t, 0, 0, 0)

	exeNodeTestCases{
		{
			code: codeLines(
				"JEZ LAB2",
				"LAB1: NOP",
				"LAB2: NOP",
				"LAB3: NOP",
			),
			acc:  42,
			bak:  0,
			ip:   1,
			mode: ModeIdle,
		},
	}.Run(t, 0, 42, 0)
}

func TestExecutionNodeOpJnz(t *testing.T) {
	exeNodeTestCases{
		{
			code: codeLines(
				"JNZ LAB2",
				"LAB1: NOP",
				"LAB2: NOP",
				"LAB3: NOP",
			),
			acc:  0,
			bak:  0,
			ip:   1,
			mode: ModeIdle,
		},
	}.Run(t, 0, 0, 0)

	exeNodeTestCases{
		{
			code: codeLines(
				"JNZ LAB2",
				"LAB1: NOP",
				"LAB2: NOP",
				"LAB3: NOP",
			),
			acc:  42,
			bak:  0,
			ip:   2,
			mode: ModeIdle,
		},
	}.Run(t, 0, 42, 0)
}

func TestExecutionNodeOpJgz(t *testing.T) {
	exeNodeTestCases{
		{
			code: codeLines(
				"JGZ LAB2",
				"LAB1: NOP",
				"LAB2: NOP",
				"LAB3: NOP",
			),
			acc:  0,
			bak:  0,
			ip:   1,
			mode: ModeIdle,
		},
	}.Run(t, 0, 0, 0)

	exeNodeTestCases{
		{
			code: codeLines(
				"JGZ LAB2",
				"LAB1: NOP",
				"LAB2: NOP",
				"LAB3: NOP",
			),
			acc:  42,
			bak:  0,
			ip:   2,
			mode: ModeIdle,
		},
	}.Run(t, 0, 42, 0)

	exeNodeTestCases{
		{
			code: codeLines(
				"JGZ LAB2",
				"LAB1: NOP",
				"LAB2: NOP",
				"LAB3: NOP",
			),
			acc:  -42,
			bak:  0,
			ip:   1,
			mode: ModeIdle,
		},
	}.Run(t, 0, -42, 0)
}

func TestExecutionNodeOpJlz(t *testing.T) {
	exeNodeTestCases{
		{
			code: codeLines(
				"JLZ LAB2",
				"LAB1: NOP",
				"LAB2: NOP",
				"LAB3: NOP",
			),
			acc:  0,
			bak:  0,
			ip:   1,
			mode: ModeIdle,
		},
	}.Run(t, 0, 0, 0)

	exeNodeTestCases{
		{
			code: codeLines(
				"JLZ LAB2",
				"LAB1: NOP",
				"LAB2: NOP",
				"LAB3: NOP",
			),
			acc:  42,
			bak:  0,
			ip:   1,
			mode: ModeIdle,
		},
	}.Run(t, 0, 42, 0)

	exeNodeTestCases{
		{
			code: codeLines(
				"JLZ LAB2",
				"LAB1: NOP",
				"LAB2: NOP",
				"LAB3: NOP",
			),
			acc:  -42,
			bak:  0,
			ip:   2,
			mode: ModeIdle,
		},
	}.Run(t, 0, -42, 0)
}
