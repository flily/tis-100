package vm

import (
	"testing"
)

func TestExecutionModeString(t *testing.T) {
	cases := []struct {
		mode   ExecutionMode
		expect string
	}{
		{ModeIdle, "IDLE"},
		{ModeWrite, "WRITE"},
		{ModeRead, "READ"},
		{ExecutionMode(-1), "UNKNOWN"},
	}

	for _, c := range cases {
		if c.mode.String() != c.expect {
			t.Errorf("ExecutionMode %d: expect String() to return %s, got %s",
				c.mode, c.expect, c.mode.String())
		}
	}
}
