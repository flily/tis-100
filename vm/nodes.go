package vm

type (
	NodeType      int
	ExecutionMode int
)

const (
	NodeTypeT20 NodeType = 20 // T20, restricted to special models of TIS, and not documented.
	NodeTypeT21 NodeType = 21 // T21, basic execution node
	NodeTypeT30 NodeType = 30 // T30, stack memory node
	NodeTypeT31 NodeType = 31 // T31, random access memory node, not yet available in TIS-100.

	ModeIdle ExecutionMode = iota
	ModeWrite
	ModeRead
)

var executionModeNames = map[ExecutionMode]string{
	ModeIdle:  "IDLE",
	ModeWrite: "WRITE",
	ModeRead:  "READ",
}

func (m ExecutionMode) String() string {
	if name, ok := executionModeNames[m]; ok {
		return name
	}

	return "UNKNOWN"
}

type Node interface {
	Type() NodeType
	Step() (bool, error)
	IsHalt() bool
}
