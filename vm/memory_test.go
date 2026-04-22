package vm

import (
	"testing"

	"slices"
)

func TestMemoryAsStaticMemoryQueueRead(t *testing.T) {
	data := []Value{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}
	m := NewMemoryWith(data)

	if m.Length() != len(data) || m.Size() != len(data) {
		t.Errorf("Got wrong length or size: length=%d size=%d, expect length=%d size=%d", m.Length(), m.Size(), len(data), len(data))
	}

	got := make([]Value, 0, len(data))
	for {
		value, ok := m.Read()
		if !ok {
			break
		}

		got = append(got, value)
	}

	if !m.EqualToValues(data) {
		t.Errorf("Read wrong result")
		t.Errorf("expect: %v", data)
		t.Errorf("got: %v", got)
	}

	if m.Index != len(data) {
		t.Errorf("Got wrong index, expect: %d, got: %d", len(data), m.Index)
	}
}

func TestMemoryAsStaticMemoryQueueWrite1(t *testing.T) {
	data := []Value{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}
	m := NewMemory(5)

	if m.Length() != 0 || m.Size() != 5 {
		t.Errorf("Got wrong length or size: length=%d size=%d, expect length=%d size=%d", m.Length(), m.Size(), 0, 5)
	}

	for _, value := range data {
		ok := m.Write(value)
		if !ok {
			break
		}
	}

	exp := []Value{2, 3, 5, 7, 11}
	if !m.EqualToValues(exp) {
		t.Errorf("Write wrong result")
		t.Errorf("expect: %v", exp)
		t.Errorf("got: %v", m.Units)
	}

	if m.Index != 5 {
		t.Errorf("Got wrong index, expect: %d, got: %d", 5, m.Index)
	}

	if m.End != 5 {
		t.Errorf("Got wrong end, expect: %d, got: %d", 5, m.End)
	}
}

func TestMemoryAsStaticMemoryQueueWrite2(t *testing.T) {
	data := []Value{2, 3, 5, 7, 11}
	m := NewMemory(8)

	for _, value := range data {
		ok := m.Write(value)
		if !ok {
			break
		}
	}

	exp := []Value{2, 3, 5, 7, 11, 0, 0, 0}
	if !slices.Equal(m.Units, exp) {
		t.Errorf("Write wrong result")
		t.Errorf("expect: %v", exp)
		t.Errorf("got: %v", m.Units)
	}

	if !m.EqualToValues(data) {
		t.Errorf("Write wrong result")
		t.Errorf("expect: %v", data)
		t.Errorf("got: %v", m.Units)
	}

	if m.Index != 5 {
		t.Errorf("Got wrong index, expect: %d, got: %d", 5, m.Index)
	}

	if m.End != 5 {
		t.Errorf("Got wrong end, expect: %d, got: %d", 5, m.End)
	}
}

func TestMemoryAsStackPop(t *testing.T) {
	data := []Value{2, 3, 5, 7, 11}
	m := NewStackMemoryWith(data)

	if m.Index != len(data) || m.End != len(data) {
		t.Errorf("Get wrong index or end: index=%d end=%d, expect index=%d end=%d", m.Index, m.End, len(data), len(data))
	}

	got := make([]Value, 0, len(data))
	for {
		value, ok := m.Pop()
		if !ok {
			break
		}

		got = append(got, value)
	}

	exp := []Value{11, 7, 5, 3, 2}
	if !slices.Equal(got, exp) {
		t.Errorf("Pop wrong result")
		t.Errorf("expect: %v", exp)
		t.Errorf("got: %v", got)
	}

	if m.Index != 0 {
		t.Errorf("Got wrong index, expect: %d, got: %d", 0, m.Index)
	}

	if m.End != 0 {
		t.Errorf("Got wrong end, expect: %d, got: %d", 0, m.End)
	}
}

func TestMemoryAsStackPush(t *testing.T) {
	data := []Value{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}
	m := NewStackMemory(5)

	if m.Index != 0 || m.End != 0 {
		t.Errorf("Get wrong index or end: index=%d end=%d, expect index=%d end=%d", m.Index, m.End, 0, 0)
	}

	for _, value := range data {
		ok := m.Push(value)
		if !ok {
			break
		}
	}

	exp := []Value{2, 3, 5, 7, 11}
	if !m.EqualToValues(exp) {
		t.Errorf("Push wrong result")
		t.Errorf("expect: %v", exp)
		t.Errorf("got: %v", m.Units)
	}

	if m.Index != 5 {
		t.Errorf("Got wrong index, expect: %d, got: %d", 5, m.Index)
	}

	if m.End != 5 {
		t.Errorf("Got wrong end, expect: %d, got: %d", 5, m.End)
	}
}
