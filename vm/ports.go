package vm

type IOMode int

const (
	IOModeIdle IOMode = iota
	IOModeReady
)

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
		p.State = IOModeReady
		return true
	}

	return false
}

type IOPortEnd struct {
	in  *ValuePort
	out *ValuePort
}

func NewIOPortEnd(in *ValuePort, out *ValuePort) *IOPortEnd {
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

type IOPipe struct {
	p1 *ValuePort
	p2 *ValuePort
}

func NewIOPipe() *IOPipe {
	p := &IOPipe{
		p1: NewValuePort(),
		p2: NewValuePort(),
	}

	return p
}

func (p *IOPipe) Ends() (IOPort, IOPort) {
	e1 := NewIOPortEnd(p.p1, p.p2)
	e2 := NewIOPortEnd(p.p2, p.p1)

	return e1, e2
}
