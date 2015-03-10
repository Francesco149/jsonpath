package jsonpath

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var boolTests = []struct {
	input    string
	fields   map[string]interface{}
	expected bool
}{
	{"true && false", nil, false},
	{"10 < 20", nil, true},
	{"10 < 20 && 20 < 30", nil, true},
	{"10 <= 20", nil, true},
	{"10 > 20", nil, false},
	{"20 >= 20", nil, true},
	{"20 > 20 || false", nil, false},
	//{"20 >= 20 || 2 == 2", nil, true},
	// {"20 > 5min && 5min < 13 && 5min > 1.99994", map[string]string{"5min": "10.23423"}, true},
	// {"20 < 5min || 5min < 13", map[string]string{"5min": "15.3423"}, false},
}

func TestBoolExpressions(t *testing.T) {
	as := assert.New(t)
	emptyFields := map[string]interface{}{}

	for _, test := range boolTests {
		if test.fields == nil {
			test.fields = emptyFields
		}

		lexer := NewSliceLexer([]byte(test.input), EXPRESSION)
		items := readerToArray(lexer)
		// trim EOF
		items = items[0 : len(items)-1]
		items_post, err := infixToPostFix(items)
		if as.NoError(err, "Could not transform to postfix\nTest: %q", test.input) {
			val, err := evaluatePostFix(items_post, test.fields)
			if as.NoError(err, "Could not evaluate postfix\nTest: %q\nError:%q", test.input, err) {
				as.Equal(test.expected, val, "%q  -> Actual: %t   Expected %t\n", test.input, val, test.expected)
			}
		}
	}
}
