package parser_test

import (
	"lang/pkg/lexer"
	"lang/pkg/parser"
	"testing"
)

func TestParserParseFunDecl(t *testing.T) {
	const test_program = `
unit test(bool a, int b) {
    c = 10;
    printf("%d", c);
}
`

	l := lexer.NewLexer("test", test_program)

	tokens, err := l.Lex()
	if err != nil {
		t.Fatal(err)
	}

	p := parser.NewParser(tokens)

	decl, _ := p.ParseFunctionDeclaration()
	if decl.TypeName.Value != "unit" {
		t.Fatalf("Expected type Void, got %s\n", decl.TypeName.Value)
	}

	if decl.Identifier.Value != "test" {
		t.Fatalf("Expected function name test, got %s\n", decl.Identifier.Value)
	}

	if len(decl.Parameters) != 2 {
		t.Fatalf("Expected 1 parameter, got %d\n", len(decl.Parameters))
	}

	if decl.Parameters[0].Name.Value != "a" {
		t.Fatalf("Expected parameter name b, got %s\n", decl.Parameters[0].Name.Value)
	}

	if decl.Parameters[0].TypeName.Value != "bool" {
		t.Fatalf("Expected parameter type Boolean, got %s\n", decl.Parameters[0].TypeName.Value)
	}
}

func TestParserExpect(t *testing.T) {
	const test_program = `
void test(int b) {
    int a == 10;
    printf("%d", a);
}
`

	l := lexer.NewLexer("test", test_program)

	tokens, err := l.Lex()
	if err != nil {
		t.Fatal(err)
	}

	p := parser.NewParser(tokens)
	p.Expect(lexer.Identifier, "void")
	p.Expect(lexer.Identifier, "test")
	p.Expect(lexer.Punctuator, "(")
	p.Expect(lexer.Identifier, "int")
	p.Expect(lexer.Identifier, "b")
	p.Expect(lexer.Punctuator, ")")
	p.Expect(lexer.Punctuator, "{")
	p.Expect(lexer.Identifier, "int")
	p.Expect(lexer.Identifier, "a")
	p.Expect(lexer.Operator, "==")
	p.Expect(lexer.Literal, "10")
	p.Expect(lexer.Punctuator, ";")
	p.Expect(lexer.Identifier, "printf")
	p.Expect(lexer.Punctuator, "(")
	p.Expect(lexer.Literal, "\"%d\"")
	p.Expect(lexer.Punctuator, ",")
	p.Expect(lexer.Identifier, "a")
	p.Expect(lexer.Punctuator, ")")
	p.Expect(lexer.Punctuator, ";")
	p.Expect(lexer.Punctuator, "}")
}

func TestParserError(t *testing.T) {
	const test_program = `
test(int b) {
    int a == 10;
    printf("%d", a);
}
`
	lexer := lexer.NewLexer("test", test_program)
	tokens, err := lexer.Lex()
	if err != nil {
		t.Fatal(err)
	}

	parser := parser.NewParser(tokens)
	_, err = parser.Parse()
	if err == nil {
		t.Logf("got no parse error")
		t.FailNow()
	}
}
