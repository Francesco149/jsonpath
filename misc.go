package jsonpath

import (
	"fmt"
)

func takeExponent(l lexer) error {
	r := l.peek()
	if r != 'e' && r != 'E' {
		return nil
	}
	l.take()
	r = l.take()
	switch r {
	case '+', '-':
		// Check digit immediately follows sign
		if d := l.peek(); !(d >= '0' && d <= '9') {
			return fmt.Errorf("Expected digit after numeric sign instead of %#U", d)
		}
		takeDigits(l)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		takeDigits(l)
	default:
		return fmt.Errorf("Expected digit after 'e' instead of %#U", r)
	}
	return nil
}

func takeJSONNumeric(l lexer) error {
	cur := l.take()
	switch cur {
	case '-':
		// Check digit immediately follows sign
		if d := l.peek(); !(d >= '0' && d <= '9') {
			return fmt.Errorf("Expected digit after dash instead of %#U", d)
		}
		takeDigits(l)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		takeDigits(l)
	default:
		return fmt.Errorf("Expected digit or dash instead of %#U", cur)
	}

	// fraction or exponent
	cur = l.peek()
	switch cur {
	case '.':
		l.take()
		// Check digit immediately follows period
		if d := l.peek(); !(d >= '0' && d <= '9') {
			return fmt.Errorf("Expected digit after '.' instead of %#U", d)
		}
		takeDigits(l)
		if err := takeExponent(l); err != nil {
			return err
		}
	case 'e', 'E':
		if err := takeExponent(l); err != nil {
			return err
		}
	}

	return nil
}

func takeDigits(l lexer) {
	for {
		d := l.peek()
		if d >= '0' && d <= '9' {
			l.take()
		} else {
			break
		}
	}
}

// Only used at the very beginning of parsing. After that, the emit() function
// automatically skips whitespace.
func ignoreSpaceRun(l lexer) {
	for {
		r := l.peek()
		if r == ' ' || r == '\t' || r == '\r' || r == '\n' {
			l.take()
		} else {
			break
		}
	}
	l.ignore()
}

func takeExactSequence(l lexer, str []byte) bool {
	for _, r := range str {
		v := l.take()
		if v != int(r) {
			return false
		}
	}
	return true
}

func readerToArray(tr tokenReader) []Item {
	vals := make([]Item, 0)
	for {
		i, ok := tr.next()
		if !ok {
			break
		}
		v := *i
		s := make([]byte, len(v.val))
		copy(s, v.val)
		v.val = s
		vals = append(vals, v)
	}
	return vals
}

// func print(q *query, i *Item) queryStateFn {
// 	printLoc(q.state.location.toArray(), i.val)
// 	return print
// }

// func printLoc(s []interface{}, vals ...interface{}) {
// 	for _, s := range s {
// 		switch v := s.(type) {
// 		case []byte:
// 			fmt.Printf("%s ", string(v))
// 		default:
// 			fmt.Printf("%v ", v)
// 		}
// 	}
// 	for _, v := range vals {
// 		switch i := v.(type) {
// 		case []byte:
// 			fmt.Printf("%s ", string(i))
// 		default:
// 			fmt.Printf("%v ", i)
// 		}
// 	}
// 	fmt.Println("")
// }

// func printLevels(s []int) {
// 	for _, s := range s {
// 		fmt.Printf("%v ", jsonTokenNames[s])
// 	}
// 	fmt.Println("")
// }
