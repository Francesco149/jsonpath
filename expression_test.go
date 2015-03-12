package jsonpath

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var exprBoolTests = []struct {
	input         string
	fields        map[string]Item
	expectedValue bool
}{
	// &&
	{"true && true", nil, true},
	{"false && true", nil, false},
	{"false && false", nil, false},

	// ||
	{"true || true", nil, true},
	{"true || false", nil, true},
	{"false ||  false", nil, false},

	// LT
	{"10 < 20", nil, true},
	{"10 < 10", nil, false},
	{"100 < 20", nil, false},
	{"@a < 50", map[string]Item{"@a": genValue(`49`, jsonNumber)}, true},
	{"@a < 50", map[string]Item{"@a": genValue(`50`, jsonNumber)}, false},
	{"@a < 50", map[string]Item{"@a": genValue(`51`, jsonNumber)}, false},

	// LE
	{"10 <= 20", nil, true},
	{"10 <= 10", nil, true},
	{"100 <= 20", nil, false},
	{"@a <= 54", map[string]Item{"@a": genValue(`53`, jsonNumber)}, true},
	{"@a <= 54", map[string]Item{"@a": genValue(`54`, jsonNumber)}, true},
	{"@a <= 54", map[string]Item{"@a": genValue(`55`, jsonNumber)}, false},

	// GT
	{"30 > 20", nil, true},
	{"20 > 20", nil, false},
	{"10 > 20", nil, false},
	{"@a > 50", map[string]Item{"@a": genValue(`49`, jsonNumber)}, false},
	{"@a > 50", map[string]Item{"@a": genValue(`50`, jsonNumber)}, false},
	{"@a > 50", map[string]Item{"@a": genValue(`51`, jsonNumber)}, true},

	// GE
	{"30 >= 20", nil, true},
	{"20 >= 20", nil, true},
	{"10 >= 20", nil, false},
	{"@a >= 50", map[string]Item{"@a": genValue(`49`, jsonNumber)}, false},
	{"@a >= 50", map[string]Item{"@a": genValue(`50`, jsonNumber)}, true},
	{"@a >= 50", map[string]Item{"@a": genValue(`51`, jsonNumber)}, true},

	// EQ
	{"20 == 20", nil, true},
	{"20 == 21", nil, false},
	{"true == true", nil, true},
	{"true == false", nil, false},
	{"@a == @b", map[string]Item{"@a": genValue(`"one"`, jsonString), "@b": genValue(`"one"`, jsonString)}, true},
	{"@a == @b", map[string]Item{"@a": genValue(`"one"`, jsonString), "@b": genValue(`"two"`, jsonString)}, false},
	{`"fire" == "fire"`, nil, true},
	{`"fire" == "water"`, nil, false},
	{`@a == "toronto"`, map[string]Item{"@a": genValue(`"toronto"`, jsonString)}, true},
	{`@a == "toronto"`, map[string]Item{"@a": genValue(`"los angeles"`, jsonString)}, false},
	{`@a == 3.4`, map[string]Item{"@a": genValue(`3.4`, jsonNumber)}, true},
	{`@a == 3.4`, map[string]Item{"@a": genValue(`3.41`, jsonNumber)}, false},
	{`@a == null`, map[string]Item{"@a": genValue(`null`, jsonNull)}, true},

	// NEQ
	{"20 != 20", nil, false},
	{"20 != 21", nil, true},
	{"true != true", nil, false},
	{"true != false", nil, true},
	{"@a != @b", map[string]Item{"@a": genValue(`"one"`, jsonString), "@b": genValue(`"one"`, jsonString)}, false},
	{"@a != @b", map[string]Item{"@a": genValue(`"one"`, jsonString), "@b": genValue(`"two"`, jsonString)}, true},
	{`"fire" != "fire"`, nil, false},
	{`"fire" != "water"`, nil, true},
	{`@a != "toronto"`, map[string]Item{"@a": genValue(`"toronto"`, jsonString)}, false},
	{`@a != "toronto"`, map[string]Item{"@a": genValue(`"los angeles"`, jsonString)}, true},
	{`@a != 3.4`, map[string]Item{"@a": genValue(`3.4`, jsonNumber)}, false},
	{`@a != 3.4`, map[string]Item{"@a": genValue(`3.41`, jsonNumber)}, true},
	{`@a != null`, map[string]Item{"@a": genValue(`null`, jsonNull)}, false},

	// Plus
	{"20 + 7 == 27", nil, true},
	{"20 + 6 == 27", nil, false},
	{"20 + 6.999999 == 27", nil, false},

	// Minus
	{"20 - 7 == 13", nil, true},
	{"20 - 6 == 13", nil, false},
	{"20 - 6.999999 == 13", nil, false},

	// Negate
	{"!true", nil, false},
	{"!false", nil, true},

	// Mix
	{"20 >= 20 || 2 == 2", nil, true},
	{"20 > $.test && $.test < 13 && $.test > 1.99994", map[string]Item{"$.test": genValue(`10.23423`, jsonNumber)}, true},
	{"20 > $.test && $.test < 13 && $.test > 1.99994", map[string]Item{"$.test": genValue(`15.3423`, jsonNumber)}, false},
}

func genValue(val string, typ int) Item {
	return Item{
		val: []byte(val),
		typ: typ,
	}
}

func TestExpressions(t *testing.T) {
	as := assert.New(t)
	emptyFields := map[string]Item{}

	for _, test := range exprBoolTests {
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
			if as.NoError(err, "Could not evaluate postfix\nTest Input: %q\nTest Values:%q\nError:%q", test.input, test.fields, err) {
				as.Equal(test.expectedValue, val, "%q  -> Actual: %t   Expected %t\n", test.input, val, test.expectedValue)
			}
		}
	}
}

var exprErrorTests = []struct {
	input                  string
	fields                 map[string]Item
	expectedErrorSubstring string
}{
	{"@a == @b", map[string]Item{"@a": genValue(`"one"`, jsonString), "@b": genValue("3.4", jsonNumber)}, "cannot be compared"},
	{"20 == null", nil, "cannot be compared"},
	{`"toronto" == null`, nil, "cannot be compared"},
	{`@a == null`, map[string]Item{"@a": genValue(`3.41`, jsonNumber)}, "cannot be compared"},
}

func TestBadExpressions(t *testing.T) {
	as := assert.New(t)
	emptyFields := map[string]Item{}

	for _, test := range exprErrorTests {
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
			as.False(val, "Expected false when error occurs")
			if as.Error(err, "Could not evaluate postfix\nTest Input: %q\nTest Values:%q\nError:%q", test.input, test.fields, err) {
				as.True(strings.Contains(err.Error(), test.expectedErrorSubstring), "Test Input: %q\nError %q does not contain %q", test.input, err.Error(), test.expectedErrorSubstring)
			}

		}
	}
}
