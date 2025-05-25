package parser

import (
	"fmt"
	"lang/pkg/ast"
	"lang/pkg/lexer"
	"os"
)

type Parser struct {
	tokens []lexer.Token
	head   int
}

func NewParser(tokens []lexer.Token) Parser {
	return Parser{
		tokens: tokens,
		head:   0,
	}
}

func (p *Parser) Peek() *lexer.Token {
	if p.head >= len(p.tokens) {
		return nil
	}
	return &p.tokens[p.head]
}

func (p *Parser) PeekNext(offset int) *lexer.Token {
	if p.head+offset >= len(p.tokens) {
		return nil
	}
	return &p.tokens[p.head+offset]
}

func (p *Parser) Next() lexer.Token {
	if p.head >= len(p.tokens) {
		fmt.Printf("Unexpected EOF")
		os.Exit(-1)
	}
	tk := p.tokens[p.head]
	p.head++
	return tk
}

// can exits
func (p *Parser) MatchCurrent(kind lexer.TokenKind, value string) bool {
	c := p.Peek()
	if c == nil {
		fmt.Printf("Unexpected EOF\n")
		os.Exit(-1)
	}
	if kind == c.Kind && (value == c.Value || value == "") {
		return true
	}
	return false
}

// can't exit
func (p *Parser) MatchNext(kind lexer.TokenKind, value string, offset int) bool {
	c := p.PeekNext(offset)
	if c == nil {
		return false
	}
	if kind == c.Kind && (value == c.Value || value == "") {
		return true
	}
	return false
}

func (p *Parser) Expect(kind lexer.TokenKind, value string) lexer.Token {
	if p.head >= len(p.tokens) {
		fmt.Printf("Unexpected EOF, expected %s %s\n", kind.String(), value)
		os.Exit(-1)
	}

	c := p.Next()

	if kind != c.Kind || (value != c.Value && value != "") {
		if value != "" {
			fmt.Printf("Expected %s %s, got %s \"%s\" at %s\n", kind.String(), value, c.Kind.String(), c.Value, c.Position.String())
		} else {
			fmt.Printf("Expected any %s, got %s \"%s\" at %s\n", kind.String(), c.Kind.String(), c.Value, c.Position.String())
		}
		os.Exit(-1)
	}
	return c
}

func (p *Parser) ParseExpression() ast.Expression {
	// determine kind of expression
	// right now we have literal expression, identifier expression and call expression
	// the easiest is identifier expression, which is just <identifier>
	// then we have the literal expression, which is just <literal>
	// then we have the call expression which is just <identifier>'(' [<expression { ',' <expression>}] ')'

	// TODO - better handling of identifiing just identifier...
	if p.MatchNext(lexer.Identifier, "", 0) && (p.MatchNext(lexer.Punctuator, ";", 1) || p.MatchNext(lexer.Punctuator, ")", 1)) {
		return &ast.IdentifierExpression{
			Identifier: ast.Identifier{
				Name: p.Next().Value,
			},
		}
	}

	if p.MatchNext(lexer.Literal, "", 0) {
		tk := p.Next()
		t, ok := ast.LiteralType(tk)
		if !ok {
			fmt.Printf("Unknown literal type at %s\n", tk.Position.String())
		}

		return &ast.LiteralExpression{
			Type:  t,
			Value: tk.Value,
		}
	}

	if p.MatchNext(lexer.Identifier, "", 0) && p.MatchNext(lexer.Punctuator, "(", 1) {
		f := p.Next()
		args := []ast.Expression{}

		// consume the bracket
		p.Next()

		// parse the args
		requires_comma := false
		for !p.MatchCurrent(lexer.Punctuator, ")") {
			if requires_comma {
				p.Expect(lexer.Punctuator, ",")
			} else {
				requires_comma = true
			}
			args = append(args, p.ParseExpression())
		}

		// consume the bracket
		p.Next()

		return &ast.CallExpression{
			Function: ast.Identifier{
				Name: f.Value,
			},
			Args: args,
		}
	}

	c := p.Peek()
	if c == nil {
		fmt.Printf("Unexpected EOF while parsing expression\n")
		os.Exit(-1)
	}

	fmt.Printf("Unknown expression at %s starting with %s\n", c.Position.String(), c.Value)
	os.Exit(-1)
	return nil
}

func (p *Parser) ParseAssignment() *ast.AssignmentStatement {
	left := p.Expect(lexer.Identifier, "")

	p.Expect(lexer.Operator, "=")

	right := p.ParseExpression()

	s := &ast.AssignmentStatement{
		Left: ast.Identifier{
			Name: left.Value,
		},
		Right: right,
	}
	return s
}

func (p *Parser) ParseVariableDeclaration() *ast.VariableDeclarationStatement {
	var_type_tk := p.Expect(lexer.Identifier, "")
	var_type, ok := ast.Types[var_type_tk.Value]
	if !ok {
		fmt.Printf("Unknown type %s at %s", var_type_tk.Value, var_type_tk.Position.String())
		os.Exit(-1)
	}

	var_name_tk := p.Expect(lexer.Identifier, "")

	p.Expect(lexer.Operator, "=")

	right := p.ParseExpression()

	return &ast.VariableDeclarationStatement{
		Left: ast.Identifier{
			Name: var_name_tk.Value,
		},
		Type:  var_type,
		Right: right,
	}
}

func (p *Parser) ParseFunctionCall() *ast.FunctionCallStatement {
	f := p.Expect(lexer.Identifier, "")

	p.Expect(lexer.Punctuator, "(")

	args := []ast.Expression{}

	requires_comma := false
	for !p.MatchCurrent(lexer.Punctuator, ")") {
		if requires_comma {
			p.Expect(lexer.Punctuator, ",")
		} else {
			requires_comma = true
		}
		args = append(args, p.ParseExpression())
	}

	// consume the bracket
	p.Expect(lexer.Punctuator, ")")

	return &ast.FunctionCallStatement{
		Function: ast.Identifier{
			Name: f.Value,
		},
		Args: args,
	}
}

func (p *Parser) ParseReturn() *ast.ReturnStatement {
	// consume the "return"
	p.Next()
	var e ast.Expression

	if p.MatchNext(lexer.Punctuator, ";", 0) {
		e = &ast.EmptyExpression{}
	} else {
		e = p.ParseExpression()
	}

	return &ast.ReturnStatement{
		Value: e,
	}
}

func (p *Parser) ParseStatement() ast.Statement {
	// determine kind of statement
	// right now we have assignment, variable declaration and function call
	// assignment is the easiest, just an <identifier> '=' <expression>';'
	// then we have function call which is <identifier> '(' [<expression> { ',' <expression> } ] ')'
	// then we have variable declaration which is <type> <identifier> '=' <expression>
	// then we have return statement which is 'return' <expression>

	if p.MatchNext(lexer.Identifier, "", 0) && p.MatchNext(lexer.Operator, "=", 1) {
		return p.ParseAssignment()
	}

	if p.MatchNext(lexer.Identifier, "", 0) && p.MatchNext(lexer.Identifier, "", 1) && p.MatchNext(lexer.Operator, "", 2) {
		return p.ParseVariableDeclaration()
	}

	if p.MatchNext(lexer.Identifier, "", 0) && p.MatchNext(lexer.Punctuator, "(", 1) {
		return p.ParseFunctionCall()
	}

	if p.MatchNext(lexer.Keyword, "return", 0) {
		return p.ParseReturn()
	}

	c := p.Next()
	return &ast.CommentStatement{
		Value: c.Value,
	}
}

func (p *Parser) ParseFunctionDeclaration() ast.FunctionDeclaration {
	f_name := ast.Identifier{}
	var f_type ast.Type
	f_params := []ast.ParameterType{}
	f_body := []ast.Statement{}

	// parse function type
	type_tk := p.Expect(lexer.Identifier, "")
	f_type, ok := ast.Types[type_tk.Value]
	if !ok {
		fmt.Printf("Unknown function return type %s at %s\n", type_tk.Value, type_tk.Position.String())
		os.Exit(-1)
	}

	name_tk := p.Expect(lexer.Identifier, "")
	f_name = ast.Identifier{
		Name: name_tk.Value,
	}

	p.Expect(lexer.Punctuator, "(")

	param_names := map[string]bool{}

	requires_comma := false

	for !p.MatchCurrent(lexer.Punctuator, ")") {
		if requires_comma {
			if !p.MatchCurrent(lexer.Punctuator, ",") {
				fmt.Printf("Expected comma at %s, got %s %s", p.Peek().Position.String(), p.Peek().Kind.String(), p.Peek().Value)
				os.Exit(-1)
			} else {
				// consume the comma
				p.Next()
			}
		} else {
			requires_comma = true
		}
		param_type_tk := p.Expect(lexer.Identifier, "")
		param_type, ok := ast.Types[param_type_tk.Value]
		if !ok {
			fmt.Printf("Unknown parameter type %s at %s\n", param_type_tk.Value, param_type_tk.Position.String())
			os.Exit(-1)
		}

		param_name_tk := p.Expect(lexer.Identifier, "")

		if param_names[param_name_tk.Value] {
			fmt.Printf("Unexpected parameter name at %s, parameter %s aleready declared", param_name_tk.Position.String(), param_name_tk.Value)
			os.Exit(-1)
		}

		param_names[param_name_tk.Value] = true

		f_params = append(f_params, ast.ParameterType{
			Type: param_type,
			Name: ast.Identifier{
				Name: param_name_tk.Value,
			},
		})
	}

	p.Expect(lexer.Punctuator, ")")
	p.Expect(lexer.Punctuator, "{")

	for !p.MatchCurrent(lexer.Punctuator, "}") {
		f_body = append(f_body, p.ParseStatement())
		p.Expect(lexer.Punctuator, ";")
	}

	p.Next()

	return ast.FunctionDeclaration{
		Name:           f_name,
		Type:           f_type,
		ParameterTypes: f_params,
		Body:           f_body,
	}
}

func (p *Parser) Parse() ast.Prog {
	res := ast.Prog{}

	for p.head < len(p.tokens) {
		res.Declarations = append(res.Declarations, p.ParseFunctionDeclaration())
	}

	return res
}
