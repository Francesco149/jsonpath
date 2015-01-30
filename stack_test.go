package jsonpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStackPush(t *testing.T) {
	as := assert.New(t)
	s := newStack()

	s.Push(5)
	as.Equal(s.Len(), 1)

	s.Push(12)
	as.Equal(s.Len(), 2)
}

func TestStackPop(t *testing.T) {
	as := assert.New(t)
	s := newStack()

	s.Push(99)
	as.Equal(s.Len(), 1)

	v := s.Pop()
	as.NotNil(v)
	iv, ok := v.(int)
	as.True(ok)
	as.Equal(99, iv)

	as.Equal(s.Len(), 0)
}

func TestStackPeek(t *testing.T) {
	as := assert.New(t)
	s := newStack()

	s.Push(99)
	v := s.Peek()
	as.NotNil(v)
	iv, ok := v.(int)
	as.True(ok)
	as.Equal(99, iv)

	s.Push(54)
	v = s.Peek()
	as.NotNil(v)
	iv, ok = v.(int)
	as.True(ok)
	as.Equal(54, iv)

	s.Pop()
	v = s.Peek()
	as.NotNil(v)
	iv, ok = v.(int)
	as.True(ok)
	as.Equal(99, iv)

	s.Pop()
	v = s.Peek()
	as.Nil(v)
}
