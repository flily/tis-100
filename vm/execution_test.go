package vm

import (
	"testing"

	"slices"
	"strings"
)

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

func TestExecutionNodeOpMovBasicIO(t *testing.T) {
	node := NewExecutionNode()
	defaultAcc := Value(13)

	cases := []struct {
		ins        string
		acc        Value
		bak        Value
		mode       ExecutionMode
		ports      []IOPort
		portsValue []Value
	}{
		{
			ins:  "MOV 42, ACC",
			acc:  42,
			bak:  0,
			mode: ModeIdle,
		},
		{
			ins:  "MOV ACC, UP",
			acc:  13,
			bak:  0,
			mode: ModeIdle,
			ports: []IOPort{
				NewValuePort(),
				NewValuePort(),
				NewValuePort(),
				NewValuePort(),
			},
			portsValue: []Value{13, 0, 0, 0},
		},
		{
			ins:  "MOV ACC, LEFT",
			acc:  13,
			bak:  0,
			mode: ModeIdle,
			ports: []IOPort{
				NewValuePort(),
				NewValuePort(),
				NewValuePort(),
				NewValuePort(),
			},
			portsValue: []Value{0, 13, 0, 0},
		},
		{
			ins:  "MOV ACC, RIGHT",
			acc:  13,
			bak:  0,
			mode: ModeIdle,
			ports: []IOPort{
				NewValuePort(),
				NewValuePort(),
				NewValuePort(),
				NewValuePort(),
			},
			portsValue: []Value{0, 0, 13, 0},
		},
		{
			ins:  "MOV ACC, DOWN",
			acc:  13,
			bak:  0,
			mode: ModeIdle,
			ports: []IOPort{
				NewValuePort(),
				NewValuePort(),
				NewValuePort(),
				NewValuePort(),
			},
			portsValue: []Value{0, 0, 0, 13},
		},
		{
			ins:  "MOV UP, ACC",
			acc:  3,
			bak:  0,
			mode: ModeIdle,
			ports: []IOPort{
				NewValuePortWithValue(3),
				NewValuePortWithValue(5),
				NewValuePortWithValue(7),
				NewValuePortWithValue(11),
			},
			portsValue: []Value{3, 5, 7, 11},
		},
		{
			ins:  "MOV LEFT, ACC",
			acc:  5,
			bak:  0,
			mode: ModeIdle,
			ports: []IOPort{
				NewValuePortWithValue(3),
				NewValuePortWithValue(5),
				NewValuePortWithValue(7),
				NewValuePortWithValue(11),
			},
			portsValue: []Value{3, 5, 7, 11},
		},
		{
			ins:  "MOV RIGHT, ACC",
			acc:  7,
			bak:  0,
			mode: ModeIdle,
			ports: []IOPort{
				NewValuePortWithValue(3),
				NewValuePortWithValue(5),
				NewValuePortWithValue(7),
				NewValuePortWithValue(11),
			},
			portsValue: []Value{3, 5, 7, 11},
		},
		{
			ins:  "MOV DOWN, ACC",
			acc:  11,
			bak:  0,
			mode: ModeIdle,
			ports: []IOPort{
				NewValuePortWithValue(3),
				NewValuePortWithValue(5),
				NewValuePortWithValue(7),
				NewValuePortWithValue(11),
			},
			portsValue: []Value{3, 5, 7, 11},
		},
	}

	for _, c := range cases {
		if i, err := node.LoadCode(c.ins); err != nil {
			t.Fatalf("LoadCode failed at instruction %d: %v", i, err)
		}

		node.Acc = defaultAcc

		if c.ports != nil {
			node.LoadPorts(c.ports)
		}

		err := node.Step()
		if err != nil {
			t.Fatalf("Step failed: %v", err)
		}

		if node.Acc != c.acc {
			t.Errorf("MOV instruction failed: expect Acc %d, got %d", c.acc, node.Acc)
		}

		if node.Backup != c.bak {
			t.Errorf("MOV instruction failed: expect Backup %d, got %d", c.bak, node.Backup)
		}

		if node.Mode != c.mode {
			t.Errorf("MOV instruction failed: expect Mode %v, got %v", c.mode, node.Mode)
		}

		if c.ports != nil {
			got := node.Snapshot()
			if !slices.Equal(got, c.portsValue) {
				t.Errorf("MOV instruction failed: ports not as expected")
				t.Errorf("expect: %v", c.portsValue)
				t.Errorf("got: %v", got)
			}
		}
	}
}

func TestExecutionNodeOpMovBlocking(t *testing.T) {
	node := NewExecutionNode()
	defaultAcc := Value(13)

	cases := []struct {
		ins        string
		acc        Value
		bak        Value
		mode       ExecutionMode
		ports      []IOPort
		portsValue []Value
	}{
		{
			ins:  "MOV UP, ACC",
			acc:  13,
			bak:  0,
			mode: ModeRead,
			ports: []IOPort{
				NewValuePort(),
				NewValuePort(),
				NewValuePort(),
				NewValuePort(),
			},
			portsValue: []Value{0, 0, 0, 0},
		},
		{
			ins:  "MOV LEFT, ACC",
			acc:  13,
			bak:  0,
			mode: ModeRead,
			ports: []IOPort{
				NewValuePort(),
				NewValuePort(),
				NewValuePort(),
				NewValuePort(),
			},
			portsValue: []Value{0, 0, 0, 0},
		},
		{
			ins:  "MOV RIGHT, ACC",
			acc:  13,
			bak:  0,
			mode: ModeRead,
			ports: []IOPort{
				NewValuePort(),
				NewValuePort(),
				NewValuePort(),
				NewValuePort(),
			},
			portsValue: []Value{0, 0, 0, 0},
		},
		{
			ins:  "MOV DOWN, ACC",
			acc:  13,
			bak:  0,
			mode: ModeRead,
			ports: []IOPort{
				NewValuePort(),
				NewValuePort(),
				NewValuePort(),
				NewValuePort(),
			},
			portsValue: []Value{0, 0, 0, 0},
		},
		{
			ins:  "MOV ACC, UP",
			acc:  13,
			bak:  0,
			mode: ModeWrite,
			ports: []IOPort{
				NewValuePortWithValue(3),
				NewValuePortWithValue(5),
				NewValuePortWithValue(7),
				NewValuePortWithValue(11),
			},
			portsValue: []Value{3, 5, 7, 11},
		},
		{
			ins:  "MOV ACC, LEFT",
			acc:  13,
			bak:  0,
			mode: ModeWrite,
			ports: []IOPort{
				NewValuePortWithValue(3),
				NewValuePortWithValue(5),
				NewValuePortWithValue(7),
				NewValuePortWithValue(11),
			},
			portsValue: []Value{3, 5, 7, 11},
		},
		{
			ins:  "MOV ACC, RIGHT",
			acc:  13,
			bak:  0,
			mode: ModeWrite,
			ports: []IOPort{
				NewValuePortWithValue(3),
				NewValuePortWithValue(5),
				NewValuePortWithValue(7),
				NewValuePortWithValue(11),
			},
			portsValue: []Value{3, 5, 7, 11},
		},
		{
			ins:  "MOV ACC, DOWN",
			acc:  13,
			bak:  0,
			mode: ModeWrite,
			ports: []IOPort{
				NewValuePortWithValue(3),
				NewValuePortWithValue(5),
				NewValuePortWithValue(7),
				NewValuePortWithValue(11),
			},
			portsValue: []Value{3, 5, 7, 11},
		},
	}

	for _, c := range cases {
		if i, err := node.LoadCode(c.ins); err != nil {
			t.Fatalf("LoadCode failed at instruction %d: %v", i, err)
		}

		node.Acc = defaultAcc

		if c.ports != nil {
			node.LoadPorts(c.ports)
		}

		err := node.Step()
		if err != nil {
			t.Fatalf("Step failed: %v", err)
		}

		if node.Acc != c.acc {
			t.Errorf("MOV instruction failed: expect Acc %d, got %d", c.acc, node.Acc)
		}

		if node.Backup != c.bak {
			t.Errorf("MOV instruction failed: expect Backup %d, got %d", c.bak, node.Backup)
		}

		if node.Mode != c.mode {
			t.Errorf("MOV instruction failed: expect Mode %v, got %v", c.mode, node.Mode)
		}

		if c.ports != nil {
			got := node.Snapshot()
			if !slices.Equal(got, c.portsValue) {
				t.Errorf("MOV instruction failed: ports not as expected")
				t.Errorf("expect: %v", c.portsValue)
				t.Errorf("got: %v", got)
			}
		}
	}
}
