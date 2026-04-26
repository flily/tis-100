package vm

type IOMode int

type IOPort interface {
	Read() (Value, bool)
	Write(value Value) bool
	WriteDone()
}

const (
	IOModeIdle IOMode = iota
	IOModeBusy
	IOModeReady
)

type BlockPort int

const (
	InvalidPort BlockPort = 0
)

func (p BlockPort) Read() (Value, bool) {
	return Value(0), false
}

func (p BlockPort) Write(_ Value) bool {
	return false
}

type ConstPort int

const (
	NilPort ConstPort = 0
)

func (p ConstPort) Read() (Value, bool) {
	return Value(p), true
}

func (p ConstPort) Write(_ Value) bool {
	return true
}

func (p ConstPort) WriteDone() {
}

type ValuePort struct {
	Value Value
	State IOMode
}

func NewValuePort() *ValuePort {
	p := &ValuePort{
		Value: 0,
		State: IOModeIdle,
	}

	return p
}

func NewValuePortWithValue(v Value) *ValuePort {
	p := &ValuePort{
		Value: v,
		State: IOModeReady,
	}

	return p
}

func (p *ValuePort) Read() (Value, bool) {
	if p.State == IOModeReady {
		p.State = IOModeIdle
		return p.Value, true
	}

	return Value(0), false
}

func (p *ValuePort) Write(v Value) bool {
	if p.State == IOModeIdle {
		p.Value = v
		p.State = IOModeBusy
		return true
	}

	return false
}

func (p *ValuePort) WriteDone() {
	if p.State == IOModeBusy {
		p.State = IOModeReady
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
