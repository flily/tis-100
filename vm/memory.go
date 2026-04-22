package vm

type Memory struct {
	Units []Value
	Index int
	Start int
	End   int
}

func NewMemory(size int) *Memory {
	units := make([]Value, size)
	m := &Memory{
		Units: units,
		Index: 0,
		Start: 0,
		End:   size,
	}

	return m
}

func NewMemoryWith(values []Value) *Memory {
	units := make([]Value, len(values))
	copy(units, values)
	m := &Memory{
		Units: units,
		Index: 0,
		Start: 0,
		End:   len(values),
	}

	return m
}

func NewStackMemory(size int) *Memory {
	units := make([]Value, size)
	m := &Memory{
		Units: units,
		Index: 0,
		Start: 0,
		End:   0,
	}

	return m
}

func NewStackMemoryWith(values []Value) *Memory {
	units := make([]Value, len(values))
	copy(units, values)
	m := &Memory{
		Units: units,
		Start: 0,
		Index: len(values),
		End:   len(values),
	}

	return m
}

func (m *Memory) Push(value Value) bool {
	if m.Index >= len(m.Units) || m.End > len(m.Units) {
		return false
	}

	m.Units[m.Index] = value
	m.Index++
	m.End++

	return true
}

func (m *Memory) Pop() (Value, bool) {
	if m.Index <= 0 || m.End <= 0 {
		return 0, false
	}

	m.Index--
	m.End--
	return m.Units[m.Index], true
}

func (m *Memory) Read() (Value, bool) {
	if m.Index < len(m.Units) && m.Index < m.End {
		result := m.Units[m.Index]
		m.Index++
		return result, true
	}

	return 0, false
}

func (m *Memory) Write(value Value) bool {
	if m.Index < len(m.Units) && m.Index < m.End {
		m.Units[m.Index] = value
		m.Index++
		return true
	}

	return false
}
