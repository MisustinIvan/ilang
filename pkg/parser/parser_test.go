package parser_test

import (
	"lang/pkg/ast"
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

	l := lexer.NewLexer(test_program)

	err := l.Lex()
	if err != nil {
		t.Fatalf("%v", err)
	}

	p := parser.NewParser(l.Tokens())

	decl, _ := p.ParseFunctionDeclaration()
	if decl.Type != ast.Unit {
		t.Fatalf("Expected type Void, got %s\n", decl.Type.String())
	}

	if decl.Name.Value != "test" {
		t.Fatalf("Expected function name test, got %s\n", decl.Name.Value)
	}

	if len(decl.Parameters) != 2 {
		t.Fatalf("Expected 1 parameter, got %d\n", len(decl.Parameters))
	}

	if decl.Parameters[0].Name.Value != "a" {
		t.Fatalf("Expected parameter name b, got %s\n", decl.Parameters[0].Name.Value)
	}

	if decl.Parameters[0].Type != ast.Boolean {
		t.Fatalf("Expected parameter type Boolean, got %s\n", decl.Parameters[0].Type.String())
	}
}

func TestParserExpect(t *testing.T) {
	const test_program = `
void test(int b) {
    int a == 10;
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
