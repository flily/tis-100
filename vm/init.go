package vm

func init() {
	for code, name := range opCodeNames {
		opCodeValues[name] = code
	}

	for reg, name := range registerNames {
		registerValues[name] = reg
	}
}
