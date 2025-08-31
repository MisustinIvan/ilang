package lexer_test

import (
	"lang/pkg/lexer"
	"testing"
)

const test_program = `
func test() {
    const a = 10;
    printf("%d", a);
}
`

func TestLexer(t *testing.T) {
	l := lexer.NewLexer(test_program)

	_, err := l.Lex()
	if err != nil {
		t.Fatalf("%v", err)
	}
}
