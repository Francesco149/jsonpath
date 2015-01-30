package jsonpath

type intstack struct {
	values []int
}

// TODO: Kill intstack
func newIntStack() *intstack {
	return &intstack{
		values: make([]int, 0),
	}
}

func (s *intstack) Len() int {
	return len(s.values)
}

func (s *intstack) Push(r int) {
	s.values = append(s.values, r)
}

func (s *intstack) Pop() *int {
	if s.Len() == 0 {
		return nil
	}
	v := s.values[len(s.values)-1]
	s.values = s.values[:len(s.values)-1]
	return &v
}

func (s *intstack) Peek() *int {
	if s.Len() == 0 {
		return nil
	}
	v := s.values[len(s.values)-1]
	return &v
}

type stack struct {
	values []interface{}
}

func newStack() *stack {
	return &stack{
		values: make([]interface{}, 0),
	}
}

func (s *stack) Len() int {
	return len(s.values)
}

func (s *stack) Push(r interface{}) {
	s.values = append(s.values, r)
}

func (s *stack) Pop() interface{} {
	if s.Len() == 0 {
		return nil
	}
	v := s.Peek()
	s.values = s.values[:len(s.values)-1]
	return v
}

func (s *stack) Peek() interface{} {
	if s.Len() == 0 {
		return nil
	}
	v := s.values[len(s.values)-1]
	return v
}

func (s *stack) Clone() *stack {
	d := stack{
		values: make([]interface{}, s.Len()),
	}
	copy(d.values, s.values)
	return &d
}

func (s *stack) ToArray() []interface{} {
	return s.values
}

// // https://gist.github.com/bemasher/1777766
// type istack struct {
// 	top  *element
// 	size int
// }

// type element struct {
// 	value interface{} // All types satisfy the empty interface, so we can store anything here.
// 	next  *element
// }

// // Return the stack's length
// func (s *istack) Len() int {
// 	return s.size
// }

// // Push a new element onto the stack
// func (s *istack) Push(value interface{}) {
// 	s.top = &element{value, s.top}
// 	s.size++
// }

// // Remove the top element from the stack and return it's value
// // If the stack is empty, return nil
// func (s *istack) Pop() (value interface{}) {
// 	if s.size > 0 {
// 		value, s.top = s.top.value, s.top.next
// 		s.size--
// 		return
// 	}
// 	return nil
// }

// func (s *istack) Peek() (value interface{}) {
// 	if s.size > 0 {
// 		return s.top.value
// 	}
// 	return nil
// }
