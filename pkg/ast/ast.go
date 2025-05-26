package ast

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Prog struct {
	Declarations []FunctionDeclaration
}

type FunctionDeclaration struct {
	Type       Type
	Name       Identifier
	Parameters []Parameter
	Body       BlockExpression
}

type Type int

const (
	Integer Type = iota
	Float
	Boolean
	String
	Unit
)

func (t Type) String() string {
	s := "UNKNOWN"
	switch t {
	case Integer:
		s = "int"
	case Float:
		s = "float"
	case Boolean:
		s = "bool"
	case String:
		s = "string"
	case Unit:
		s = "unit"
	}
	return s
}

var Types = map[string]Type{
	"int":    Integer,
	"float":  Float,
	"bool":   Boolean,
	"string": String,
	"unit":   Unit,
}

type Parameter struct {
	Name Identifier
	Type Type
}

type BlockExpression struct {
	BasePrimaryExpression
	Body                     []Expression
	ImplicitReturnExpression Expression
}

type Expression interface {
	expression_mark()
}

type BaseExpression struct{}

func (e *BaseExpression) expression_mark() {}

type BindExpression struct {
	BaseExpression
	Left  Identifier
	Type  Type
	Right Expression
}

type ReturnExpression struct {
	BaseExpression
	Value Expression
}

type AssignmentExpression struct {
	BaseExpression
	Left  Identifier
	Right Expression
}

type BinaryExpression struct {
	BaseExpression
	Left     Expression
	Operator BinaryOperator
	Right    Expression
}

type BinaryOperator int

const (
	Addition BinaryOperator = iota
	Subtraction
	Multiplication
	Division
	Equality
	LesserThan
	GreaterThan
	LesserOrEqualThan
	GreaterOrEqualThan
	LeftShift
	RightShift
	LogicAnd
	LogicOr
)

func (o BinaryOperator) String() string {
	s := "UNKNOWN"
	switch o {
	case Addition:
		s = "Addition"
	case Subtraction:
		s = "Subtraction"
	case Multiplication:
		s = "Multiplication"
	case Division:
		s = "Division"
	case Equality:
		s = "Equality"
	case LesserThan:
		s = "LesserThan"
	case GreaterThan:
		s = "GreaterThan"
	case LesserOrEqualThan:
		s = "LesserOrEqualThan"
	case GreaterOrEqualThan:
		s = "GreaterOrEqualThan"
	case LeftShift:
		s = "LeftShift"
	case RightShift:
		s = "RightShift"
	case LogicAnd:
		s = "LogicAnd"
	case LogicOr:
		s = "LogicOr"
	}

	return s
}

var BinaryOperators = map[string]BinaryOperator{
	"+":  Addition,
	"-":  Subtraction,
	"*":  Multiplication,
	"/":  Division,
	"==": Equality,
	"<":  LesserThan,
	">":  GreaterThan,
	"<=": LesserOrEqualThan,
	">=": GreaterOrEqualThan,
	"<<": LeftShift,
	">>": RightShift,
	"&&": LogicAnd,
	"||": LogicOr,
}

type ConditionalExpression struct {
	PrimaryExpression
	Condition Expression
	IfBody    BlockExpression
	ElseBody  BlockExpression
}

type PrimaryExpression interface {
	Expression
	primary_expression_mark()
}

type BasePrimaryExpression struct {
	BaseExpression
}

func (e *BasePrimaryExpression) primary_expression_mark() {}

type Literal struct {
	BasePrimaryExpression
	Value string
	Type  Type
}

func isInteger(x string) bool {
	_, err := strconv.ParseInt(x, 10, 64)
	return err == nil
}

func isFloat(x string) bool {
	_, err := strconv.ParseFloat(x, 10)
	return err == nil
}

func isBoolean(x string) bool {
	return x == "true" || x == "false"
}

func isString(x string) bool {
	return strings.HasPrefix(x, "\"") && strings.HasSuffix(x, "\"")
}

func isUnit(x string) bool {
	return x == "unit"
}

func LiteralType(l string) Type {
	switch {
	case isInteger(l):
		return Integer
	case isFloat(l):
		return Float
	case isBoolean(l):
		return Boolean
	case isString(l):
		return String
	case isUnit(l):
		return Unit
	default:
		fmt.Printf("Literal %s has unknown type\n", l)
		os.Exit(-1)
		return Unit
	}
}

type Identifier struct {
	BasePrimaryExpression
	Value string
}

type FunctionCall struct {
	BasePrimaryExpression
	Function Identifier
	Params   []Expression
}

// block expression implemented above

type SeparatedExpression struct {
	BasePrimaryExpression
	Value Expression
}
