package vm

type IOMode int

const (
	IOModeIdle IOMode = iota
	IOModeBusy
	IOModeReady
)

type IOPort interface {
	Value() Value
	Mode() IOMode
	Read() (Value, bool)
	Write(value Value) bool
	WriteDone()
}

type BlockPort int

const (
	InvalidPort BlockPort = 0
)

func (p BlockPort) Value() Value {
	return Value(0)
}

func (p BlockPort) Mode() IOMode {
	return IOModeBusy
}

func (p BlockPort) Read() (Value, bool) {
	return Value(0), false
}

func (p BlockPort) Write(_ Value) bool {
	return false
}

func (p BlockPort) WriteDone() {
}

type ConstPort int

const (
	NilPort ConstPort = 0
)

func (p ConstPort) Value() Value {
	return Value(p)
}

func (p ConstPort) Mode() IOMode {
	return IOModeReady
}

func (p ConstPort) Read() (Value, bool) {
	return Value(p), true
}

func (p ConstPort) Write(_ Value) bool {
	return true
}

func (p ConstPort) WriteDone() {
}

type ValuePort struct {
	value Value
	state IOMode
}

func NewValuePort() *ValuePort {
	p := &ValuePort{
		value: 0,
		state: IOModeIdle,
	}

	return p
}

func NewValuePortWithValue(v Value) *ValuePort {
	p := &ValuePort{
		value: v,
		state: IOModeReady,
	}

	return p
}

func (p *ValuePort) Value() Value {
	return p.value
}

func (p *ValuePort) Mode() IOMode {
	return p.state
}

func (p *ValuePort) Read() (Value, bool) {
	if p.state == IOModeReady {
		p.state = IOModeIdle
		return p.value, true
	}

	return Value(0), false
}

func (p *ValuePort) Write(v Value) bool {
	if p.state == IOModeIdle {
		p.value = v
		p.state = IOModeBusy
		return true
	}

	return false
}

func (p *ValuePort) WriteDone() {
	if p.state == IOModeBusy {
		p.state = IOModeReady
	}
}

type IOPortEnd struct {
	in  IOPort
	out IOPort
}

func NewIOPortEnd(in IOPort, out IOPort) *IOPortEnd {
	e := &IOPortEnd{
		in:  in,
		out: out,
	}

	return e
}

func (e *IOPortEnd) Value() Value {
	return e.out.Value()
}

func (e *IOPortEnd) Mode() IOMode {
	return e.out.Mode()
}

func (e *IOPortEnd) Read() (Value, bool) {
	return e.out.Read()
}

func (e *IOPortEnd) Write(v Value) bool {
	return e.in.Write(v)
}

func (e *IOPortEnd) WriteDone() {
	e.in.WriteDone()
}

type IOPipe struct {
	p1 IOPort
	p2 IOPort
}

func NewOneWayPipe() *IOPipe {
	p := &IOPipe{
		p1: NewValuePort(),
		p2: NilPort,
	}

	return p
}

func NewRoundWayPipe() *IOPipe {
	p := &IOPipe{
		p1: NewValuePort(),
		p2: NewValuePort(),
	}

	return p
}

func (p *IOPipe) Ports() (IOPort, IOPort) {
	e1 := NewIOPortEnd(p.p1, p.p2)
	e2 := NewIOPortEnd(p.p2, p.p1)

	// write port is e1, read port is e2
	return e1, e2
}
