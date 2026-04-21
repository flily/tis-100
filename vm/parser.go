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

func skipSpace(content []rune, ins *Instruction, start int) int {
	i := start
	for i < len(content) && (isSpace(content[i]) || ins.ExpectBreakpoint(content[i])) {
		if ins.ExpectBreakpoint(content[i]) {
			ins.Breakpoint = true
		}

		i++
	}

	return i
}

func parseOprand(content []rune, ins *Instruction, base *Context, start int, types OprandType) (Oprand, *Context, int, error) {
	i := start
	for i < len(content) {
		c := content[i]
		if isSpace(c) || ins.ExpectBreakpoint(c) {
			break
		}

		i++
	}

	oprandStr := string(content[start:i])
	ctx := base.Mark(start, i)

	if len(oprandStr) <= 0 {
		err := base.MarkAll().Error(errFormatMissingOperand)
		return nil, nil, i, err
	}

	if types.Include(OprandRegister) {
		reg := NewRegister(oprandStr)
		if reg != InvalidRegister {
			return reg, ctx, i, nil
		}
	}

	if types.Include(OprandLiteral) {
		value, err := ParseLiteral(oprandStr)
		if err != nil {
			ctx := base.Mark(start, i)
			return nil, nil, i, ctx.Error(errFormatInvalidExpression, oprandStr)
		}

		return value, ctx, i, nil

	} else {
		// if types.Include(OprandLabel)
		label := NewLabel(oprandStr)
		return label, ctx, i, nil
	}
}

func parseOpcode(content []rune, ins *Instruction, base *Context, start int) error {
	i := start
	leadingSpace := false
	for i < len(content) {
		c := content[i]
		if c == CharLabel {
			labelString := string(content[start:i])
			labelCtx := base.Mark(start, i)
			ins.SetLabel(labelString, labelCtx)
			start = i + 1
			leadingSpace = true

		} else if isSpace(c) || ins.ExpectBreakpoint(c) {
			if ins.ExpectBreakpoint(c) {
				ins.Breakpoint = true
			}

			if leadingSpace {
				start = i + 1

			} else {
				break
			}

		} else {
			leadingSpace = false
		}

		i++
	}

	opcodeStr := string(content[start:i])
	if len(opcodeStr) <= 0 {
		// This is a label line, not an instruction line
		return nil
	}

	opcode := NewOpcode(opcodeStr)
	if opcode == InvalidOpCode {
		ctx := base.Mark(start, i)
		return ctx.Error(errFormatInvalidOpcode, opcodeStr)
	}

	opcodeCtx := base.Mark(start, i)
	ins.SetOpcode(opcode, opcodeCtx)

	if i < len(content) && content[i] == CharBreakpoint {
		// NOP!
		//    ^
		// current bang is marked as breakpoint, but not skipped, next process will read it as part of operand
		i++
	}
	i = skipSpace(content, ins, i)

	oprandTypes := opcode.AcceptOprands()
	oprands := make([]Oprand, 0, len(oprandTypes))
	oprandCtxs := make([]*Context, 0, len(oprandTypes))
	oprandErrs := make([]error, 0, len(oprandTypes))
	for j := range oprandTypes {
		op, ctx, next, err := parseOprand(content, ins, base, i, oprandTypes[j])
		oprands = append(oprands, op)
		oprandCtxs = append(oprandCtxs, ctx)
		oprandErrs = append(oprandErrs, err)
		i = skipSpace(content, ins, next)
	}

	if i < len(content) {
		ctx := base.Mark(i, len(content))
		return ctx.Error(errFormatTooManyOperands)
	}

	for k, oprandErr := range oprandErrs {
		if oprandErr != nil {
			return oprandErr
		}

		ins.AddOprand(oprands[k], oprandCtxs[k])
	}

	return nil
}

func ParseInstruction(line []rune) (Instruction, error) {
	base := NewContext(line)

	ins := &Instruction{
		Opcode: OpEmpty,
	}

	i := skipSpace(line, ins, 0)
	err := parseOpcode(line, ins, base, i)
	if err != nil {
		return InvalidInstruction, err
	}

	return *ins, nil
}
