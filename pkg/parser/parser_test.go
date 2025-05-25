package parser_test

import (
	"fmt"
	"lang/pkg/ast"
	"lang/pkg/lexer"
	"lang/pkg/parser"
	"testing"
)

func TestParserParseFunDecl(t *testing.T) {
	const test_program = `
void test(bool a, int b) {
    c = 10;
    printf("%d", c);
}
`

	l := lexer.NewLexer(test_program)

	err := l.Lex()
	if err != nil {
		t.Fatalf("%v", err)
	}

	p := parser.NewParser(l.Tokens())

	decl := p.ParseFunctionDeclaration()
	if decl.Type != ast.Void {
		t.Fatalf("Expected type Void, got %s\n", decl.Type.String())
	}

	if decl.Name.Name != "test" {
		t.Fatalf("Expected function name test, got %s\n", decl.Name.Name)
	}

	if len(decl.ParameterTypes) != 2 {
		t.Fatalf("Expected 1 parameter, got %d\n", len(decl.ParameterTypes))
	}

	if decl.ParameterTypes[0].Name.Name != "a" {
		t.Fatalf("Expected parameter name b, got %s\n", decl.ParameterTypes[0].Name.Name)
	}

	if decl.ParameterTypes[0].Type != ast.Boolean {
		t.Fatalf("Expected parameter type Boolean, got %s\n", decl.ParameterTypes[0].Type.String())
	}

	fmt.Printf("decl.Body: %v\n", decl.Body[0])

	t.Fail()
}

func TestParserExpect(t *testing.T) {
	const test_program = `
void test(int b) {
    int a = 10;
    printf("%d", a);
}
`

	l := lexer.NewLexer(test_program)

	err := l.Lex()
	if err != nil {
		t.Fatalf("%v", err)
	}

	p := parser.NewParser(l.Tokens())
	p.Expect(lexer.Identifier, "void")
	p.Expect(lexer.Identifier, "test")
	p.Expect(lexer.Punctuator, "(")
	p.Expect(lexer.Identifier, "int")
	p.Expect(lexer.Identifier, "b")
	p.Expect(lexer.Punctuator, ")")
	p.Expect(lexer.Punctuator, "{")
	p.Expect(lexer.Identifier, "int")
	p.Expect(lexer.Identifier, "a")
	p.Expect(lexer.Operator, "=")
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
