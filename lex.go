package jsonpath

type Pos int
type stateFn func(lexer, *stack) stateFn

const (
	lexError = 0 // must match jsonError and pathError
	eof      = -1
	noValue  = -2
)

type Item struct {
	typ int
	pos Pos // The starting position, in bytes, of this Item in the input string.
	val []byte
}

// Used by evaluator
type tokenReader interface {
	next() (*Item, bool)
}

// Used by state functions
type lexer interface {
	tokenReader
	take() int
	peek() int
	emit(int)
	ignore()
	errorf(string, ...interface{}) stateFn
	reset()
}

type lex struct {
	initialState   stateFn
	currentStateFn stateFn
	item           Item
	hasItem        bool
	stack          stack
}

func newLex(initial stateFn) lex {
	return lex{
		initialState:   initial,
		currentStateFn: initial,
		item:           Item{},
		stack:          *newStack(),
	}
}

func itemsDescription(items []Item, nameMap map[int]string) []string {
	vals := make([]string, len(items))
	for i, item := range items {
		vals[i] = itemDescription(&item, nameMap)
	}
	return vals
}

func itemDescription(item *Item, nameMap map[int]string) string {
	var found bool
	val, found := nameMap[item.typ]
	if !found {
		return string(item.val)
	}
	return val
}
