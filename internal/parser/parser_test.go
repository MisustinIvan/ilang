package parser

import (
	"fmt"
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

func tFatalf(t *testing.T, format string, args ...any) {
	t.Helper()
	t.Fatalf(format, args...)
}

func TestParseCondition(t *testing.T) {
	tests := []struct {
		input             string
		expectedCondition string
		expectedBody      string
		expectedElse      string // "" if no else clause
	}{
		{
			input:             "int main() { if true 1 }",
			expectedCondition: "true",
			expectedBody:      "1",
			expectedElse:      "",
		},
		{
			input:             "int main() { if false 0 else 1 }",
			expectedCondition: "false",
			expectedBody:      "0",
			expectedElse:      "1",
		},
	}

	for _, tt := range tests {
		l := lexer.New(lexer.NewSourceFile("test", tt.input))
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

		condition, ok := decl.Body.ImplicitReturn.(*ast.Condition)
		if !ok {
			t.Fatalf("Expected Condition expression, got %T", decl.Body.ImplicitReturn)
		}

		// Test Condition
		condLiteral, ok := condition.Condition.(*ast.Literal)
		if !ok {
			t.Fatalf("Expected Literal for condition, got %T", condition.Condition)
		}
		if condLiteral.Value != tt.expectedCondition {
			t.Fatalf("Expected condition to be '%s', got '%s'", tt.expectedCondition, condLiteral.Value)
		}

		// Test Body
		bodyLiteral, ok := condition.Body.(*ast.Literal)
		if !ok {
			t.Fatalf("Expected Literal for body, got %T", condition.Body)
		}
		if bodyLiteral.Value != tt.expectedBody {
			t.Fatalf("Expected body to be '%s', got '%s'", tt.expectedBody, bodyLiteral.Value)
		}

		// Test Else
		if tt.expectedElse == "" {
			if condition.Else != nil {
				t.Fatalf("Expected no else clause, but got one: %v", condition.Else)
			}
		} else {
			elseLiteral, ok := condition.Else.(*ast.Literal)
			if !ok {
				t.Fatalf("Expected Literal for else, got %T", condition.Else)
			}
			if elseLiteral.Value != tt.expectedElse {
				t.Fatalf("Expected else to be '%s', got '%s'", tt.expectedElse, elseLiteral.Value)
			}
		}
	}
}

func TestParseReturn(t *testing.T) {
	input := "int main() { return 1 }"
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
	if len(decl.Body.Body) != 0 {
		t.Fatalf("Expected 0 statements in body, got %d", len(decl.Body.Body))
	}

	if decl.Body.ImplicitReturn == nil {
		t.Fatalf("Expected implicit return")
	}

	ret, ok := decl.Body.ImplicitReturn.(*ast.Return)
	if !ok {
		t.Fatalf("Expected Return expression, got %T", decl.Body.ImplicitReturn)
	}

	val, ok := ret.Value.(*ast.Literal)
	if !ok {
		t.Fatalf("Expected Literal for return value, got %T", ret.Value)
	}
	if val.Value != "1" {
		t.Fatalf("Expected return value to be '1', got '%s'", val.Value)
	}
}

func TestParseBind(t *testing.T) {
	input := "int main() { let a: int = 1; }"
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
	if len(decl.Body.Body) != 1 {
		t.Fatalf("Expected 1 statement in body, got %d", len(decl.Body.Body))
	}

	bind, ok := decl.Body.Body[0].(*ast.Bind)
	if !ok {
		t.Fatalf("Expected Bind expression, got %T", decl.Body.Body[0])
	}

	if bind.Identifier.Name != "a" {
		t.Fatalf("Expected identifier 'a', got '%s'", bind.Identifier.Name)
	}

	if bind.Type != ast.Int {
		t.Fatalf("Expected type 'int', got '%s'", bind.Type)
	}

	val, ok := bind.Value.(*ast.Literal)
	if !ok {
		t.Fatalf("Expected Literal for bind value, got %T", bind.Value)
	}
	if val.Value != "1" {
		t.Fatalf("Expected bind value to be '1', got '%s'", val.Value)
	}
}

func TestParseAssignment(t *testing.T) {
	input := "int main() { a = 1; }"
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
	if len(decl.Body.Body) != 1 {
		t.Fatalf("Expected 1 statement in body, got %d", len(decl.Body.Body))
	}

	assign, ok := decl.Body.Body[0].(*ast.Assignment)
	if !ok {
		t.Fatalf("Expected Assignment expression, got %T", decl.Body.Body[0])
	}

	if assign.Identifier.Name != "a" {
		t.Fatalf("Expected identifier 'a', got '%s'", assign.Identifier.Name)
	}

	val, ok := assign.Value.(*ast.Literal)
	if !ok {
		t.Fatalf("Expected Literal for assignment value, got %T", assign.Value)
	}
	if val.Value != "1" {
		t.Fatalf("Expected assignment value to be '1', got '%s'", val.Value)
	}
}

func TestParseExamples(t *testing.T) {
	Examples := []string{
		`
		int add(int a, int b) {
			let result: int = a + b;
			return result;
		}
		`,
		`
		int main() {
			if true { if false 0 else 1 } else 2
		}
		`,
		`
		int calc(int x, int y) {
			return (1 + 2);
		}
		`,
		`
		extrn int printf(string format)
		int main() {
			let x: int = 10;
			printf("Hello, %d", x);
			return 0;
		}
		`,
		`
		int get_val() {
			{
				let a: int = 5;
				a + 5
			}
		}
		`,
	}

	for i, example := range Examples {
		t.Run(fmt.Sprintf("Example%d", i), func(t *testing.T) {
			l := lexer.New(lexer.NewSourceFile(fmt.Sprintf("example_%d.ilang", i), example))
			tokens, err := l.Lex()
			if err != nil {
				t.Fatalf("Lexing failed for example %d: %v", i, err)
			}

			p := New(tokens)
			_, err = p.Parse()
			if err != nil {
				t.Fatalf("Parsing failed for example %d: %v", i, err)
			}
		})
	}
}
