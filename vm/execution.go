package vm

type NodePorts struct {
	Ports [4]IOPort // 0: UP, 1: LEFT, 2: RIGHT, 3: DOWN
	Last  Register
}

func NewNodePorts() *NodePorts {
	p := &NodePorts{
		Ports: [4]IOPort{},
		Last:  RegisterNA,
	}

	return p
}

var (
	portMap = map[int]Register{
		0: RegisterUp,
		1: RegisterLeft,
		2: RegisterRight,
		3: RegisterDown,
	}
	readAnyOrder  = []int{1, 2, 0, 3}
	writeAnyOrder = []int{0, 1, 2, 3}
)

func (p *NodePorts) ReadPort(port Register) (Value, bool) {
	switch port {
	case RegisterUp:
		return p.Ports[0].Read()
	case RegisterLeft:
		return p.Ports[1].Read()
	case RegisterRight:
		return p.Ports[2].Read()
	case RegisterDown:
		return p.Ports[3].Read()
	default:
		return Value(0), false
	}
}

func (p *NodePorts) ReadAny() (Value, bool) {
	for _, i := range readAnyOrder {
		v, ok := p.Ports[i].Read()
		if ok {
			p.Last = portMap[i]
			return v, true
		}
	}

	return Value(0), false
}

func (p *NodePorts) WritePort(port Register, v Value) bool {
	switch port {
	case RegisterUp:
		return p.Ports[0].Write(v)
	case RegisterLeft:
		return p.Ports[1].Write(v)
	case RegisterRight:
		return p.Ports[2].Write(v)
	case RegisterDown:
		return p.Ports[3].Write(v)
	default:
		return false
	}
}

func (p *NodePorts) WriteAny(v Value) bool {
	for _, i := range writeAnyOrder {
		ok := p.Ports[i].Write(v)
		if ok {
			p.Last = portMap[i]
			return true
		}
	}

	return false
}

func (p *NodePorts) WriteDone() {
	for _, port := range p.Ports {
		port.WriteDone()
	}
}

func (p *NodePorts) LoadPorts(ports []IOPort) {
	copy(p.Ports[:], ports)
}

func (p *NodePorts) Snapshot() []Value {
	snapshot := make([]Value, len(p.Ports))
	for i, port := range p.Ports {
		v := port.Value()
		snapshot[i] = v
	}

	return snapshot
}

// Type T21
type BasicExecutionNode struct {
	NodePorts
	Codes  Code
	IP     int
	PC     int
	Acc    Value
	Backup Value
	Mode   ExecutionMode
	labels map[Label]int
}

func NewExecutionNode() *BasicExecutionNode {
	n := &BasicExecutionNode{
		NodePorts: *NewNodePorts(),
		Codes:     nil,
		IP:        0,
		PC:        0,
		Acc:       0,
		Mode:      ModeIdle,
	}

	return n
}

func (n *BasicExecutionNode) Type() NodeType {
	return NodeTypeT21
}

func (n *BasicExecutionNode) checkLabels() (int, error) {
	for i, ins := range n.Codes {
		if ins.Opcode < OpJMP || ins.Opcode > OpJLZ {
			continue
		}

		label := ins.Oprand1.(Label)
		if _, found := n.labels[label]; !found {
			return i, ins.Oprand1Ctx.Error(errFormatUndefinedLabel)
		}
	}

	return -1, nil
}

func (n *BasicExecutionNode) LoadLabels() (int, error) {
	n.labels = make(map[Label]int)

	for i, ins := range n.Codes {
		label := ins.Label
		if len(label) <= 0 {
			continue
		}

		_, found := n.labels[label]
		if found {
			return i, ins.LabelCtx.Error(errFormatDuplicateLabel)
		}

		n.labels[label] = len(n.labels)
	}

	return n.checkLabels()
}

func (n *BasicExecutionNode) Load(instructions Code) (int, error) {
	n.Codes = instructions
	n.IP = 0
	n.PC = 0
	n.Acc = 0
	n.Mode = ModeIdle

	if i, err := n.LoadLabels(); err != nil {
		return i, err
	}

	return n.checkLabels()
}

func (n *BasicExecutionNode) LoadCode(code string) (int, error) {
	instructions, i, err := ParseCode(code)
	if err != nil {
		return i, err
	}

	return n.Load(instructions)
}

func (n *BasicExecutionNode) FetchValueFromRegister(r Register) (Value, bool) {
	switch r {
	case RegisterAcc:
		return n.Acc, true

	case RegisterUp, RegisterDown, RegisterLeft, RegisterRight:
		return n.ReadPort(r)

	case RegisterNil:
		return NilPort.Read()

	case RegisterLast:
		return n.ReadPort(n.Last)

	default:
		return Value(0), false
	}
}

func (n *BasicExecutionNode) FetchValue(o Oprand) (Value, bool) {
	switch o.OprandType() {
	case OprandValue:
		v := o.(Value)
		return v, true

	case OprandRegister:
		reg := o.(Register)
		return n.FetchValueFromRegister(reg)

	default:
		return Value(0), false
	}
}

func (n *BasicExecutionNode) WriteValue(reg Register, v Value) bool {
	switch reg {
	case RegisterAcc:
		n.Acc = v
		return true

	case RegisterUp, RegisterDown, RegisterLeft, RegisterRight:
		return n.WritePort(reg, v)

	case RegisterNil:
		return NilPort.Write(v)

	default:
		return false
	}
}

func (n *BasicExecutionNode) nextIP() Instruction {
	for range len(n.Codes) {
		n.IP = (n.IP + 1) % len(n.Codes)
		inst := n.Codes[n.IP]
		if inst.Opcode != OpEmpty {
			return inst
		}
	}

	// Empty code segment.
	return InvalidInstruction
}

func (n *BasicExecutionNode) Step() error {
	inst := n.Codes[n.IP]
	n.nextIP()
	return n.RunInst(inst)
}

func (n *BasicExecutionNode) RunInst(inst Instruction) error {
	switch inst.Opcode {
	case OpNOP:
		// Do nothing

	case OpMOV:
		o1, ok := n.FetchValue(inst.Oprand1)
		if !ok {
			n.Mode = ModeRead
			break
		}

		ok = n.WriteValue(inst.Oprand2.(Register), o1)
		if !ok {
			n.Mode = ModeWrite
			break
		}

	case OpSWP:
		n.Acc, n.Backup = n.Backup, n.Acc

	case OpSAV:
		n.Backup = n.Acc

	case OpADD:
		o1, ok := n.FetchValue(inst.Oprand1)
		if !ok {
			n.Mode = ModeRead
			break
		}

		n.Acc = (n.Acc + o1).Limit()

	case OpSUB:
		o1, ok := n.FetchValue(inst.Oprand1)
		if !ok {
			n.Mode = ModeRead
			break
		}

		n.Acc = (n.Acc - o1).Limit()

	case OpNEG:
		n.Acc = -n.Acc

	}

	n.IP += 1
	return nil
}
