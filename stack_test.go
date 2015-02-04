package jsonpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStackPush(t *testing.T) {
	as := assert.New(t)
	s := newStack()

	s.push(5)
	as.Equal(s.len(), 1)

	s.push(12)
	as.Equal(s.len(), 2)
}

func TestStackPop(t *testing.T) {
	as := assert.New(t)
	s := newStack()

	s.push(99)
	as.Equal(s.len(), 1)

	v := s.pop()
	as.NotNil(v)
	iv, ok := v.(int)
	as.True(ok)
	as.Equal(99, iv)

	as.Equal(s.len(), 0)
}

func TestStackPeek(t *testing.T) {
	as := assert.New(t)
	s := newStack()

	s.push(99)
	v := s.peek()
	as.NotNil(v)
	iv, ok := v.(int)
	as.True(ok)
	as.Equal(99, iv)

	s.push(54)
	v = s.peek()
	as.NotNil(v)
	iv, ok = v.(int)
	as.True(ok)
	as.Equal(54, iv)

	s.pop()
	v = s.peek()
	as.NotNil(v)
	iv, ok = v.(int)
	as.True(ok)
	as.Equal(99, iv)

	s.pop()
	v = s.peek()
	as.Nil(v)
}
