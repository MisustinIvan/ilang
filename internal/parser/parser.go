/*
Implements a recursive-descent parser for the language.

The parser consumes a slice of tokens produced by the lexer and produces
an AST representation of the source code.
*/
package parser

import (
	"fmt"

	"github.com/MisustinIvan/ilang/internal/ast"
	"github.com/MisustinIvan/ilang/internal/lexer"
)

// ParseError represents a parsing error with position info.
type ParseError struct {
	Message  string
	Position lexer.Position
}

func (e ParseError) Error() string {
	return fmt.Sprintf("%s ParseError: %s", e.Position.String(), e.Message)
}

// parseError creates a new ParseError.
func parseError(msg string, pos lexer.Position) ParseError {
	return ParseError{
		Message:  msg,
		Position: pos,
	}
}

// Parser holds tokens and the curent parsing position.
type Parser struct {
	tokens     []lexer.Token
	head       int
	tokens_len int
}

// Creates the Parser from a Token slice.
func New(tokens []lexer.Token) *Parser {
	return &Parser{
		tokens:     tokens,
		tokens_len: len(tokens),
	}
}

// headInBounds returns whether the parser head is in bounds, we don't need
// to check the the negative indexes as we should never decrement the head.
func (p *Parser) headInBounds() bool {
	return p.head < p.tokens_len
}

// headInBounds returns whether the offset parser head is in bounds.
func (p *Parser) offsetHeadInBounds(offset int) bool {
	return p.head+offset >= 0 && p.head+offset < p.tokens_len
}

// peek returns the current token without advancing.
func (p *Parser) peek() *lexer.Token {
	if !p.headInBounds() {
		return nil
	}
	return &p.tokens[p.head]
}

// peekNext returns the token at an offset to the current position.
func (p *Parser) peekNext(offset int) *lexer.Token {
	if !p.offsetHeadInBounds(offset) {
		return nil
	}
	return &p.tokens[p.head+offset]
}

// next returns the current token and advances the parser.
// Returns an error if at EOF.
func (p *Parser) next() (*lexer.Token, error) {
	if !p.headInBounds() {
		if p.offsetHeadInBounds(-1) {
			previousToken := p.tokens[p.head-1]
			return nil, parseError("unexpected EOF", previousToken.Position)
		}
		return nil, fmt.Errorf("empty file")
	}
	token := &p.tokens[p.head]
	p.head++
	return token, nil
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
		var msg string
		if value == "" {
			msg = fmt.Sprintf("expected %s, got %s '%s'", kind.String(), tk.Kind.String(), tk.Value)
		} else {
			msg = fmt.Sprintf("expected %s '%s', got %s '%s'", kind.String(), value, tk.Kind.String(), tk.Value)
		}
		return nil, parseError(
			msg,
			tk.Position,
		)
	}
	return tk, nil
}

// ParseExternalDeclaration parses an external function declaration according
// to the grammar:
//
// external_declaration ::= "extrn" type identifier "(" [ function_parameter { "," function_parameter } ] ")"
func (p *Parser) ParseExternalDeclaration() (*ast.ExternalDeclaration, error) {
	var Type *ast.Type
	var Identifier *ast.Identifier
	var Parameters []ast.Parameter

	// expect "extrn" keyword
	_, err := p.Expect(lexer.Keyword, lexer.KeywordExtrn)
	if err != nil {
		return nil, err
	}

	// parse type
	Type, err = p.ParseType()
	if err != nil {
		return nil, err
	}

	// parse identifier
	Identifier, err = p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	// expect opening paren
	_, err = p.Expect(lexer.Punctuator, "(")
	if err != nil {
		return nil, err
	}

	// parse parameters
	for !p.matchCurrent(lexer.Punctuator, ")") {
		parameter, err := p.ParseFunctionParameter()
		if err != nil {
			return nil, err
		}
		Parameters = append(Parameters, *parameter)
	}

	// consume closing paren
	_, err = p.Expect(lexer.Punctuator, ")")
	if err != nil {
		return nil, err
	}

	decl := &ast.ExternalDeclaration{
		Type:       *Type,
		Identifier: Identifier,
		Params:     Parameters,
	}

	return decl, nil
}

// ParseDeclaration parses a function declaration according to the grammar:
//
// declaration          ::= type identifier "(" [ function_parameter { "," function_parameter } ] ")" block
func (p *Parser) ParseDeclaration() (*ast.Declaration, error) {
	var Type *ast.Type
	var Identifier *ast.Identifier
	var Parameters []ast.Parameter
	var Body *ast.Block

	// parse type
	Type, err := p.ParseType()
	if err != nil {
		return nil, err
	}

	// parse identifier
	Identifier, err = p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	// expect opening paren
	_, err = p.Expect(lexer.Punctuator, "(")
	if err != nil {
		return nil, err
	}

	// parse parameters
	for !p.matchCurrent(lexer.Punctuator, ")") {
		parameter, err := p.ParseFunctionParameter()
		if err != nil {
			return nil, err
		}
		Parameters = append(Parameters, *parameter)
	}

	// expect closing paren
	_, err = p.Expect(lexer.Punctuator, ")")
	if err != nil {
		return nil, err
	}

	// parse body
	Body, err = p.ParseBlock()
	if err != nil {
		return nil, err
	}

	return &ast.Declaration{
		Type:       *Type,
		Identifier: Identifier,
		Params:     Parameters,
		Body:       *Body,
	}, nil
}

// ParseBlock parses a block expesssion according to the grammar:
//
// block                ::= "{" { expression ";" } [ expression ] "}"
func (p *Parser) ParseBlock() (*ast.Block, error) {
	var Body []ast.Expression
	var ImplicitReturn ast.Expression

	start_tk, err := p.Expect(lexer.Punctuator, "{")
	if err != nil {
		return nil, err
	}

	for !p.matchCurrent(lexer.Punctuator, "}") {
		expr, err := p.ParseExpression()
		if err != nil {
			return nil, err
		}

		if p.matchCurrent(lexer.Punctuator, ";") {
			// consume the semicolon
			p.next()
			Body = append(Body, expr)
		} else if ImplicitReturn == nil {
			ImplicitReturn = expr
		} else {
			return nil, fmt.Errorf("unexpected expression at %v", expr.GetPosition())
		}
	}

	_, err = p.Expect(lexer.Punctuator, "}")
	if err != nil {
		return nil, err
	}

	block := &ast.Block{
		Body:           Body,
		ImplicitReturn: ImplicitReturn,
	}
	block.SetPosition(start_tk.Position)

	return block, nil
}

// ParseExpression parses an expression according to the grammar:
//
//	expression           ::= return
//	                       | bind
//	                       | assignment
//	                       | value
func (p *Parser) ParseExpression() (ast.Expression, error) {
	switch {
	case p.matchCurrent(lexer.Keyword, lexer.KeywordReturn):
		return p.ParseReturn()
	case p.matchCurrent(lexer.Keyword, lexer.KeywordLet):
		return p.ParseBind()
	case p.matchCurrent(lexer.Identifier, "") && p.matchNext(lexer.Operator, "=", 1):
		return p.ParseAssignment()
	default:
		return p.ParseValue()
	}
}

// ParseReturn parses a return expression according to the grammar:
//
// return               ::= "return" value
func (p *Parser) ParseReturn() (*ast.Return, error) {
	var Value ast.Value

	let_tk, err := p.Expect(lexer.Keyword, lexer.KeywordReturn)
	if err != nil {
		return nil, err
	}

	Value, err = p.ParseValue()
	if err != nil {
		return nil, err
	}

	ret := &ast.Return{
		Value: Value,
	}
	ret.SetPosition(let_tk.Position)

	return ret, nil
}

// ParseBind parses a bind expression according to the grammar:
//
// bind                 ::= "let" identifier ":" type "=" value
func (p *Parser) ParseBind() (*ast.Bind, error) {
	var Identifier *ast.Identifier
	var Type *ast.Type
	var Value ast.Value

	let_tk, err := p.Expect(lexer.Keyword, lexer.KeywordLet)
	if err != nil {
		return nil, err
	}

	Identifier, err = p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	_, err = p.Expect(lexer.Punctuator, ":")
	if err != nil {
		return nil, err
	}

	Type, err = p.ParseType()
	if err != nil {
		return nil, err
	}

	_, err = p.Expect(lexer.Operator, "=")
	if err != nil {
		return nil, err
	}

	Value, err = p.ParseValue()
	if err != nil {
		return nil, err
	}

	bind := &ast.Bind{
		Identifier: Identifier,
		Type:       *Type,
		Value:      Value,
	}
	bind.SetPosition(let_tk.Position)

	return bind, nil
}

// ParseAssignment parses an assignment expression according to the grammar:
//
// assignment           ::= identifier "=" value
func (p *Parser) ParseAssignment() (*ast.Assignment, error) {
	var Identifier *ast.Identifier
	var Value ast.Value

	Identifier, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	_, err = p.Expect(lexer.Operator, "=")
	if err != nil {
		return nil, err
	}

	Value, err = p.ParseValue()
	if err != nil {
		return nil, err
	}

	assignment := &ast.Assignment{
		Identifier: Identifier,
		Value:      Value,
	}
	assignment.SetPosition(Identifier.Position)

	return assignment, nil
}

// ParseValue parses a value expression according to the grammar:
//
//	value                ::= primary
//	                       | binary
//	                       | unary
func (p *Parser) ParseValue() (ast.Value, error) {
	if p.matchCurrent(lexer.Operator, "") {
		return p.ParseUnary()
	}

	primary, err := p.ParsePrimary()
	if err != nil {
		return nil, err
	}

	if p.matchCurrent(lexer.Operator, "") {
		operator_tk, err := p.Expect(lexer.Operator, "")
		if err != nil {
			return nil, err
		}

		operator, ok := ast.BinaryOperatorTokens[operator_tk.Value]
		if !ok {
			return nil, fmt.Errorf("unexpected binary operator %v", operator_tk)
		}

		right, err := p.ParseValue()
		if err != nil {
			return nil, err
		}

		binary := &ast.Binary{
			Left:     primary,
			Operator: operator,
			Right:    right,
		}
		binary.SetPosition(primary.GetPosition())

		return binary, nil
	}

	return primary, nil
}

// ParsePrimary parses a primary expression according to the grammar:
//
//	primary              ::= literal
//	                       | identifier
//	                       | call
//	                       | separated
//	                       | block
//	                       | condition
func (p *Parser) ParsePrimary() (ast.Primary, error) {
	switch {
	case p.matchCurrent(lexer.Literal, ""):
		return p.ParseLiteral()
	case p.matchCurrent(lexer.Identifier, "") && p.matchNext(lexer.Punctuator, "(", 1):
		return p.ParseCall()
	case p.matchCurrent(lexer.Identifier, ""):
		return p.ParseIdentifier()
	case p.matchCurrent(lexer.Punctuator, "("):
		return p.ParseSeparated()
	case p.matchCurrent(lexer.Punctuator, "{"):
		return p.ParseBlock()
	case p.matchCurrent(lexer.Keyword, lexer.KeywordIf):
		return p.ParseCondition()
	default:
		return nil, fmt.Errorf("unexpected primary expression")
	}
}

// ParseLiteral parses a literal according to the grammar:
//
// literal              ::= "*."
func (p *Parser) ParseLiteral() (*ast.Literal, error) {
	literal_tk, err := p.Expect(lexer.Literal, "")
	if err != nil {
		return nil, err
	}

	literal := &ast.Literal{
		Value: literal_tk.Value,
	}
	literal.SetPosition(literal_tk.Position)

	return literal, nil
}

// ParseCall parses a call expression according to the grammar:
//
// call                 ::= identifier "(" [ value { "," value } ] ")"
func (p *Parser) ParseCall() (*ast.Call, error) {
	var Identifier *ast.Identifier
	var Arguments []ast.Value

	Identifier, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	_, err = p.Expect(lexer.Punctuator, "(")
	if err != nil {
		return nil, err
	}

	for !p.matchCurrent(lexer.Punctuator, ")") {
		value, err := p.ParseValue()
		if err != nil {
			return nil, err
		}
		Arguments = append(Arguments, value)
	}

	_, err = p.Expect(lexer.Punctuator, ")")
	if err != nil {
		return nil, err
	}

	call := &ast.Call{
		Identifier: Identifier,
		Arguments:  Arguments,
	}
	call.SetPosition(Identifier.Position)

	return call, nil
}

// ParseSeparated parses a separated expression according to the grammar:
//
// separated            ::= "(" value ")"
func (p *Parser) ParseSeparated() (*ast.Separated, error) {
	var Value ast.Value

	start_tk, err := p.Expect(lexer.Punctuator, "(")
	if err != nil {
		return nil, err
	}

	Value, err = p.ParseValue()
	if err != nil {
		return nil, err
	}

	_, err = p.Expect(lexer.Punctuator, ")")
	if err != nil {
		return nil, err
	}

	separated := &ast.Separated{
		Value: Value,
	}
	separated.SetPosition(start_tk.Position)

	return separated, nil
}

// ParseCondition parses a condition expression according to the grammar:
//
//	condition            ::= "if" value value
//	                         [ "else" value ]
func (p *Parser) ParseCondition() (*ast.Condition, error) {
	var Condition ast.Value
	var Body ast.Value
	var Else ast.Value

	if_tk, err := p.Expect(lexer.Keyword, lexer.KeywordIf)
	if err != nil {
		return nil, err
	}

	Condition, err = p.ParseValue()
	if err != nil {
		return nil, err
	}

	Body, err = p.ParseValue()
	if err != nil {
		return nil, err
	}

	if p.matchCurrent(lexer.Keyword, lexer.KeywordElse) {
		_, err = p.Expect(lexer.Keyword, lexer.KeywordElse)
		if err != nil {
			return nil, err
		}

		Else, err = p.ParseValue()
		if err != nil {
			return nil, err
		}
	}

	condition := &ast.Condition{
		Condition: Condition,
		Body:      Body,
		Else:      Else,
	}
	condition.SetPosition(if_tk.Position)

	return condition, nil
}

// ParseUnary parses a unary expression according to the grammar:
//
// unary                ::= unary_operator primary
func (p *Parser) ParseUnary() (*ast.Unary, error) {
	var Operator ast.UnaryOperator
	var Value ast.Primary

	operator_tk, err := p.Expect(lexer.Operator, "")
	if err != nil {
		return nil, err
	}

	Operator, ok := ast.UnaryOperatorTokens[operator_tk.Value]
	if !ok {
		return nil, fmt.Errorf("unexpected unary operator %v", operator_tk)
	}

	Value, err = p.ParsePrimary()
	if err != nil {
		return nil, err
	}

	unary := &ast.Unary{
		Operator: Operator,
		Value:    Value,
	}
	unary.SetPosition(operator_tk.Position)

	return unary, nil
}

// ParseFunctionParameter parses a function parameter according to the grammar:
//
// function_parameter   ::= type identifier
func (p *Parser) ParseFunctionParameter() (*ast.Parameter, error) {
	var Type *ast.Type
	var Identifier *ast.Identifier

	Type, err := p.ParseType()
	if err != nil {
		return nil, err
	}

	Identifier, err = p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	return &ast.Parameter{
		Type:       *Type,
		Identifier: Identifier,
	}, nil
}

// ParseType parses a type according to the grammar:
//
// type                 ::= "int" | "bool" | "float" | "string" | "unit"
func (p *Parser) ParseType() (*ast.Type, error) {
	token, err := p.Expect(lexer.Identifier, "")
	if err != nil {
		return nil, err
	}

	var Type ast.Type = ast.Undefined

	switch token.Value {
	case "int":
		Type = ast.Int
	case "bool":
		Type = ast.Bool
	case "float":
		Type = ast.Float
	case "string":
		Type = ast.String
	case "unit":
		Type = ast.Unit
	}

	return &Type, nil
}

// ParseIdentifier parses an identifier according to the grammar:
//
// identifier           ::= "*."
func (p *Parser) ParseIdentifier() (*ast.Identifier, error) {
	token, err := p.Expect(lexer.Identifier, "")
	if err != nil {
		return nil, err
	}

	identifier := &ast.Identifier{Name: token.Value}
	identifier.SetPosition(token.Position)

	return identifier, err
}

// Parses the tokens, returning the *ast.Program.
func (p *Parser) Parse() (*ast.Program, error) {
	program := &ast.Program{}

	for p.headInBounds() {
		if p.matchCurrent(lexer.Keyword, lexer.KeywordExtrn) {
			decl, err := p.ParseExternalDeclaration()
			if err != nil {
				return nil, err
			}
			program.ExternalDeclarations = append(program.ExternalDeclarations, decl)
		} else {
			decl, err := p.ParseDeclaration()
			if err != nil {
				return nil, err
			}
			program.Declarations = append(program.Declarations, decl)
		}
	}

	return program, nil
}
