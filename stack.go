package jsonpath

// https://gist.github.com/bemasher/1777766
type stack struct {
	top  *element
	size int
}

type element struct {
	value interface{} // All types satisfy the empty interface, so we can store anything here.
	next  *element
}

// Return the stack's length
func (s *stack) Len() int {
	return s.size
}

// Push a new element onto the stack
func (s *stack) Push(value interface{}) {
	s.top = &element{value, s.top}
	s.size++
}

// Remove the top element from the stack and return it's value
// If the stack is empty, return nil
func (s *stack) Pop() (value interface{}) {
	if s.size > 0 {
		value, s.top = s.top.value, s.top.next
		s.size--
		return
	}
	return nil
}

func (s *stack) Peek() (value interface{}) {
	if s.size > 0 {
		return s.top.value
	}
	return nil
}
