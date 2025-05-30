package parser

import (
	"fmt"
	"lang/pkg/ast"
	"lang/pkg/lexer"
	"os"
)

type ParseError struct {
	Message  string
	Position lexer.TokenPosition
}

func (e ParseError) Error() string {
	return fmt.Sprintf("%s at %s", e.Message, e.Position.String())
}

func parseError(msg string, pos lexer.TokenPosition) ParseError {
	return ParseError{
		Message:  msg,
		Position: pos,
	}
}

type Parser struct {
	tokens []lexer.Token
	head   int
	errors []ParseError
}

func (p *Parser) Report() {
	for _, e := range p.errors {
		fmt.Printf("%s\n", e.Error())
	}
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

func (p *Parser) ParseBindExpression() (*ast.BindExpression, []ParseError) {
	errs := []ParseError{}
	var e_left ast.Identifier
	var e_type ast.Type
	var e_right ast.Expression

	expr_start := p.Peek().Position
	p.Expect(lexer.Keyword, "let")

	left_tk := p.Expect(lexer.Identifier, "")
	e_left = ast.Identifier{
		Value: left_tk.Value,
	}

	p.Expect(lexer.Punctuator, ":")

	type_tk := p.Expect(lexer.Identifier, "")
	e_type, ok := ast.ParseType(type_tk.Value)
	if !ok {
		errs = append(errs, parseError(fmt.Sprintf("unknown type %s", type_tk.Value), type_tk.Position))
	}

	p.Expect(lexer.Operator, "=")

	e_right, r_errs := p.ParseExpression()
	errs = append(errs, r_errs...)

	r := ast.BindExpression{
		Left:  e_left,
		Right: e_right,
	}
	r.Type = e_type
	r.TokenPosition = expr_start

	return &r, errs
}

func (p *Parser) ParseReturnExpression() (*ast.ReturnExpression, []ParseError) {
	var e_value ast.Expression
	expr_start := p.Peek().Position

	p.Expect(lexer.Keyword, "return")

	e_value, errs := p.ParseExpression()

	r := ast.ReturnExpression{Value: e_value}
	r.TokenPosition = expr_start
	return &r, errs
}

func (p *Parser) ParseAssignmentExpression() (*ast.AssignmentExpression, []ParseError) {
	expr_start := p.Peek().Position
	var e_left ast.Identifier
	var e_right ast.Expression

	left_tk := p.Expect(lexer.Identifier, "")
	e_left = ast.Identifier{
		Value: left_tk.Value,
	}

	p.Expect(lexer.Operator, "=")

	e_right, errs := p.ParseExpression()

	r := ast.AssignmentExpression{
		Left:  e_left,
		Right: e_right,
	}
	r.TokenPosition = expr_start
	return &r, errs
}

func (p *Parser) ParseLiteral() (*ast.Literal, []ParseError) {
	expr_start := p.Peek().Position
	l_value := p.Expect(lexer.Literal, "")
	l_type := ast.LiteralType(l_value.Value)

	r := ast.Literal{Value: l_value.Value}
	r.Type = l_type
	r.TokenPosition = expr_start
	return &r, nil
}

func (p *Parser) ParseFunctionCall() (*ast.FunctionCall, []ParseError) {
	expr_start := p.Peek().Position
	errs := []ParseError{}
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

		p, p_errs := p.ParseExpression()
		errs = append(errs, p_errs...)
		c_params = append(c_params, p)
	}

	// consume bracket
	p.Expect(lexer.Punctuator, ")")

	r := ast.FunctionCall{
		Function: ast.Identifier{
			Value: name_tk.Value,
		},
		Params: c_params,
	}

	r.TokenPosition = expr_start
	return &r, errs
}

func (p *Parser) ParseIdentifier() (*ast.Identifier, []ParseError) {
	expr_start := p.Peek().Position
	id_tk := p.Expect(lexer.Identifier, "")
	r := ast.Identifier{
		Value: id_tk.Value,
	}
	r.TokenPosition = expr_start
	return &r, nil
}

func (p *Parser) ParseSeparatedExpression() (*ast.SeparatedExpression, []ParseError) {
	expr_start := p.Peek().Position
	var e_value ast.Expression
	p.Expect(lexer.Punctuator, "(")

	e_value, errs := p.ParseExpression()

	p.Expect(lexer.Punctuator, ")")

	r := ast.SeparatedExpression{
		Value: e_value,
	}
	r.TokenPosition = expr_start
	return &r, errs
}

func (p *Parser) ParseConditionalExpression() (*ast.ConditionalExpression, []ParseError) {
	expr_start := p.Peek().Position
	errs := []ParseError{}
	var e_condition ast.Expression
	var e_if_body *ast.BlockExpression
	var e_else_body *ast.BlockExpression = &ast.BlockExpression{
		Body:                     []ast.Expression{},
		ImplicitReturnExpression: nil,
	}

	p.Expect(lexer.Keyword, "if")

	e_condition, c_errs := p.ParseExpression()
	errs = append(errs, c_errs...)
	e_if_body, b_errs := p.ParseBlockExpression()
	errs = append(errs, b_errs...)

	if p.MatchNext(lexer.Keyword, "else", 0) {
		// consume 'else'
		p.Next()
		e_else_b, e_errs := p.ParseBlockExpression()
		e_else_body = e_else_b
		errs = append(errs, e_errs...)
	}

	r := ast.ConditionalExpression{
		Condition: e_condition,
		IfBody:    *e_if_body,
		ElseBody:  *e_else_body,
	}
	r.TokenPosition = expr_start
	return &r, errs
}

func (p *Parser) ParsePrimaryExpression() (ast.PrimaryExpression, []ParseError) {
	// according to grammar:
	// pexpr : literal
	//       | ident
	//       | call_expr
	//       | block_expr
	//       | sep_expr
	//       | con_expr

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

	// parse conditional expression
	if p.MatchNext(lexer.Keyword, "if", 0) {
		return p.ParseConditionalExpression()
	}

	return nil, nil
}

// returns either a primary or a binary expression as defined in the grammar
func (p *Parser) ParseBinaryExpression() (ast.Expression, []ParseError) {
	expr_start := p.Peek().Position
	var e_left ast.Expression

	e_left, errs := p.ParsePrimaryExpression()

	isCurrentBinop := func() bool {
		_, ok := ast.BinaryOperators[p.Peek().Value]
		return ok && p.MatchNext(lexer.Operator, "", 0)
	}

	for isCurrentBinop() {
		op := ast.BinaryOperators[p.Next().Value]
		e_right, r_errs := p.ParseExpression()
		errs = append(errs, r_errs...)
		e_left = &ast.BinaryExpression{
			Left:     e_left,
			Operator: op,
			Right:    e_right,
		}
	}

	if e, ok := e_left.(*ast.BinaryExpression); ok {
		e.TokenPosition = expr_start
	}

	return e_left, errs
}

func (p *Parser) ParseForExpression() (*ast.ForExpression, []ParseError) {
	expr_start := p.Peek().Position
	p.Expect(lexer.Keyword, "for")

	condition, errs := p.ParseExpression()
	body, b_errs := p.ParseBlockExpression()
	errs = append(errs, b_errs...)

	r := ast.ForExpression{
		Condition: condition,
		Body:      body,
	}
	r.TokenPosition = expr_start
	return &r, errs
}

func (p *Parser) ParseBreakExpression() (*ast.BreakExpression, []ParseError) {
	expr_start := p.Peek().Position
	p.Expect(lexer.Keyword, "break")
	r := ast.BreakExpression{}
	r.TokenPosition = expr_start
	return &r, nil
}

func (p *Parser) ParseUnaryExpression() (*ast.UnaryExpression, []ParseError) {
	expr_start := p.Peek().Position
	op_tk := p.Expect(lexer.Operator, "")
	op, ok := ast.UnaryOperators[op_tk.Value]
	val, errs := p.ParseExpression()
	if !ok {
		errs = append(errs, parseError(fmt.Sprintf("unknown unary operator %s", op_tk.Value), op_tk.Position))
	}

	r := ast.UnaryExpression{
		Operator: op,
		Value:    val,
	}
	r.TokenPosition = expr_start
	return &r, errs
}

func (p *Parser) ParseExpression() (ast.Expression, []ParseError) {
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

	// parse for expression
	if p.MatchNext(lexer.Keyword, "for", 0) {
		return p.ParseForExpression()
	}

	if p.MatchNext(lexer.Keyword, "break", 0) {
		return p.ParseBreakExpression()
	}

	// parse assignment expression
	if p.MatchNext(lexer.Identifier, "", 0) && p.MatchNext(lexer.Operator, "=", 1) {
		return p.ParseAssignmentExpression()
	}

	// parse unary expression
	if p.MatchNext(lexer.Operator, "", 0) {
		if _, ok := ast.UnaryOperators[p.Peek().Value]; ok {
			return p.ParseUnaryExpression()
		}
	}

	// parse binary or primary expression
	return p.ParseBinaryExpression()
}

func (p *Parser) ParseBlockExpression() (*ast.BlockExpression, []ParseError) {
	errs := []ParseError{}
	expr_start := p.Peek().Position
	var e_body []ast.Expression
	var e_return_expression ast.Expression

	p.Expect(lexer.Punctuator, "{")

	has_return_expression := false
	for !p.MatchNext(lexer.Punctuator, "}", 0) {
		expr_start := p.Peek().Position
		expression, e_errs := p.ParseExpression()
		errs = append(errs, e_errs...)

		if !p.MatchNext(lexer.Punctuator, ";", 0) {
			if has_return_expression {
				errs = append(errs, parseError("unexpected expression, block expression already has a return expression", expr_start))
				continue
			}
			e_return_expression = expression
			has_return_expression = true
		} else {
			e_body = append(e_body, expression)
			p.Next()
		}
	}

	// consume curly brace
	p.Expect(lexer.Punctuator, "}")

	r := ast.BlockExpression{
		Body:                     e_body,
		ImplicitReturnExpression: e_return_expression,
	}
	r.TokenPosition = expr_start

	return &r, errs
}

func (p *Parser) ParseFunctionDeclaration() (ast.FunctionDeclaration, []ParseError) {
	errs := []ParseError{}
	var f_name ast.Identifier
	var f_type ast.Type
	var f_params []ast.Parameter
	var f_body *ast.BlockExpression

	// parse function type
	type_tk := p.Expect(lexer.Identifier, "")
	f_type, ok := ast.ParseType(type_tk.Value)
	if !ok {
		errs = append(errs, parseError(fmt.Sprintf("unknown function return type %s", type_tk.Value), type_tk.Position))
	}

	start_loc := type_tk.Position

	name_tk := p.Expect(lexer.Identifier, "")
	f_name = ast.Identifier{
		Value: name_tk.Value,
	}
	f_name.TokenPosition = name_tk.Position

	p.Expect(lexer.Punctuator, "(")

	param_names := map[string]bool{}

	requires_comma := false

	for !p.MatchCurrent(lexer.Punctuator, ")") {
		if requires_comma {
			if !p.MatchCurrent(lexer.Punctuator, ",") {
				errs = append(errs, parseError(fmt.Sprintf("expected comma at %s, got %s %s", p.Peek().Position.String(), p.Peek().Kind.String(), p.Peek().Value), p.Peek().Position))
			} else {
				// consume the comma
				p.Next()
			}
		} else {
			requires_comma = true
		}
		param_type_tk := p.Expect(lexer.Identifier, "")
		param_type, ok := ast.ParseType(param_type_tk.Value)
		if !ok {
			errs = append(errs, parseError(fmt.Sprintf("unknown parameter type %s", param_type_tk.Value), param_type_tk.Position))
		}

		param_name_tk := p.Expect(lexer.Identifier, "")

		if param_names[param_name_tk.Value] {
			errs = append(errs, parseError(fmt.Sprintf("parameter %s already declared", param_name_tk.Value), param_name_tk.Position))
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

	f_body, b_errs := p.ParseBlockExpression()
	errs = append(errs, b_errs...)

	return ast.FunctionDeclaration{
		Name:       f_name,
		Type:       f_type,
		Parameters: f_params,
		Body:       *f_body,
		Position:   start_loc,
	}, errs
}

func (p *Parser) Parse() (ast.Prog, bool) {
	res := ast.Prog{}

	for p.head < len(p.tokens) {
		d, errs := p.ParseFunctionDeclaration()
		res.Declarations = append(res.Declarations, d)
		p.errors = append(p.errors, errs...)
	}

	return res, len(p.errors) == 0
}
