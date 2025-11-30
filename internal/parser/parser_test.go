package parser

import (
	"testing"

	"github.com/MisustinIvan/ilang/internal/ast"
	"github.com/MisustinIvan/ilang/internal/lexer"
)

func TestParseBinary(t *testing.T) {
	input := "int main() { 1 + 2 }"
	l := lexer.New(lexer.NewSourceFile("test", input))
	tokens, err := l.Lex()
	if err != nil {
		t.Fatalf("Lexing failed: %v", err)
	}

	p := New(tokens)
	program, err := p.Parse()
	if err != nil {
		tFatalf(t, "Parsing failed: %v", err)
	}

	if len(program.Declarations) != 1 {
		t.Fatalf("Expected 1 declaration, got %d", len(program.Declarations))
	}

	decl := program.Declarations[0]
	if decl.Identifier.Name != "main" {
		t.Fatalf("Expected function name 'main', got '%s'", decl.Identifier.Name)
	}

	if len(decl.Body.Body) != 0 {
		t.Fatalf("Expected 0 statements in body, got %d", len(decl.Body.Body))
	}

	if decl.Body.ImplicitReturn == nil {
		t.Fatalf("Expected implicit return")
	}

	binary, ok := decl.Body.ImplicitReturn.(*ast.Binary)
	if !ok {
		t.Fatalf("Expected Binary expression, got %T", decl.Body.ImplicitReturn)
	}

	if binary.Operator != ast.Addition {
		t.Fatalf("Expected operator %s, got %s", ast.Addition, binary.Operator)
	}

	left, ok := binary.Left.(*ast.Literal)
	if !ok {
		tFatalf(t, "Expected Literal for left operand, got %T", binary.Left)
	}
	if left.Value != "1" {
		t.Fatalf("Expected left operand to be '1', got '%s'", left.Value)
	}

	right, ok := binary.Right.(*ast.Literal)
	if !ok {
		tFatalf(t, "Expected Literal for right operand, got %T", binary.Right)
	}
	if right.Value != "2" {
		t.Fatalf("Expected right operand to be '2', got '%s'", right.Value)
	}
}

func TestParseUnary(t *testing.T) {
	input := "int main() { -10 }"
	l := lexer.New(lexer.NewSourceFile("test", input))
	tokens, err := l.Lex()
	if err != nil {
		t.Fatalf("Lexing failed: %v", err)
	}

	p := New(tokens)
	program, err := p.Parse()
	if err != nil {
		tFatalf(t, "Parsing failed: %v", err)
	}

	if len(program.Declarations) != 1 {
		t.Fatalf("Expected 1 declaration, got %d", len(program.Declarations))
	}

	decl := program.Declarations[0]
	if decl.Identifier.Name != "main" {
		t.Fatalf("Expected function name 'main', got '%s'", decl.Identifier.Name)
	}

	if len(decl.Body.Body) != 0 {
		t.Fatalf("Expected 0 statements in body, got %d", len(decl.Body.Body))
	}

	if decl.Body.ImplicitReturn == nil {
		t.Fatalf("Expected implicit return")
	}

	unary, ok := decl.Body.ImplicitReturn.(*ast.Unary)
	if !ok {
		t.Fatalf("Expected Unary expression, got %T", decl.Body.ImplicitReturn)
	}

	if unary.Operator != ast.Inversion {
		t.Fatalf("Expected operator %s, got %s", ast.Inversion, unary.Operator)
	}

	literal, ok := unary.Value.(*ast.Literal)
	if !ok {
		tFatalf(t, "Expected Literal for value, got %T", unary.Value)
	}
	if literal.Value != "10" {
		t.Fatalf("Expected value to be '10', got '%s'", literal.Value)
	}
}

func tFatalf(t *testing.T, format string, args ...interface{}) {
	t.Helper()
	t.Fatalf(format, args...)
}
