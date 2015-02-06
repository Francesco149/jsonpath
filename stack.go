package jsonpath

type stack struct {
	values []int
}

func newStack() *stack {
	return &stack{
		values: make([]int, 0),
	}
}

func (s *stack) len() int {
	return len(s.values)
}

func (s *stack) push(r int) {
	s.values = append(s.values, r)
}

func (s *stack) pop() (int, bool) {
	if s.len() == 0 {
		return 0, false
	}
	v, _ := s.peek()
	s.values = s.values[:len(s.values)-1]
	return v, true
}

func (s *stack) peek() (int, bool) {
	if s.len() == 0 {
		return 0, false
	}
	v := s.values[len(s.values)-1]
	return v, true
}

func (s *stack) clone() *stack {
	d := stack{
		values: make([]int, s.len()),
	}
	copy(d.values, s.values)
	return &d
}

func (s *stack) toArray() []int {
	return s.values
}
