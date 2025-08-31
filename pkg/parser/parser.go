package parser

import (
	"errors"
	"fmt"
	"lang/pkg/ast"
	"lang/pkg/lexer"
)

// ParseError represents a parsing error with position info.
type ParseError struct {
	Message  string
	Position lexer.Position
}

func (e ParseError) Error() string {
	return fmt.Sprintf("%s at %s", e.Message, e.Position.String())
}

// parseError creates a new ParseError.
func parseError(msg string, pos lexer.Position) ParseError {
	return ParseError{
		Message:  msg,
		Position: pos,
	}
}

// Parser holds tokens and the current parsing position.
type Parser struct {
	tokens []lexer.Token
	head   int
}

// NewParser creates a new parser from a token slice.
func NewParser(tokens []lexer.Token) *Parser {
	return &Parser{
		tokens: tokens,
		head:   0,
	}
}

// peek returns the current token without advancing.
func (p *Parser) peek() *lexer.Token {
	if p.head >= len(p.tokens) {
		return nil
	}
	return &p.tokens[p.head]
}

// peekNext returns the token at an offset to the current position.
func (p *Parser) peekNext(offset int) *lexer.Token {
	if p.head+offset >= len(p.tokens) {
		return nil
	}
	return &p.tokens[p.head+offset]
}

// next returns the current token and advances the parser.
// Returns an error if at EOF.
func (p *Parser) next() (*lexer.Token, error) {
	if p.head >= len(p.tokens) {
		if p.head-1 >= 0 {
			prevTk := p.tokens[p.head-1]
			return nil, parseError("unexpected EOF", prevTk.Position)
		}
		return nil, fmt.Errorf("empty file")
	}
	tk := &p.tokens[p.head]
	p.head++
	return tk, nil
}

// matchCurrent checks if the current token matches the kind and value.
func (p *Parser) matchCurrent(kind lexer.TokenKind, value string) bool {
	c := p.peek()
	return c != nil && kind == c.Kind && (value == "" || value == c.Value)
}

// matchNext checks if the token at an offset to the current position matches kind and value.
func (p *Parser) matchNext(kind lexer.TokenKind, value string, offset int) bool {
	c := p.peekNext(offset)
	return c != nil && kind == c.Kind && (value == "" || value == c.Value)
}

// Expect returns the current token and advances the head
// if it matches kind and value, otherwise returns an error.
func (p *Parser) Expect(kind lexer.TokenKind, value string) (*lexer.Token, error) {
	tk, err := p.next()
	if err != nil {
		return nil, err
	}
	if tk.Kind != kind || (value != "" && tk.Value != value) {
		return nil, parseError(
			fmt.Sprintf("expected %s '%s', got %s '%s'", kind.String(), value, tk.Kind.String(), tk.Value),
			tk.Position,
		)
	}
	return tk, nil
}

func (p *Parser) parseIdentifier() (*ast.IdentifierExpression, error) {
	ident_tk, err := p.Expect(lexer.Identifier, "")
	if err != nil {
		return nil, err
	}

	id := &ast.IdentifierExpression{
		Value: ident_tk.Value,
	}
	id.Position = ident_tk.Position

	return id, nil
}

func (p *Parser) parseBindExpression() (*ast.BindExpression, []error) {
	errs := []error{}
	var e_ident *ast.IdentifierExpression
	var e_type_name *ast.IdentifierExpression
	var e_value ast.SimpleExpression
	var expr_start lexer.Position

	first_tk, err := p.Expect(lexer.Keyword, "let")
	if err != nil {
		errs = append(errs, err)
	}
	expr_start = first_tk.Position

	e_ident, err = p.parseIdentifier()
	if err != nil {
		errs = append(errs, err)
	}

	_, err = p.Expect(lexer.Punctuator, ":")
	if err != nil {
		errs = append(errs, err)
	}

	e_type_name, err = p.parseIdentifier()
	if err != nil {
		errs = append(errs, err)
	}

	_, err = p.Expect(lexer.Operator, "=")
	if err != nil {
		errs = append(errs, err)
	}

	e_value, r_errs := p.parseSimpleExpression()
	errs = append(errs, r_errs...)

	r := ast.BindExpression{
		Identifier: e_ident,
		TypeName:   e_type_name,
		Value:      e_value,
	}
	r.Position = expr_start

	return &r, errs
}

func (p *Parser) parseReturnExpression() (*ast.ReturnExpression, []error) {
	var errs []error
	var e_value ast.SimpleExpression
	var expr_start lexer.Position

	ret_tk, err := p.Expect(lexer.Keyword, "return")
	if err != nil {
		errs = append(errs, err)
	}
	if ret_tk != nil {
		expr_start = ret_tk.Position
	}

	e_value, e_errs := p.parseSimpleExpression()
	errs = append(errs, e_errs...)

	r := &ast.ReturnExpression{
		Value: e_value,
	}
	r.Position = expr_start
	return r, errs
}

func (p *Parser) parseAssignmentExpression() (*ast.AssignmentExpression, []error) {
	var errs []error
	var expr_start lexer.Position
	var e_ident *ast.IdentifierExpression
	var e_value ast.SimpleExpression

	e_ident, err := p.parseIdentifier()
	if err != nil {
		errs = append(errs, err)
	}
	if e_ident != nil {
		expr_start = e_ident.Position
	}

	_, err = p.Expect(lexer.Operator, "=")
	if err != nil {
		errs = append(errs, err)
	}

	e_value, e_errs := p.parseSimpleExpression()
	errs = append(errs, e_errs...)

	r := ast.AssignmentExpression{
		Identifier: e_ident,
		Value:      e_value,
	}
	r.Position = expr_start
	return &r, errs
}

func (p *Parser) parseSimpleExpression() (ast.SimpleExpression, []error) {
	var errs []error
	// according to grammar
	// simple_expr ::= primary
	//               | bin_expr
	//               | unary_expr

	// the simplest case is the unary expression
	if p.matchCurrent(lexer.Operator, "") {
		return p.parseUnaryExpression()
	}

	// then we parse a primary expression and distinguish between
	// only a primary expression and a binary expression
	primary, p_errs := p.parsePrimaryExpression()
	errs = append(errs, p_errs...)

	if p.matchCurrent(lexer.Operator, "") {
		operator_tk, err := p.Expect(lexer.Operator, "")
		if err != nil {
			errs = append(errs, err)
		}

		operator, ok := ast.BinaryOperators[operator_tk.Value]
		if !ok {
			errs = append(errs, parseError(fmt.Sprintf("unknown binary operator: %s", operator_tk.Value), operator_tk.Position))
		}

		simple, s_errs := p.parseSimpleExpression()
		errs = append(errs, s_errs...)

		return &ast.BinaryExpression{
			Left:     primary,
			Operator: operator,
			Right:    simple,
		}, errs
	}

	return primary, errs
}

func (p *Parser) parseLiteral() (*ast.LiteralExpression, []error) {
	var errs []error
	var expr_start lexer.Position
	var value string

	literal_tk, err := p.Expect(lexer.Literal, "")
	if err != nil {
		errs = append(errs, err)
	}
	if literal_tk != nil {
		value = literal_tk.Value
		expr_start = literal_tk.Position
	}

	literal := &ast.LiteralExpression{
		Value: value,
	}
	literal.Position = expr_start

	return literal, nil
}

func (p *Parser) parseFunctionCall() (*ast.CallExpression, []error) {
	var errs []error
	var expr_start lexer.Position
	var ident *ast.IdentifierExpression
	var params []ast.SimpleExpression

	ident, err := p.parseIdentifier()
	if err != nil {
		errs = append(errs, err)
	}
	if ident != nil {
		expr_start = ident.Position
	}

	// consume bracket
	_, err = p.Expect(lexer.Punctuator, "(")
	if err != nil {
		errs = append(errs, err)
	}

	needs_comma := false
	for !p.matchNext(lexer.Punctuator, ")", 0) && p.peek() != nil {
		if needs_comma {
			_, err = p.Expect(lexer.Punctuator, ",")
			if err != nil {
				errs = append(errs, err)
			}
		} else {
			needs_comma = true
		}

		p, p_errs := p.parseSimpleExpression()
		errs = append(errs, p_errs...)
		params = append(params, p)
	}

	// consume bracket
	_, err = p.Expect(lexer.Punctuator, ")")
	if err != nil {
		errs = append(errs, err)
	}

	r := ast.CallExpression{
		Identifier: ident,
		Params:     params,
	}
	r.Position = expr_start
	return &r, errs
}

func (p *Parser) parseSeparatedExpression() (*ast.SeparatedExpression, []error) {
	var errs []error
	var expr_start lexer.Position
	var body ast.SimpleExpression

	_, err := p.Expect(lexer.Punctuator, "(")
	if err != nil {
		errs = append(errs, err)
	}

	body, b_errs := p.parseSimpleExpression()
	errs = append(errs, b_errs...)

	_, err = p.Expect(lexer.Punctuator, ")")
	if err != nil {
		errs = append(errs, err)
	}

	e := &ast.SeparatedExpression{
		Body: body,
	}
	e.Position = expr_start

	return e, errs
}

func (p *Parser) parseConditionalExpression() (*ast.ConditionalExpression, []error) {
	var errs []error
	var expr_start lexer.Position
	var condition ast.SimpleExpression
	var if_body *ast.BlockExpression
	var else_body *ast.BlockExpression

	if_tk, err := p.Expect(lexer.Keyword, "if")
	if err != nil {
		errs = append(errs, err)
	}
	if if_tk != nil {
		expr_start = if_tk.Position
	}

	condition, c_errs := p.parseSimpleExpression()
	errs = append(errs, c_errs...)
	if_body, b_errs := p.parseBlockExpression()
	errs = append(errs, b_errs...)

	if p.matchNext(lexer.Keyword, "else", 0) {
		// consume 'else'
		p.next()
		else_b, e_errs := p.parseBlockExpression()
		else_body = else_b
		errs = append(errs, e_errs...)
	}

	e := &ast.ConditionalExpression{
		Condition: condition,
		IfBody:    if_body,
		ElseBody:  else_body,
	}
	e.Position = expr_start

	return e, errs
}

func (p *Parser) parsePrimaryExpression() (ast.PrimaryExpression, []error) {
	// according to grammar:
	// primary ::= literal
	//           | ident
	//           | call_expr
	//           | block_expr
	//           | sep_expr
	//           | con_expr

	// parse literal
	if p.matchNext(lexer.Literal, "", 0) {
		return p.parseLiteral()
	}

	// parse function call
	if p.matchNext(lexer.Identifier, "", 0) && p.matchNext(lexer.Punctuator, "(", 1) {
		return p.parseFunctionCall()
	}

	// parse identifier
	if p.matchNext(lexer.Identifier, "", 0) {
		id, err := p.parseIdentifier()
		return id, []error{err}
	}

	// parse block expression
	if p.matchNext(lexer.Punctuator, "{", 0) {
		return p.parseBlockExpression()
	}

	// parse separated expression
	if p.matchNext(lexer.Punctuator, "(", 0) {
		return p.parseSeparatedExpression()
	}

	// parse conditional expression
	if p.matchNext(lexer.Keyword, "if", 0) {
		return p.parseConditionalExpression()
	}

	var pos lexer.Position
	if p.peek() != nil {
		pos = p.peek().Position
	}

	return nil, []error{parseError("unknown primary expression", pos)}
}

func (p *Parser) parseUnaryExpression() (*ast.UnaryExpression, []error) {
	var errs []error
	var expr_start lexer.Position
	var operator ast.UnaryOperator
	var value ast.PrimaryExpression

	operator_tk, err := p.Expect(lexer.Operator, "")
	if err != nil {
		errs = append(errs, err)
	}
	if operator_tk != nil {
		expr_start = operator_tk.Position
		op, ok := ast.UnaryOperators[operator_tk.Value]
		if !ok {
			errs = append(errs, parseError(fmt.Sprintf("unknown unary operator: %s", operator_tk.Value), operator_tk.Position))
		} else {
			operator = op
		}
	}

	value, v_errs := p.parsePrimaryExpression()
	errs = append(errs, v_errs...)

	r := ast.UnaryExpression{
		Operator: operator,
		Value:    value,
	}
	r.Position = expr_start
	return &r, errs
}

func (p *Parser) parseExpression() (ast.Expression, []error) {
	// according to grammar:
	// expr : bind_expr
	//      | return_expr
	//      | assg_expr
	//      | simple_expr

	// parse bind expression
	if p.matchNext(lexer.Keyword, "let", 0) {
		return p.parseBindExpression()
	}

	// parse return expression
	if p.matchNext(lexer.Keyword, "return", 0) {
		return p.parseReturnExpression()
	}

	// parse assignment expression
	if p.matchNext(lexer.Identifier, "", 0) && p.matchNext(lexer.Operator, "=", 1) {
		return p.parseAssignmentExpression()
	}

	// parse simple expression
	return p.parseSimpleExpression()
}

// parseBlockExpression parses a block expression, returning it and all the
// errors it encountered along the way.
func (p *Parser) parseBlockExpression() (*ast.BlockExpression, []error) {
	errs := []error{}
	var expr_start lexer.Position
	if tk := p.peek(); tk != nil {
		expr_start = tk.Position
	}
	var e_body []ast.Expression
	var e_return_expression ast.Expression

	// expect opening curly brace, if not found can't parse expression at all
	_, err := p.Expect(lexer.Punctuator, "{")
	if err != nil {
		errs = append(errs, err)
		return nil, errs
	}

	has_return_expression := false
	for p.peek() != nil && !p.matchNext(lexer.Punctuator, "}", 0) {
		expression, e_errs := p.parseExpression()
		errs = append(errs, e_errs...)

		if !p.matchNext(lexer.Punctuator, ";", 0) {
			if has_return_expression {
				errs = append(errs, parseError("unexpected expression, block expression already has a return expression", expression.GetPosition()))
				continue
			}
			e_return_expression = expression
			has_return_expression = true
		} else {
			e_body = append(e_body, expression)
			_, err = p.Expect(lexer.Punctuator, ";")
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	// consume curly brace
	_, err = p.Expect(lexer.Punctuator, "}")
	if err != nil {
		errs = append(errs, err)
	}

	r := ast.BlockExpression{
		Body:           e_body,
		ImplicitReturn: e_return_expression,
	}
	r.Position = expr_start

	return &r, errs
}

// ParseFunctionDeclaration parses a function declaration along with its body
// starting at the position of the Parser head. Any errors encountered along
// the way will be returned.
func (p *Parser) ParseFunctionDeclaration() (ast.FunctionDeclaration, []error) {
	errs := []error{}
	var f_ident *ast.IdentifierExpression
	var f_type_name *ast.IdentifierExpression
	var f_params []ast.ParameterDefinition
	var f_body *ast.BlockExpression

	// parse function type
	f_type_name, err := p.parseIdentifier()
	if err != nil {
		errs = append(errs, err)
	}

	// parse function name
	f_ident, err = p.parseIdentifier()
	if err != nil {
		errs = append(errs, err)
	}

	// expect opening parenthesis
	_, err = p.Expect(lexer.Punctuator, "(")
	if err != nil {
		errs = append(errs, err)
	}

	param_names := map[string]bool{}
	requires_comma := false

	// parse parameters
	for p.peek() != nil && !p.matchCurrent(lexer.Punctuator, ")") {
		if requires_comma {
			_, err = p.Expect(lexer.Punctuator, ",")
			if err != nil {
				errs = append(errs, err)
			}
		} else {
			requires_comma = true
		}
		// parse parameter type
		param_type_name, err := p.parseIdentifier()
		if err != nil {
			errs = append(errs, err)
		}

		// parse parameter name
		param_name, err := p.parseIdentifier()
		if err != nil {
			errs = append(errs, err)
		}
		if param_names[param_name.Value] {
			errs = append(errs, parseError(fmt.Sprintf("parameter %s already defined", param_name.Value), param_name.Position))
		}

		param_names[param_name.Value] = true

		f_params = append(f_params, ast.ParameterDefinition{
			Name:     param_name,
			TypeName: param_type_name,
		})
	}

	// expect closing parenthesis
	_, err = p.Expect(lexer.Punctuator, ")")
	if err != nil {
		errs = append(errs, err)
	}

	// parse function body
	f_body, body_parse_errors := p.parseBlockExpression()
	errs = append(errs, body_parse_errors...)

	return ast.FunctionDeclaration{
		Identifier: f_ident,
		TypeName:   f_type_name,
		Parameters: f_params,
		Body:       f_body,
	}, errs
}

func (p *Parser) Parse() (ast.Program, []error) {
	res := ast.Program{}
	if len(p.tokens) == 0 {
		return res, []error{errors.New("empty file")}
	}

	errors := []error{}
	for p.head < len(p.tokens) {
		d, errs := p.ParseFunctionDeclaration()
		res.Declarations = append(res.Declarations, d)
		errors = append(errors, errs...)
	}

	return res, errors
}
