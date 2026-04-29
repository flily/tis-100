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
		node.IP = ip

		if c.ports != nil {
			node.LoadPorts(c.ports)
		}

		node.Step()

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

		if c.portsValue != nil {
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

	if node.Type() != NodeTypeT21 {
		t.Errorf("LoadCode failure: expect NodeTypeT21, got %v", node.Type())
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

	moveForward, looped := node.Step()
	if !moveForward {
		t.Errorf("Step failed: %v", moveForward)
	}

	if looped {
		t.Errorf("Step should not loop when there are more instructions")
	}

	if node.IP != 1 {
		t.Errorf("Step failure: expect IP 1, got %d", node.IP)
	}

	moveForward, looped = node.Step()
	if !moveForward {
		t.Errorf("Step failed: %v", moveForward)
	}

	if looped {
		t.Errorf("Step should not loop when there are more instructions")
	}

	if node.IP != 2 {
		t.Errorf("Step failure: expect IP 2, got %d", node.IP)
	}

	moveForward, looped = node.Step()
	if !moveForward {
		t.Errorf("Step failed: %v", moveForward)
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
		{
			code:  "ADD UP",
			acc:   defaultAcc + 3,
			bak:   0,
			mode:  ModeIdle,
			ports: portsWithData(3, 5, 7, 11),
		},
		{
			code:  "ADD LEFT",
			acc:   defaultAcc + 5,
			bak:   0,
			mode:  ModeIdle,
			ports: portsWithData(3, 5, 7, 11),
		},
		{
			code:  "ADD RIGHT",
			acc:   defaultAcc + 7,
			bak:   0,
			mode:  ModeIdle,
			ports: portsWithData(3, 5, 7, 11),
		},
		{
			code:  "ADD DOWN",
			acc:   defaultAcc + 11,
			bak:   0,
			mode:  ModeIdle,
			ports: portsWithData(3, 5, 7, 11),
		},
		{
			code:  "ADD UP",
			acc:   defaultAcc,
			bak:   0,
			mode:  ModeRead,
			ports: noPortsData(),
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
		{
			code:  "SUB UP",
			acc:   defaultAcc - 3,
			bak:   0,
			mode:  ModeIdle,
			ports: portsWithData(3, 5, 7, 11),
		},
		{
			code:  "SUB LEFT",
			acc:   defaultAcc - 5,
			bak:   0,
			mode:  ModeIdle,
			ports: portsWithData(3, 5, 7, 11),
		},
		{
			code:  "SUB RIGHT",
			acc:   defaultAcc - 7,
			bak:   0,
			mode:  ModeIdle,
			ports: portsWithData(3, 5, 7, 11),
		},
		{
			code:  "SUB DOWN",
			acc:   defaultAcc - 11,
			bak:   0,
			mode:  ModeIdle,
			ports: portsWithData(3, 5, 7, 11),
		},
		{
			code:  "SUB UP",
			acc:   defaultAcc,
			bak:   0,
			mode:  ModeRead,
			ports: noPortsData(),
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

func TestExecutionNodeOpJroMoveForward(t *testing.T) {
	exeNodeTestCases{
		{
			code: codeLines(
				"NOP",
				"NOP",
				"JRO 1",
				"NOP",
				"NOP",
				"NOP",
			),
			acc:  0,
			bak:  0,
			ip:   3,
			mode: ModeIdle,
		},
		{
			code: codeLines(
				"NOP",
				"NOP",
				"JRO 2",
				"NOP",
				"NOP",
				"NOP",
			),
			acc:  0,
			bak:  0,
			ip:   4,
			mode: ModeIdle,
		},
		{
			code: codeLines(
				"NOP",
				"NOP",
				"JRO 3",
				"NOP",
				"NOP",
				"NOP",
			),
			acc:  0,
			bak:  0,
			ip:   5,
			mode: ModeIdle,
		},
		{
			code: codeLines(
				"NOP",
				"NOP",
				"JRO 4",
				"NOP",
				"NOP",
				"NOP",
			),
			acc:  0,
			bak:  0,
			ip:   5,
			mode: ModeIdle,
		},
	}.Run(t, 2, 0, 0)
}

func TestExecutionNodeOpJroMoveBackward(t *testing.T) {
	exeNodeTestCases{
		{
			code: codeLines(
				"NOP",
				"NOP",
				"JRO -1",
				"NOP",
				"NOP",
				"NOP",
			),
			acc:  0,
			bak:  0,
			ip:   1,
			mode: ModeIdle,
		},
		{
			code: codeLines(
				"NOP",
				"NOP",
				"JRO -2",
				"NOP",
				"NOP",
				"NOP",
			),
			acc:  0,
			bak:  0,
			ip:   0,
			mode: ModeIdle,
		},
		{
			code: codeLines(
				"NOP",
				"NOP",
				"JRO -3",
				"NOP",
				"NOP",
				"NOP",
			),
			acc:  0,
			bak:  0,
			ip:   0,
			mode: ModeIdle,
		},
		{
			code: codeLines(
				"NOP",
				"NOP",
				"JRO 0",
				"NOP",
				"NOP",
			),
			acc:  0,
			bak:  0,
			ip:   2,
			mode: ModeIdle,
		},
	}.Run(t, 2, 0, 0)
}

func TestExecutionNodeOpJroRegister(t *testing.T) {
	exeNodeTestCases{
		{
			code: codeLines(
				"NOP",
				"NOP",
				"JRO ACC",
				"NOP",
				"NOP",
				"NOP",
			),
			acc:  2,
			bak:  0,
			ip:   4,
			mode: ModeIdle,
		},
		{
			code: codeLines(
				"NOP",
				"NOP",
				"JRO UP",
				"NOP",
				"NOP",
				"NOP",
			),
			acc:   2,
			bak:   0,
			ip:    5,
			mode:  ModeIdle,
			ports: portsWithData(3, 5, 7, 11),
		},
		{
			code: codeLines(
				"NOP",
				"NOP",
				"JRO UP",
				"NOP",
				"NOP",
				"NOP",
			),
			acc:   2,
			bak:   0,
			ip:    2,
			mode:  ModeRead,
			ports: noPortsData(),
		},
	}.Run(t, 2, 2, 0)
}

func TestLoadCodeError(t *testing.T) {
	node := NewExecutionNode()
	i, err := node.LoadCode("INVALIDOPCODE")
	if err == nil {
		t.Errorf("LoadCode should return error for invalid opcode")
	}
	_ = i
}

func TestLoadLabelsDuplicate(t *testing.T) {
	node := NewExecutionNode()
	code := codeLines(
		"LAB1: NOP",
		"LAB1: NOP",
	)
	i, err := node.LoadCode(code)
	if err == nil {
		t.Errorf("LoadCode should return error for duplicate label")
	}
	if i != 1 {
		t.Errorf("LoadCode should return line 1 for duplicate label, got %d", i)
	}
}

func TestLoadLabelsUndefined(t *testing.T) {
	node := NewExecutionNode()
	code := "JMP UNDEFINED"
	_, err := node.LoadCode(code)
	if err == nil {
		t.Errorf("LoadCode should return error for undefined label")
	}
}

func TestFetchValueFromRegisterDefault(t *testing.T) {
	node := NewExecutionNode()
	v, ok := node.FetchValueFromRegister(Register(999))
	if ok {
		t.Errorf("FetchValueFromRegister with invalid register should return false")
	}
	if v != Value(0) {
		t.Errorf("FetchValueFromRegister with invalid register should return 0, got %d", v)
	}
}

func TestFetchValueFromRegisterNil(t *testing.T) {
	node := NewExecutionNode()
	v, ok := node.FetchValueFromRegister(RegisterNil)
	if !ok {
		t.Errorf("FetchValueFromRegister with RegisterNil should return true")
	}
	if v != Value(0) {
		t.Errorf("FetchValueFromRegister with RegisterNil should return 0, got %d", v)
	}
}

func TestFetchValueFromRegisterLast(t *testing.T) {
	node := NewExecutionNode()
	ports := []IOPort{
		NewValuePortWithValue(99),
		NewValuePortWithValue(0),
		NewValuePortWithValue(0),
		NewValuePortWithValue(0),
	}
	node.LoadPorts(ports)
	// Set Last to RegisterUp by reading any port
	node.ReadAny()
	v, ok := node.FetchValueFromRegister(RegisterLast)
	// After ReadAny, Last should be set to a port; but that port was read so it's now empty
	_ = ok
	_ = v
}

func TestWriteValueDefault(t *testing.T) {
	node := NewExecutionNode()
	ok := node.WriteValue(Register(999), 42)
	if ok {
		t.Errorf("WriteValue with invalid register should return false")
	}
}

func TestWriteValueNil(t *testing.T) {
	node := NewExecutionNode()
	ok := node.WriteValue(RegisterNil, 42)
	if !ok {
		t.Errorf("WriteValue with RegisterNil should return true")
	}
}

func TestPrevIPLooped(t *testing.T) {
	node := NewExecutionNode()
	code := codeLines(
		"NOP",
		"NOP",
		"NOP",
	)
	if _, err := node.LoadCode(code); err != nil {
		t.Fatalf("LoadCode failed: %v", err)
	}

	node.IP = 0
	_, looped := node.prevIP()
	if !looped {
		t.Errorf("prevIP at start should loop around")
	}
}

func TestReadPortDefault(t *testing.T) {
	ports := NewNodePorts()
	ports.Ports = [4]IOPort{
		NewValuePortWithValue(1),
		NewValuePortWithValue(2),
		NewValuePortWithValue(3),
		NewValuePortWithValue(4),
	}

	v, ok := ports.ReadPort(Register(999))
	if ok {
		t.Errorf("ReadPort with invalid register should return false")
	}
	if v != Value(0) {
		t.Errorf("ReadPort with invalid register should return 0, got %d", v)
	}
}

func TestWritePortDefault(t *testing.T) {
	ports := NewNodePorts()
	ports.Ports = [4]IOPort{
		NewValuePort(),
		NewValuePort(),
		NewValuePort(),
		NewValuePort(),
	}

	ok := ports.WritePort(Register(999), 42)
	if ok {
		t.Errorf("WritePort with invalid register should return false")
	}
}

func TestReadAnyEmpty(t *testing.T) {
	ports := NewNodePorts()
	ports.Ports = [4]IOPort{
		NewValuePort(),
		NewValuePort(),
		NewValuePort(),
		NewValuePort(),
	}

	v, ok := ports.ReadAny()
	if ok {
		t.Errorf("ReadAny on empty ports should return false")
	}
	if v != Value(0) {
		t.Errorf("ReadAny on empty ports should return 0, got %d", v)
	}
}

func TestWriteAnyFull(t *testing.T) {
	ports := NewNodePorts()
	ports.Ports = [4]IOPort{
		NewValuePortWithValue(1),
		NewValuePortWithValue(2),
		NewValuePortWithValue(3),
		NewValuePortWithValue(4),
	}
	// Mark all as Busy (not Idle) — write again to ports already in Ready state
	// Actually ValuePortWithValue puts them in IOModeReady, which is not Idle, so Write returns false
	ok := ports.WriteAny(99)
	if ok {
		t.Errorf("WriteAny on full ports should return false")
	}
}

func TestMovNil(t *testing.T) {
	exeNodeTestCases{
		{
			code: "MOV ACC, NIL",
			acc:  13,
			bak:  0,
			mode: ModeIdle,
		},
		{
			code: "MOV NIL, ACC",
			acc:  0,
			bak:  0,
			mode: ModeIdle,
		},
	}.Run(t, 0, 13, 0)
}

func TestFetchValueDefault(t *testing.T) {
	node := NewExecutionNode()
	// Label has OprandType() == OprandLabel which falls to the default case
	v, ok := node.FetchValue(NewLabel("TEST"))
	if ok {
		t.Errorf("FetchValue with Label operand should return false")
	}
	if v != Value(0) {
		t.Errorf("FetchValue with Label operand should return 0, got %d", v)
	}
}

func TestPrevIPOnAllEmptyCode(t *testing.T) {
	// A code with only empty instructions triggers the "Empty code segment" path
	node := &BasicExecutionNode{
		NodePorts: *NewNodePorts(),
		Codes: Code{
			{Opcode: OpEmpty},
			{Opcode: OpEmpty},
		},
		IP: 0,
	}
	node.labels = make(map[Label]int)
	_, _ = node.prevIP()
}
