// Implements the AST nodes for the grammar.
package ast

import (
	"github.com/MisustinIvan/ilang/internal/lexer"
)

type (
	Visitor interface {
		VisitProgram(p *Program) error
		VisitDeclaration(d *Declaration) error
		VisitExternalDeclaration(d *ExternalDeclaration) error
		VisitParameter(p *Parameter) error
		VisitType(t *Type) error
		VisitReturn(r *Return) error
		VisitBind(b *Bind) error
		VisitLiteral(l *Literal) error
		VisitIdentifier(i *Identifier) error
		VisitCall(c *Call) error
		VisitSeparated(s *Separated) error
		VisitUnary(u *Unary) error
		VisitBinary(u *Binary) error
		VisitBlock(b *Block) error
		VisitCondition(c *Condition) error
		VisitAssignment(a *Assignment) error
	}

	Node interface{ Accept(Visitor) error }
)

type (
	Program struct {
		Declarations         []*Declaration
		ExternalDeclarations []*ExternalDeclaration
	}

	Declaration struct {
		Type       Type
		Identifier *Identifier
		Params     []Parameter
		Body       Block
	}

	ExternalDeclaration struct {
		Type       Type
		Identifier *Identifier
		Params     []Parameter
	}

	Parameter struct {
		Type       Type
		Identifier *Identifier
	}
)

func (p *Program) Accept(v Visitor) error             { return v.VisitProgram(p) }
func (d *Declaration) Accept(v Visitor) error         { return v.VisitDeclaration(d) }
func (d *ExternalDeclaration) Accept(v Visitor) error { return v.VisitExternalDeclaration(d) }
func (p *Parameter) Accept(v Visitor) error           { return v.VisitParameter(p) }

// types
//
//go:generate stringer -type=Type
type Type int

const (
	Undefined Type = iota
	Int
	Bool
	Float
	String
	Unit
)

func (t *Type) Accept(v Visitor) error {
	return v.VisitType(t)
}

// operators
//
//go:generate stringer -type=BinaryOperator
type BinaryOperator int

const (
	Addition BinaryOperator = iota
	Subtraction
	Multiplication
	Division
	Equality
	LessThan
	GreaterThan
	LessThanEqual
	GreaterThanEqual
	LogicAnd
	LogicOr
)

var BinaryOperatorTokens = map[string]BinaryOperator{
	"+":  Addition,
	"-":  Subtraction,
	"*":  Multiplication,
	"/":  Division,
	"==": Equality,
	"<":  LessThan,
	">":  GreaterThan,
	"<=": LessThanEqual,
	">=": GreaterThanEqual,
	"&&": LogicAnd,
	"||": LogicOr,
}

//go:generate stringer -type=UnaryOperator
type UnaryOperator int

const (
	Inversion UnaryOperator = iota
	LogicNegation
)

var UnaryOperatorTokens = map[string]UnaryOperator{
	"-": Inversion,
	"!": LogicNegation,
}

// expressions

type (
	Expression interface {
		Node
		GetType() Type
		SetType(Type)
		GetPosition() lexer.Position
		SetPosition(lexer.Position)
	}

	ExpressionBase struct {
		Type     Type
		Position lexer.Position
	}
)

func (e *ExpressionBase) GetType() Type                { return e.Type }
func (e *ExpressionBase) SetType(t Type)               { e.Type = t }
func (e *ExpressionBase) GetPosition() lexer.Position  { return e.Position }
func (e *ExpressionBase) SetPosition(p lexer.Position) { e.Position = p }

type (
	Return struct {
		ExpressionBase
		Value Value
	}
	Bind struct {
		ExpressionBase
		Identifier *Identifier
		Type       Type
		Value      Value
	}
	Assignment struct {
		ExpressionBase
		Identifier *Identifier
		Value      Value
	}
)

func (r *Return) Accept(v Visitor) error     { return v.VisitReturn(r) }
func (b *Bind) Accept(v Visitor) error       { return v.VisitBind(b) }
func (a *Assignment) Accept(v Visitor) error { return v.VisitAssignment(a) }

type (
	Value     interface{ Expression }
	ValueBase struct{ ExpressionBase }
)

type (
	Binary struct {
		ValueBase
		Left     Primary
		Operator BinaryOperator
		Right    Value
	}
	Unary struct {
		ValueBase
		Operator UnaryOperator
		Value    Primary
	}
)

func (b *Binary) Accept(v Visitor) error { return v.VisitBinary(b) }
func (u *Unary) Accept(v Visitor) error  { return v.VisitUnary(u) }

type (
	Primary     interface{ Value }
	PrimaryBase struct{ ValueBase }
)

type (
	Literal struct {
		PrimaryBase
		Value string
	}
	Identifier struct {
		PrimaryBase
		Name string
	}
	Call struct {
		PrimaryBase
		Identifier *Identifier
		Arguments  []Value
	}
	Separated struct {
		PrimaryBase
		Value Primary
	}
	Block struct {
		PrimaryBase
		Body           []Expression
		ImplicitReturn Expression
	}
	Condition struct {
		PrimaryBase
		Condition Value
		Body      Value
		Else      Value
	}
)

func (l *Literal) Accept(v Visitor) error    { return v.VisitLiteral(l) }
func (i *Identifier) Accept(v Visitor) error { return v.VisitIdentifier(i) }
func (c *Call) Accept(v Visitor) error       { return v.VisitCall(c) }
func (s *Separated) Accept(v Visitor) error  { return v.VisitSeparated(s) }
func (b *Block) Accept(v Visitor) error      { return v.VisitBlock(b) }
func (c *Condition) Accept(v Visitor) error  { return v.VisitCondition(c) }
