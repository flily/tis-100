package vm

import (
	"testing"
)

func TestBlockPort(t *testing.T) {
	for i := range Value(100) {
		ok := InvalidPort.Write(i)
		if ok {
			t.Errorf("InvalidPort should always reject write")
		}

		v, ok := InvalidPort.Read()
		if ok {
			t.Errorf("InvalidPort should always return false for read")
		}

		if v != Value(0) {
			t.Errorf("InvalidPort should always return 0 for read, got %d", v)
		}
	}
}

func TestNilPort(t *testing.T) {
	for i := range Value(100) {
		ok := NilPort.Write(i)
		if !ok {
			t.Errorf("NilPort should always accept write")
		}

		v, ok := NilPort.Read()
		if !ok {
			t.Errorf("NilPort should always return true for read")
		}

		if v != Value(0) {
			t.Errorf("NilPort should always return 0 for read, got %d", v)
		}
	}
}

func TestValuePort(t *testing.T) {
	p := NewValuePort()

	// Read empty port
	v, ok := p.Read()
	if ok {
		t.Errorf("ValuePort should return false for read when idle")
	}

	if v != Value(0) {
		t.Errorf("ValuePort should return 0 for read when idle, got %d", v)
	}

	// Write to port
	ok = p.Write(42)
	if !ok {
		t.Errorf("ValuePort should accept write when idle")
	}

	if p.Value() != 42 {
		t.Errorf("ValuePort should store the written value, expect 42, got %d", p.Value())
	}

	// write to busy port
	ok = p.Write(100)
	if ok {
		t.Errorf("ValuePort should reject write when busy")
	}

	// Read from busy port
	v, ok = p.Read()
	if ok {
		t.Errorf("ValuePort should return false for read when busy")
	}

	if v != 0 {
		t.Errorf("ValuePort should return the stored value for read, expect 0, got %d", v)
	}

	p.WriteDone()

	// Read from ready port
	v, ok = p.Read()
	if !ok {
		t.Errorf("ValuePort should return true for read when ready")
	}

	if v != 42 {
		t.Errorf("ValuePort should return the stored value for read, expect 42, got %d", v)
	}

	// Read again should be idle
	v, ok = p.Read()
	if ok {
		t.Errorf("ValuePort should return false for read after being read")
	}

	if v != Value(0) {
		t.Errorf("ValuePort should return 0 for read after being read, got %d", v)
	}
}

func TestIOPipe(t *testing.T) {
	pipe := NewRoundWayPipe()

	p1, p2 := pipe.Ports()

	{
		ok := p1.Write(42)
		if !ok {
			t.Errorf("p1 should accept write when idle")
		}

		v, ok := p2.Read()
		if ok {
			t.Errorf("p2 should return false for read when busy")
		}

		if v != 0 {
			t.Errorf("p2 should return 0 for read when busy, got %d", v)
		}

		p1.WriteDone()

		v, ok = p2.Read()
		if !ok {
			t.Errorf("p2 should return true for read when ready")
		}

		if v != 42 {
			t.Errorf("p2 should return the value written to p1, expect 42, got %d", v)
		}
	}

	{
		ok := p2.Write(100)
		if !ok {
			t.Errorf("p2 should accept write when idle")
		}

		v, ok := p1.Read()
		if ok {
			t.Errorf("p1 should return false for read when busy")
		}

		if v != 0 {
			t.Errorf("p1 should return 0 for read when busy, got %d", v)
		}

		p2.WriteDone()

		v, ok = p1.Read()
		if !ok {
			t.Errorf("p1 should return true for read when ready")
		}

		if v != 100 {
			t.Errorf("p1 should return the value written to p2, expect 100, got %d", v)
		}
	}
}
