package jsonpath

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var expressionTests = []lexTest{
	{"empty", "", []int{exprEOF}},
	{"spaces", "     \t\r\n", []int{exprEOF}},
	{"numbers", "1 1.2 1.3e10 0 1.2", []int{exprNumber, exprNumber, exprNumber, exprNumber, exprNumber, exprNumber, exprEOF}},
	// {"numbers with signs", "+1 -2.23", []int{exprNumber, exprOpPlus, exprNumber, exprEOF}},
	{"paths", "@.aKey[2].bKey $.sKey", []int{exprPath, exprPath, exprEOF}},
	{"expressions without spaces", "4+-19", []int{exprNumber, exprOpPlus, exprOpMinus, exprNumber, exprEOF}},
	{"expressions without spaces x2", "4+19", []int{exprNumber, exprOpPlus, exprNumber, exprEOF}},
	{"expressions without spaces x3", "4-19", []int{exprNumber, exprOpMinus, exprNumber, exprEOF}},

	{"parens", "(  () () )", []int{exprParenLeft, exprParenLeft, exprParenRight, exprParenLeft, exprParenRight, exprParenRight, exprEOF}},
	{"operations", "+-==", []int{exprOpPlus, exprOpMinus, exprOpEq, exprEOF}},
	{"numerical comparisons", " <<=>>=", []int{exprOpLt, exprOpLe, exprOpGt, exprOpGe, exprEOF}},
}

func TestExpressionTokens(t *testing.T) {
	as := assert.New(t)
	for _, test := range expressionTests {
		lexer := NewSliceLexer([]byte(test.input), EXPRESSION)
		items := readerToArray(lexer)
		types := itemsToTypes(items)

		for _, i := range items {
			if i.typ == exprError {
				fmt.Println(string(i.val))
			}
		}

		as.Equal(types, test.tokenTypes, "Testing of %s: \nactual\n\t%+v\nexpected\n\t%v", test.name, typesDescription(types, exprTokenNames), typesDescription(test.tokenTypes, exprTokenNames))
	}
}
