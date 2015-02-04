package jsonpath

type stack interface {
	len() int
	push(interface{})
	pop() interface{}
	peek() interface{}
	clone() stack
	toArray() []interface{}
}

type arrayStack struct {
	values []interface{}
}

func newStack() *arrayStack {
	return &arrayStack{
		values: make([]interface{}, 0),
	}
}

func (s *arrayStack) len() int {
	return len(s.values)
}

func (s *arrayStack) push(r interface{}) {
	s.values = append(s.values, r)
}

func (s *arrayStack) pop() interface{} {
	if s.len() == 0 {
		return nil
	}
	v := s.peek()
	s.values = s.values[:len(s.values)-1]
	return v
}

func (s *arrayStack) peek() interface{} {
	if s.len() == 0 {
		return nil
	}
	v := s.values[len(s.values)-1]
	return v
}

func (s *arrayStack) clone() stack {
	d := arrayStack{
		values: make([]interface{}, s.len()),
	}
	copy(d.values, s.values)
	return stack(&d)
}

func (s *arrayStack) toArray() []interface{} {
	return s.values
}
