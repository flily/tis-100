package vm

// Label support -`~$%^&*()+=_./?'"
// Comma in label is acepted, but this label can not be used as operand, because it will be split into two operands

const (
	CharBreakpoint = '!'
	CharLabel      = ':'

	errFormatInvalidOpcode     = `INVALID OPCODE "%s"`
	errFormatInvalidExpression = `INVALID EXPRESSION "%s"`
	errFormatMissingOperand    = `MISSING OPERAND`
	errFormatTooManyOperands   = `TOO MANY OPERANDS`
	errFormatUndefinedLabel    = `UNDEFINED LABEL`
	errFormatDuplicateLabel    = `LABEL IS ALREADY DEFINED`
)

type SyntaxError struct {
	ctx     Context
	message string
}

func (e *SyntaxError) Line() int {
	return e.ctx.Line
}

func (e *SyntaxError) Position() (int, int) {
	return e.ctx.Start, e.ctx.End
}

func (e *SyntaxError) Error() string {
	return e.ctx.Message(e.message)
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == ','
}

func skipSpace(content []rune, start int) int {
	i := start
	for i < len(content) && isSpace(content[i]) {
		i++
	}

	return i
}

func parseOprand(content []rune, _ *Instruction, base *Context, start int, types OprandType) (Oprand, int, error) {
	i := start
	for i < len(content) {
		c := content[i]
		if isSpace(c) {
			break
		}

		i++
	}

	oprandStr := string(content[start:i])
	if len(oprandStr) <= 0 {
		return nil, i, nil
	}

	if types.Include(OprandRegister) {
		reg := NewRegister(oprandStr)
		if reg != InvalidRegister {
			return reg, i, nil
		}
	}

	if types.Include(OprandLiteral) {
		value, err := ParseLiteral(oprandStr)
		if err != nil {
			ctx := base.Mark(start, i)
			return nil, -1, ctx.Error(errFormatInvalidExpression, oprandStr)
		}

		return value, i, nil
	}

	if types.Include(OprandLabel) {
		label := NewLabel(oprandStr)
		return label, i, nil
	}

	return nil, i, nil
}

func parseOpcode(content []rune, ins *Instruction, base *Context, start int) (int, error) {
	i := start
	for i < len(content) {
		c := content[i]
		if c == CharLabel {
			labelString := string(content[start:i])
			ins.Label = labelString
			start = i + 1

		} else if isSpace(c) {
			break
		}

		i++
	}

	opcodeStr := string(content[start:i])
	if len(opcodeStr) <= 0 && len(ins.Label) > 0 {
		// This is a label line, not an instruction line
		return i, nil
	}

	opcode := NewOpcode(opcodeStr)
	if opcode == InvalidOpCode {
		ctx := base.Mark(start, i)
		return -1, ctx.Error(errFormatInvalidOpcode, opcodeStr)
	}

	ins.Opcode = opcode
	i = skipSpace(content, i)

	oprandTypes := opcode.AcceptOprands()
	if len(oprandTypes) <= 0 {
		return i, nil
	}

	if len(oprandTypes) > 0 {
		op1, next, err := parseOprand(content, ins, base, i, oprandTypes[0])
		if err != nil {
			return -1, err
		}

		ins.Oprands1 = op1
		i = skipSpace(content, next)

		if len(oprandTypes) > 1 {
			op2, next, err := parseOprand(content, ins, base, i, oprandTypes[1])
			if err != nil {
				return -1, err
			}

			ins.Oprands2 = op2
			i = skipSpace(content, next)
		}
	}

	return i, nil
}

func ParseInstruction(line []rune) (Instruction, error) {
	base := NewContext(line)

	ins := &Instruction{
		Opcode: OpEmpty,
	}

	i := 0
	for i < len(line) {
		c := line[i]
		if !ins.Breakpoint && c == CharBreakpoint {
			ins.Breakpoint = true
			i++

		} else if isSpace(c) {
			i++

		} else {
			next, err := parseOpcode(line, ins, base, i)
			if err != nil {
				return InvalidInstruction, err
			}

			i = next
		}
	}

	return *ins, nil
}
