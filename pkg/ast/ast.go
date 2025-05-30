package ast

import "lang/pkg/lexer"

type Expression interface {
	expression_mark()
	GetType() Type
}

type PrimaryExpression interface {
	Expression
	primary_expression_mark()
}

type Expression_i struct {
	// methods
	Expression
	// fields
	Type
	lexer.TokenPosition
}

func (e *Expression_i) GetType() Type {
	return e.Type
}

type PrimaryExpression_i struct {
	// methods
	PrimaryExpression
	// fields
	Expression_i
}

func (e *PrimaryExpression_i) GetType() Type {
	return e.Type
}

type Prog struct {
	Declarations []FunctionDeclaration
}

type FunctionDeclaration struct {
	Position   lexer.TokenPosition
	Type       Type
	Name       Identifier
	Parameters []Parameter
	Body       BlockExpression
}

type Parameter struct {
	Type
	Name Identifier
}

// actual expressions

type BlockExpression struct {
	PrimaryExpression_i
	Body                     []Expression
	ImplicitReturnExpression Expression
}

type Literal struct {
	PrimaryExpression_i
	Value       string
	LookupValue string
}

type Identifier struct {
	PrimaryExpression_i
	Value       string
	LookupValue string
}

type ReturnExpression struct {
	Expression_i
	Value Expression
}

type UnaryExpression struct {
	PrimaryExpression_i
	Operator UnaryOperator
	Value    Expression
}

type BindExpression struct {
	Expression_i
	Left  Identifier
	Right Expression
}

type AssignmentExpression struct {
	Expression_i
	Left  Identifier
	Right Expression
}

type BinaryExpression struct {
	Expression_i
	Left     Expression
	Operator BinaryOperator
	Right    Expression
}

type SeparatedExpression struct {
	PrimaryExpression_i
	Value Expression
}

type ConditionalExpression struct {
	PrimaryExpression_i
	Condition Expression
	IfBody    BlockExpression
	ElseBody  BlockExpression
}

type ForExpression struct {
	PrimaryExpression_i
	Condition Expression
	Body      *BlockExpression
}

type BreakExpression struct {
	Expression_i
}

type FunctionCall struct {
	PrimaryExpression_i
	Function Identifier
	Params   []Expression
}
