package ast

import (
	"fmt"
	"lang/pkg/lexer"
	"strconv"
	"strings"
)

type Prog struct {
	Declarations []FunctionDeclaration
}

type Identifier struct {
	Name string
}

type Type int

const (
	Void Type = iota
	Integer
	Float
	Boolean
	String
)

func (t *Type) String() string {
	s := "UNKNOWN"

	switch *t {
	case Void:
		s = "Void"
	case Integer:
		s = "Integer"
	case Float:
		s = "Float"
	case Boolean:
		s = "Boolean"
	case String:
		s = "String"
	}

	return s
}

var Types = map[string]Type{
	"void":   Void,
	"int":    Integer,
	"float":  Float,
	"bool":   Boolean,
	"string": String,
}

func isInteger(val string) bool {
	_, err := strconv.ParseInt(val, 10, 64)
	return err == nil
}

func isFloat(val string) bool {
	_, err := strconv.ParseFloat(val, 64)
	return err == nil
}

func isBoolean(val string) bool {
	return val == "true" || val == "false"
}

func isString(val string) bool {
	return strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")
}

func LiteralType(token lexer.Token) (Type, bool) {
	if token.Kind != lexer.Literal {
		return Void, false
	}

	if isInteger(token.Value) {
		return Integer, true
	}

	if isFloat(token.Value) {
		return Float, true
	}

	if isBoolean(token.Value) {
		return Boolean, true
	}

	if isString(token.Value) {
		return String, true
	}

	return Void, false
}

type FunctionDeclaration struct {
	Name           Identifier
	Type           Type
	ParameterTypes []ParameterType
	Body           []Statement
}

type ParameterType struct {
	Type Type
	Name Identifier
}

type Statement interface {
	statement_mark()
	String() string
}

type FunctionCallStatement struct {
	Function Identifier
	Args     []Expression
}

func (s *FunctionCallStatement) statement_mark() {}
func (s *FunctionCallStatement) String() string {
	r := s.Function.Name
	r += "("
	for _, arg := range s.Args {
		r += arg.String() + ", "
	}
	if len(s.Args) > 0 {
		r = r[:len(r)-3]
	}
	r += ")"
	return r
}

type AssignmentStatement struct {
	Left  Identifier
	Right Expression
}

func (s *AssignmentStatement) statement_mark() {}
func (s *AssignmentStatement) String() string {
	return fmt.Sprintf("%s = %s", s.Left.Name, s.Right.String())
}

type VariableDeclarationStatement struct {
	Left  Identifier
	Type  Type
	Right Expression
}

func (s *VariableDeclarationStatement) statement_mark() {}
func (s *VariableDeclarationStatement) String() string {
	return fmt.Sprintf("%s %s = %s", s.Type.String(), s.Left.Name, s.Right.String())
}

type CommentStatement struct {
	Value string
}

func (s *CommentStatement) statement_mark() {}
func (s *CommentStatement) String() string  { return s.Value }

type ReturnStatement struct {
	Value Expression
}

func (s *ReturnStatement) statement_mark() {}
func (s *ReturnStatement) String() string {
	return "return " + s.Value.String()
}

type Expression interface {
	expression_mark()
	String() string
}

type EmptyExpression struct{}

func (e *EmptyExpression) expression_mark() {}
func (e *EmptyExpression) String() string   { return "empty expression" }

type LiteralExpression struct {
	Type  Type
	Value string
}

func (e *LiteralExpression) expression_mark() {}
func (e *LiteralExpression) String() string {
	return e.Value
}

type IdentifierExpression struct{ Identifier Identifier }

func (e *IdentifierExpression) expression_mark() {}
func (e *IdentifierExpression) String() string {
	return e.Identifier.Name
}

type CallExpression struct {
	Function Identifier
	Args     []Expression
}

func (e *CallExpression) expression_mark() {}
func (e *CallExpression) String() string {
	r := e.Function.Name
	r += "("
	for _, arg := range e.Args {
		r += arg.String() + ", "
	}
	if len(e.Args) > 0 {
		r = r[:len(e.Args)-3]
	}
	r += ")"
	return r
}
