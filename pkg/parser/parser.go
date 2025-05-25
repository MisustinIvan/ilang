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

// can exit
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

func (p *Parser) ParseBindExpression() *ast.BindExpression {
	var e_left ast.Identifier
	var e_type ast.Type
	var e_right ast.Expression

	p.Expect(lexer.Keyword, "let")

	left_tk := p.Expect(lexer.Identifier, "")
	e_left = ast.Identifier{
		Value: left_tk.Value,
	}

	p.Expect(lexer.Punctuator, ":")

	type_tk := p.Expect(lexer.Identifier, "")
	e_type, ok := ast.Types[type_tk.Value]
	if !ok {
		fmt.Printf("Unknown type %s at %s\n", type_tk.Value, type_tk.Position.String())
		os.Exit(-1)
	}

	p.Expect(lexer.Operator, "=")

	e_right = p.ParseExpression()

	return &ast.BindExpression{
		Left:  e_left,
		Type:  e_type,
		Right: e_right,
	}
}

func (p *Parser) ParseReturnExpression() *ast.ReturnExpression {
	var e_value ast.Expression

	p.Expect(lexer.Keyword, "return")

	e_value = p.ParseExpression()

	return &ast.ReturnExpression{
		Value: e_value,
	}
}

func (p *Parser) ParseAssignmentExpression() *ast.AssignmentExpression {
	var e_left ast.Identifier
	var e_right ast.Expression

	left_tk := p.Expect(lexer.Identifier, "")
	e_left = ast.Identifier{
		Value: left_tk.Value,
	}

	p.Expect(lexer.Operator, "=")

	e_right = p.ParseExpression()

	return &ast.AssignmentExpression{
		Left:  e_left,
		Right: e_right,
	}
}

func (p *Parser) ParseLiteral() *ast.Literal {
	l_value := p.Expect(lexer.Literal, "")
	l_type := ast.LiteralType(l_value.Value)

	return &ast.Literal{
		Value: l_value.Value,
		Type:  l_type,
	}
}

func (p *Parser) ParseFunctionCall() *ast.FunctionCall {
	name_tk := p.Expect(lexer.Identifier, "")
	var c_params []ast.Expression

	// consume bracket
	p.Expect(lexer.Punctuator, "(")

	needs_comma := false
	for !p.MatchNext(lexer.Punctuator, ")", 0) {
		if needs_comma {
			p.Expect(lexer.Punctuator, ",")
		} else {
			needs_comma = true
		}

		c_params = append(c_params, p.ParseExpression())
	}

	// consume bracket
	p.Expect(lexer.Punctuator, ")")

	return &ast.FunctionCall{
		Function: ast.Identifier{
			Value: name_tk.Value,
		},
		Params: c_params,
	}
}

func (p *Parser) ParseIdentifier() *ast.Identifier {
	id_tk := p.Expect(lexer.Identifier, "")
	return &ast.Identifier{
		Value: id_tk.Value,
	}
}

func (p *Parser) ParseSeparatedExpression() *ast.SeparatedExpression {
	var e_value ast.Expression
	p.Expect(lexer.Punctuator, "(")

	e_value = p.ParseExpression()

	p.Expect(lexer.Punctuator, ")")

	return &ast.SeparatedExpression{
		Value: e_value,
	}
}

func (p *Parser) ParsePrimaryExpression() ast.PrimaryExpression {
	// according to grammar:
	// pexpr : literal
	//       | ident
	//       | call_expr
	//       | block_expr
	//       | sep_expr

	// parse literal
	if p.MatchNext(lexer.Literal, "", 0) {
		return p.ParseLiteral()
	}

	// parse function call
	if p.MatchNext(lexer.Identifier, "", 0) && p.MatchNext(lexer.Punctuator, "(", 1) {
		return p.ParseFunctionCall()
	}

	// parse identifier
	if p.MatchNext(lexer.Identifier, "", 0) {
		return p.ParseIdentifier()
	}

	// parse block expression
	if p.MatchNext(lexer.Punctuator, "{", 0) {
		return p.ParseBlockExpression()
	}

	// parse separated expression
	if p.MatchNext(lexer.Punctuator, "(", 0) {
		return p.ParseSeparatedExpression()
	}

	return &ast.BasePrimaryExpression{}
}

// returns either a primary or a binary expression as defined in the grammar
func (p *Parser) ParseBinaryExpression() ast.Expression {
	var e_left ast.Expression

	e_left = p.ParsePrimaryExpression()

	isCurrentBinop := func() bool {
		_, ok := ast.BinaryOperators[p.Peek().Value]
		return ok && p.MatchNext(lexer.Operator, "", 0)
	}

	for isCurrentBinop() {
		op := ast.BinaryOperators[p.Next().Value]
		e_right := p.ParseExpression()
		e_left = &ast.BinaryExpression{
			Left:     e_left,
			Operator: op,
			Right:    e_right,
		}
	}

	return e_left
}

func (p *Parser) ParseExpression() ast.Expression {
	// according to grammar:
	// expr : bind_expr
	//      | return_expr
	//      | assg_expr
	//      | bin_expr

	// parse bind expression
	if p.MatchNext(lexer.Keyword, "let", 0) {
		return p.ParseBindExpression()
	}

	// parse return expression
	if p.MatchNext(lexer.Keyword, "return", 0) {
		return p.ParseReturnExpression()
	}

	// parse assignment expression
	if p.MatchNext(lexer.Identifier, "", 0) && p.MatchNext(lexer.Operator, "=", 1) {
		return p.ParseAssignmentExpression()
	}

	// parse binary or primary expression
	return p.ParseBinaryExpression()
}

func (p *Parser) ParseBlockExpression() *ast.BlockExpression {
	var e_body []ast.Expression
	var e_return_expression ast.Expression

	p.Expect(lexer.Punctuator, "{")

	has_return_expression := false
	for !p.MatchNext(lexer.Punctuator, "}", 0) {
		expression := p.ParseExpression()

		if !p.MatchNext(lexer.Punctuator, ";", 0) {
			e_return_expression = expression
			has_return_expression = true
		} else {
			if has_return_expression {
				fmt.Printf("Unexpected expression at %s\n", p.Peek().Position.String())
				os.Exit(-1)
			}
			e_body = append(e_body, expression)
			p.Next()
		}
	}

	// consume curly brace
	p.Expect(lexer.Punctuator, "}")

	return &ast.BlockExpression{
		Body:             e_body,
		ReturnExpression: e_return_expression,
	}
}

func (p *Parser) ParseFunctionDeclaration() ast.FunctionDeclaration {
	var f_name ast.Identifier
	var f_type ast.Type
	var f_params []ast.Parameter
	var f_body ast.BlockExpression

	// parse function type
	type_tk := p.Expect(lexer.Identifier, "")
	f_type, ok := ast.Types[type_tk.Value]
	if !ok {
		fmt.Printf("Unknown function return type %s at %s\n", type_tk.Value, type_tk.Position.String())
		os.Exit(-1)
	}

	name_tk := p.Expect(lexer.Identifier, "")
	f_name = ast.Identifier{
		Value: name_tk.Value,
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

		f_params = append(f_params, ast.Parameter{
			Type: param_type,
			Name: ast.Identifier{
				Value: param_name_tk.Value,
			},
		})
	}

	p.Expect(lexer.Punctuator, ")")

	f_body = *p.ParseBlockExpression()

	return ast.FunctionDeclaration{
		Name:       f_name,
		Type:       f_type,
		Parameters: f_params,
		Body:       f_body,
	}
}

func (p *Parser) Parse() ast.Prog {
	res := ast.Prog{}

	for p.head < len(p.tokens) {
		res.Declarations = append(res.Declarations, p.ParseFunctionDeclaration())
	}

	return res
}
