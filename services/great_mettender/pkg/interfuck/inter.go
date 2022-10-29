package interfuck

import (
	"bytes"
	"errors"
	"fmt"
)

var (
	ErrOpsLimit = errors.New("ops limit reached")
	ErrCE       = errors.New("compilation error")
	ErrOOB      = errors.New("pointer out of bounds")
)

type State struct {
	mem [30000]uint16
	ptr int
	pc  int

	stdin  bytes.Buffer
	stdout bytes.Buffer
}

type Program struct {
	state *State

	ops      []op
	opsCnt   int
	opsLimit int
}

func (p *Program) Run(input []byte) (output string, ops int, err error) {
	p.state.stdin.Reset()
	p.state.stdout.Reset()
	if _, err := p.state.stdin.Write(input); err != nil {
		return "", 0, fmt.Errorf("initializing stdin: %w", err)
	}

	for p.state.pc < len(p.ops) {
		if err = p.step(); err != nil {
			break
		}
	}
	return p.state.stdout.String(), p.opsCnt, err
}

func (p *Program) step() error {
	if p.opsCnt == p.opsLimit {
		return ErrOpsLimit
	}

	savedPc := p.state.pc
	if err := p.ops[p.state.pc].Apply(p.state); err != nil {
		return fmt.Errorf("executing op %v: %w", savedPc, err)
	}
	p.state.pc++
	p.opsCnt++
	return nil
}

func Compile(program string, opsLimit int) (*Program, error) {
	p := &Program{
		state:    &State{},
		ops:      make([]op, 0, len(program)),
		opsLimit: opsLimit,
	}

	// Precalc jump table.
	type bracket struct {
		Close bool
		Pos   int
	}
	var brStack []bracket
	jt := make(map[int]int)

	i := 0
	for _, c := range program {
		switch c {
		case '[':
			brStack = append(brStack, bracket{Pos: i})
			i++
		case ']':
			if len(brStack) == 0 {
				return nil, fmt.Errorf("mismatched bracket on position %d: %w", i, ErrCE)
			}

			br := brStack[len(brStack)-1]
			brStack = brStack[:len(brStack)-1]

			jt[br.Pos] = i
			jt[i] = br.Pos
			i++
		case '<', '>', '+', '-', '.', ',':
			i++
		}
	}
	if len(brStack) > 0 {
		return nil, fmt.Errorf("%d mismatched opening brackets", len(brStack))
	}

	// Compile.
	i = 0
	for _, c := range program {
		switch c {
		case '<':
			p.ops = append(p.ops, opPtrDec{})
			i++
		case '>':
			p.ops = append(p.ops, opPtrInc{})
			i++
		case '+':
			p.ops = append(p.ops, opInc{})
			i++
		case '-':
			p.ops = append(p.ops, opDec{})
			i++
		case '.':
			p.ops = append(p.ops, opPrint{})
			i++
		case ',':
			p.ops = append(p.ops, opRead{})
			i++
		case '[':
			p.ops = append(p.ops, opJmpFwd{to: jt[i]})
			i++
		case ']':
			p.ops = append(p.ops, opJmpBwd{to: jt[i]})
			i++
		}
	}

	return p, nil
}
