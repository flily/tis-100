package vm

import (
	"strings"
)

func init() {
	for code, name := range opCodeNames {
		if !strings.HasPrefix(name, "#") {
			opCodeValues[name] = code
		}
	}

	for reg, name := range registerNames {
		if !strings.HasPrefix(name, "#") {
			registerValues[name] = reg
		}
	}
}
