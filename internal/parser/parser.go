/*
Implements a recursive-descent parser for the language.

The parser consumes a slice of tokens produced by the lexer and produces
an AST representation of the source code.
*/
package parser

import (
	"fmt"
	"strconv"

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
// external_declaration ::= "extrn" basic_type identifier "(" [ function_argument { "," function_argument } ["," "..."] ] | "..." ")"
func (p *Parser) ParseExternalDeclaration() (*ast.ExternalDeclaration, error) {
	var Type ast.Type
	var Identifier *ast.Identifier
	var Arguments []ast.Argument
	var Variadic = false

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

	// parse arguments
	if p.matchCurrent(lexer.Punctuator, ")") {
		// No arguments
	} else if p.matchCurrent(lexer.Punctuator, "...") {
		Variadic = true
		_, err = p.Expect(lexer.Punctuator, "...")
		if err != nil {
			return nil, err
		}
	} else {
		argument, err := p.ParseFunctionArgument()
		if err != nil {
			return nil, err
		}
		Arguments = append(Arguments, *argument)

		for p.matchCurrent(lexer.Punctuator, ",") {
			_, err = p.Expect(lexer.Punctuator, ",")
			if err != nil {
				return nil, err
			}

			if p.matchCurrent(lexer.Punctuator, "...") {
				Variadic = true
				_, err = p.Expect(lexer.Punctuator, "...")
				if err != nil {
					return nil, err
				}
				break
			}

			argument, err := p.ParseFunctionArgument()
			if err != nil {
				return nil, err
			}
			Arguments = append(Arguments, *argument)
		}
	}

	// consume closing paren
	_, err = p.Expect(lexer.Punctuator, ")")
	if err != nil {
		return nil, err
	}

	decl := &ast.ExternalDeclaration{
		Type:       Type,
		Identifier: Identifier,
		Args:       Arguments,
		Variadic:   Variadic,
	}

	return decl, nil
}

// ParseDeclaration parses a function declaration according to the grammar:
//
// declaration          ::= basic_type identifier "(" [ function_argument { "," function_argument } ] ")" block
func (p *Parser) ParseDeclaration() (*ast.Declaration, error) {
	var Type *ast.BasicType
	var Identifier *ast.Identifier
	var Arguments []ast.Argument
	var Body *ast.Block

	// parse type
	Type, err := p.ParseBasicType()
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

	// parse arguments
	if p.matchCurrent(lexer.Punctuator, ")") {
		// No arguments
	} else {
		argument, err := p.ParseFunctionArgument()
		if err != nil {
			return nil, err
		}
		Arguments = append(Arguments, *argument)

		for p.matchCurrent(lexer.Punctuator, ",") {
			_, err = p.Expect(lexer.Punctuator, ",")
			if err != nil {
				return nil, err
			}
			argument, err := p.ParseFunctionArgument()
			if err != nil {
				return nil, err
			}
			Arguments = append(Arguments, *argument)
		}
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
		Args:       Arguments,
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

		if p.matchCurrent(lexer.Punctuator, ";") && ImplicitReturn == nil {
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
		fallthrough
	case p.matchCurrent(lexer.Operator, "@") && p.matchNext(lexer.Identifier, "", 1):
		return p.ParseAssignment()
	case p.matchCurrent(lexer.Identifier, "") && p.matchNext(lexer.Punctuator, "[", 1):
		idx, err := p.ParseIndex()
		if err != nil {
			return nil, err
		}
		if p.matchCurrent(lexer.Operator, "=") {
			_, err := p.Expect(lexer.Operator, "=")
			if err != nil {
				return nil, err
			}
			val, err := p.ParseValue()
			if err != nil {
				return nil, err
			}
			assignment := &ast.Assignment{
				Target: idx,
				Value:  val,
			}
			assignment.SetPosition(idx.GetPosition())
			return assignment, nil
		} else {
			return p.parseBinary(idx)
		}
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
	var Type ast.Type
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
		Type:       Type,
		Value:      Value,
	}
	bind.SetPosition(let_tk.Position)

	return bind, nil
}

// ParseDereference parses a dereference according to the grammar:
//
// deref                ::= "@" identifier
func (p *Parser) ParseDereference() (*ast.Dereference, error) {
	if _, err := p.Expect(lexer.Operator, "@"); err != nil {
		return nil, err
	}

	Identifier, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	deref := &ast.Dereference{
		Value: Identifier,
	}
	deref.SetPosition(Identifier.GetPosition())

	return deref, nil
}

// ParseAssignment parses an assignment expression according to the grammar:
// (without the index, just the identifier)
//
// assignment           ::= identifier | index | deref "=" value
func (p *Parser) ParseAssignment() (*ast.Assignment, error) {
	var target ast.Primary

	if p.matchCurrent(lexer.Operator, "@") {
		deref, err := p.ParseDereference()
		if err != nil {
			return nil, err
		}
		target = deref
	} else {
		id, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		target = id
	}

	_, err := p.Expect(lexer.Operator, "=")
	if err != nil {
		return nil, err
	}

	Value, err := p.ParseValue()
	if err != nil {
		return nil, err
	}

	assignment := &ast.Assignment{
		Target: target,
		Value:  Value,
	}
	assignment.SetPosition(target.GetPosition())

	return assignment, nil
}

// ParseValue parses a value expression according to the grammar:
//
//	value                ::= primary
//	                       | binary
//	                       | unary
func (p *Parser) ParseValue() (ast.Value, error) {
	if p.matchCurrent(lexer.Operator, "") && !p.matchCurrent(lexer.Operator, "@") {
		return p.ParseUnary()
	}

	primary, err := p.ParsePrimary()
	if err != nil {
		return nil, err
	}

	return p.parseBinary(primary)
}

func (p *Parser) parseBinary(left ast.Primary) (ast.Value, error) {
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
			Left:     left,
			Operator: operator,
			Right:    right,
		}
		binary.SetPosition(left.GetPosition())

		return binary, nil
	}

	return left, nil
}

// ParsePrimary parses a primary expression according to the grammar:
//
//	primary              ::= literal
//	                       | identifier
//	                       | call
//	                       | separated
//	                       | block
//	                       | condition
//	                       | index
//	                       | deref
func (p *Parser) ParsePrimary() (ast.Primary, error) {
	switch {
	case p.matchCurrent(lexer.Operator, "@"):
		return p.ParseDereference()
	case p.matchCurrent(lexer.Literal, ""):
		return p.ParseLiteral()
	case p.matchCurrent(lexer.Identifier, "") && p.matchNext(lexer.Punctuator, "(", 1):
		return p.ParseCall()
	case p.matchCurrent(lexer.Identifier, "") && p.matchNext(lexer.Punctuator, "[", 1):
		return p.ParseIndex()
	case p.matchCurrent(lexer.Identifier, ""):
		return p.ParseIdentifier()
	case p.matchCurrent(lexer.Punctuator, "("):
		return p.ParseSeparated()
	case p.matchCurrent(lexer.Punctuator, "{"):
		return p.ParseBlock()
	case p.matchCurrent(lexer.Keyword, lexer.KeywordIf):
		return p.ParseCondition()
	case p.matchCurrent(lexer.Punctuator, "["):
		return p.ParseArrayLiteral()
	default:
		return nil, fmt.Errorf("unexpected primary expression")
	}
}

// ParseArrayLiteral parses an array literal according to the grammar:
//
// array_literal        ::= "[" [ value { "," value } ] "]"
func (p *Parser) ParseArrayLiteral() (*ast.ArrayLiteral, error) {
	var Values []ast.Value

	start_tk, err := p.Expect(lexer.Punctuator, "[")
	if err != nil {
		return nil, err
	}

	for !p.matchCurrent(lexer.Punctuator, "]") {
		val, err := p.ParseValue()
		if err != nil {
			return nil, err
		}
		Values = append(Values, val)

		if p.matchCurrent(lexer.Punctuator, ",") {
			p.next()
		} else if !p.matchCurrent(lexer.Punctuator, "]") {
			return nil, parseError("expected ',' or ']' in array literal", p.peek().Position)
		}
	}

	_, err = p.Expect(lexer.Punctuator, "]")
	if err != nil {
		return nil, err
	}

	arrayLiteral := &ast.ArrayLiteral{
		Values: Values,
	}
	arrayLiteral.SetPosition(start_tk.Position)

	return arrayLiteral, nil
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

		if p.matchCurrent(lexer.Punctuator, ",") {
			_, err = p.Expect(lexer.Punctuator, ",")
			if err != nil {
				return nil, err
			}
		} else if !p.matchCurrent(lexer.Punctuator, ")") {
			return nil, parseError("expected ',' or ')' in function call arguments", p.peek().Position)
		}
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

// ParseIndex parses an index expression according to the grammar:
//
// index                ::= identifier "[" primary "]"
func (p *Parser) ParseIndex() (*ast.Index, error) {
	var Identifier *ast.Identifier
	var Index ast.Primary

	Identifier, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	if _, err := p.Expect(lexer.Punctuator, "["); err != nil {
		return nil, err
	}

	Index, err = p.ParsePrimary()
	if err != nil {
		return nil, err
	}

	if _, err := p.Expect(lexer.Punctuator, "]"); err != nil {
		return nil, err
	}

	idx := ast.Index{
		Identifier: Identifier,
		Index:      Index,
	}

	idx.SetPosition(Identifier.Position)

	return &idx, nil
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

// ParseFunctionArgument parses a function argument according to the grammar:
//
// function_argument    ::= type identifier
func (p *Parser) ParseFunctionArgument() (*ast.Argument, error) {
	var Type ast.Type
	var Identifier *ast.Identifier

	Type, err := p.ParseType()
	if err != nil {
		return nil, err
	}

	Identifier, err = p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	return &ast.Argument{
		Type:       Type,
		Identifier: Identifier,
	}, nil
}

// ParseType parses a type according to the grammar:
//
// type                 ::= basic_type | array_type | slice_type | pointer_type
func (p *Parser) ParseType() (ast.Type, error) {
	if p.matchCurrent(lexer.Punctuator, "[") {
		return p.ParseBracketedType()
	} else if p.matchCurrent(lexer.Operator, "^") {
		p.next()
		t, err := p.ParseBasicType()
		if err != nil {
			return nil, err
		}
		return &ast.PointerType{
			Inner: t,
		}, nil
	} else if p.matchCurrent(lexer.Identifier, "") {
		return p.ParseBasicType()
	}
	return nil, parseError("invalid type", p.peek().Position)
}

// ParseBracketedType parses a type that starts with an opening bracket,
// which can be either an array_type or a slice_type (anonymous or named).
//
// array_type           ::= "[" int_literal "]" basic_type
// slice_type           ::= "[" [identifier] "]" basic_type
func (p *Parser) ParseBracketedType() (ast.Type, error) {
	if _, err := p.Expect(lexer.Punctuator, "["); err != nil {
		return nil, err
	}

	var length *int
	var lengthId *ast.Identifier

	if p.matchCurrent(lexer.Literal, "") {
		tk, err := p.next()
		if err != nil {
			return nil, err
		}
		l, err := strconv.Atoi(tk.Value)
		if err != nil {
			return nil, err
		}
		length = &l
	} else if p.matchCurrent(lexer.Identifier, "") {
		id, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		lengthId = id
	}

	if _, err := p.Expect(lexer.Punctuator, "]"); err != nil {
		return nil, err
	}

	basicType, err := p.ParseBasicType()
	if err != nil {
		return nil, err
	}

	if length != nil {
		return &ast.ArrayType{
			Element: *basicType,
			Length:  *length,
		}, nil
	}

	return &ast.SliceType{
		Element:          *basicType,
		LengthIdentifier: lengthId,
	}, nil
}

// ParseBasicType parses a basic type according to the grammar:
//
// basic_type           ::= "int" | "bool" | "float" | "string" | "unit"
func (p *Parser) ParseBasicType() (*ast.BasicType, error) {
	var BasicType ast.BasicType

	tk, err := p.Expect(lexer.Identifier, "")
	if err != nil {
		return nil, err
	}

	switch tk.Value {
	case "int":
		BasicType = ast.Int
	case "bool":
		BasicType = ast.Bool
	case "float":
		BasicType = ast.Float
	case "string":
		BasicType = ast.String
	case "unit":
		BasicType = ast.Unit
	default:
		return nil, parseError(fmt.Sprintf("invalid type %s", tk.Value), tk.Position)
	}

	return &BasicType, nil
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
