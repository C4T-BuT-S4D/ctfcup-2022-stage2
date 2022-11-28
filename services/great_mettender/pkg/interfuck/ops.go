package interfuck

type op interface {
	Apply(s *State) error
}

type opPtrDec struct{}

func (opPtrDec) Apply(s *State) error {
	if s.ptr == 0 {
		return ErrOOB
	}
	s.ptr--
	return nil
}

type opPtrInc struct{}

func (opPtrInc) Apply(s *State) error {
	if s.ptr == len(s.mem)-1 {
		return ErrOOB
	}
	s.ptr++
	return nil
}

type opInc struct{}

func (opInc) Apply(s *State) error {
	s.mem[s.ptr]++
	return nil
}

type opDec struct{}

func (opDec) Apply(s *State) error {
	s.mem[s.ptr]--
	return nil
}

type opJmpFwd struct {
	to int
}

func (o opJmpFwd) Apply(s *State) error {
	if s.mem[s.ptr] == 0 {
		s.pc = o.to
	}
	return nil
}

type opJmpBwd struct {
	to int
}

func (o opJmpBwd) Apply(s *State) error {
	if s.mem[s.ptr] != 0 {
		s.pc = o.to
	}
	return nil
}

type opPrint struct{}

func (opPrint) Apply(s *State) error {
	s.stdout.WriteByte(byte(s.mem[s.ptr] % 256))
	return nil
}

type opRead struct{}

func (opRead) Apply(s *State) error {
	b, err := s.stdin.ReadByte()
	if err != nil {
		b = 0
	}
	s.mem[s.ptr] = uint16(b)
	return nil
}
