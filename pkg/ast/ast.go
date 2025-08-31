package ast

import "lang/pkg/lexer"

// Contains definitions for ast nodes as defined by the grammar

type Program struct {
	Declarations []FunctionDeclaration
}

type FunctionDeclaration struct {
	Identifier *IdentifierExpression
	Parameters []ParameterDefinition
	Body       *BlockExpression
}

type ParameterDefinition struct {
	Name     string
	Type     Type
	Position lexer.Position
}

type Expression interface {
	expression_mark()
	GetType() Type
	GetPosition() lexer.Position
}

type Expression_i struct {
	Type     Type
	Position lexer.Position
}

func (e *Expression_i) expression_mark()            {}
func (e *Expression_i) GetType() Type               { return e.Type }
func (e *Expression_i) GetPosition() lexer.Position { return e.Position }

type SimpleExpression interface {
	simple_expression_mark()
	Expression
}

type SimpleExpression_i struct {
	Expression_i
}

func (e *SimpleExpression_i) simple_expression_mark() {}

type PrimaryExpression interface {
	primary_expression_mark()
	SimpleExpression
}

type PrimaryExpression_i struct {
	SimpleExpression_i
}

func (e *PrimaryExpression_i) primary_expression_mark() {}

// actual expressions implementing grammar
// starting with primary expressions

type LiteralExpression struct {
	PrimaryExpression_i
	Value string
}

type IdentifierExpression struct {
	PrimaryExpression_i
	Value string
}

type CallExpression struct {
	PrimaryExpression_i
	Identifier *IdentifierExpression
	Params     []SimpleExpression
}

type BlockExpression struct {
	PrimaryExpression_i
	Body           []Expression
	ImplicitReturn Expression
}

type SeparatedExpression struct {
	PrimaryExpression_i
	Body SimpleExpression
}

type ConditionalExpression struct {
	PrimaryExpression_i
	Condition SimpleExpression
	IfBody    SimpleExpression
	ElseBody  SimpleExpression
}

// next are simple expressions

type BinaryExpression struct {
	SimpleExpression_i
	Left     PrimaryExpression
	Operator BinaryOperator
	Right    SimpleExpression
}

type UnaryExpression struct {
	SimpleExpression_i
	Operator UnaryOperator
	Value    PrimaryExpression
}

// next are highest level expressions

type BindExpression struct {
	Expression_i
	Identifier *IdentifierExpression
	TypeName   *IdentifierExpression
	Value      SimpleExpression
}

type ReturnExpression struct {
	Expression_i
	Value SimpleExpression
}

type AssignmentExpression struct {
	Expression_i
	Identifier *IdentifierExpression
	Value      SimpleExpression
}
