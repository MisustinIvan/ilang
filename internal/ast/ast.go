// Implements the AST nodes for the grammar.
package ast

import (
	"fmt"

	"github.com/MisustinIvan/ilang/internal/lexer"
)

type (
	Visitor interface {
		VisitProgram(p *Program) error
		VisitDeclaration(d *Declaration) error
		VisitExternalDeclaration(d *ExternalDeclaration) error
		VisitArgument(a *Argument) error
		VisitBasicType(t *BasicType) error
		VisitArrayType(t *ArrayType) error
		VisitSliceType(t *SliceType) error
		VisitPointerType(t *PointerType) error
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
		VisitIndex(i *Index) error
		VisitAssignment(a *Assignment) error
		VisitArrayLiteral(a *ArrayLiteral) error
		VisitDereference(d *Dereference) error
	}

	Node interface{ Accept(Visitor) error }
)

type (
	Program struct {
		Declarations         []*Declaration
		ExternalDeclarations []*ExternalDeclaration
	}

	Declaration struct {
		Type       BasicType
		Identifier *Identifier
		Args       []Argument
		Body       Block
	}

	ExternalDeclaration struct {
		Type       Type
		Identifier *Identifier
		Args       []Argument
		Variadic   bool
	}

	Argument struct {
		Type       Type
		Identifier *Identifier
	}
)

func (p *Program) Accept(v Visitor) error             { return v.VisitProgram(p) }
func (d *Declaration) Accept(v Visitor) error         { return v.VisitDeclaration(d) }
func (d *ExternalDeclaration) Accept(v Visitor) error { return v.VisitExternalDeclaration(d) }
func (a *Argument) Accept(v Visitor) error            { return v.VisitArgument(a) }

type Type interface {
	String() string
	Size() int
	Accept(v Visitor) error
	Equals(o Type) bool
}

//go:generate stringer -type=BasicType
type BasicType int

const (
	Int BasicType = iota
	Float
	Bool
	String
	Unit
	Undefined
)

func BasicTypePtr(t BasicType) *BasicType {
	return &t
}

func (b *BasicType) Size() int {
	switch *b {
	case Int, Float, Bool, String:
		return 8
	case Unit, Undefined:
		fallthrough
	default:
		return 0
	}
}

func (b *BasicType) Accept(v Visitor) error {
	return v.VisitBasicType(b)
}

func (b *BasicType) Equals(o Type) bool {
	val, ok := o.(*BasicType)
	return ok && *val == *b
}

type ArrayType struct {
	Element BasicType
	Length  int
}

func (t *ArrayType) Size() int {
	return t.Element.Size() * t.Length
}

func (t *ArrayType) String() string {
	return fmt.Sprintf("[%d]%s", t.Length, t.Element.String())
}

func (t *ArrayType) Accept(v Visitor) error {
	return v.VisitArrayType(t)
}

func (t *ArrayType) Equals(o Type) bool {
	if val, ok := o.(*ArrayType); ok {
		return t.Length == val.Length && t.Element == val.Element
	}
	return false
}

type SliceType struct {
	Element          BasicType
	LengthIdentifier *Identifier
}

func (t *SliceType) Size() int {
	return 16 // pointer + length
}

func (t *SliceType) String() string {
	if t.LengthIdentifier != nil {
		return fmt.Sprintf("[%s]%s", t.LengthIdentifier.Name, t.Element.String())
	}
	return fmt.Sprintf("[]%s", t.Element.String())
}

func (t *SliceType) Accept(v Visitor) error {
	return v.VisitSliceType(t)
}

func (t *SliceType) Equals(o Type) bool {
	if val, ok := o.(*ArrayType); ok {
		return t.Element == val.Element
	}
	if val, ok := o.(*SliceType); ok {
		return t.Element == val.Element
	}
	return false
}

type PointerType struct {
	Inner *BasicType
}

func (t *PointerType) Size() int {
	return 8
}

func (t *PointerType) String() string {
	return fmt.Sprintf("^%s", t.Inner.String())
}

func (t *PointerType) Accept(v Visitor) error {
	return v.VisitPointerType(t)
}

func (t *PointerType) Equals(o Type) bool {
	if pointerType, ok := o.(*PointerType); ok {
		return t.Inner.Equals(pointerType.Inner)
	}
	return false
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
	Inequality
	Less
	Greater
	LessEqual
	GreaterEqual
	ShiftLeft
	ShiftRight
	LogicAnd
	LogicOr
)

var BinaryOperatorTokens = map[string]BinaryOperator{
	"+":  Addition,
	"-":  Subtraction,
	"*":  Multiplication,
	"/":  Division,
	"==": Equality,
	"!=": Inequality,
	"<":  Less,
	">":  Greater,
	"<=": LessEqual,
	">=": GreaterEqual,
	"<<": ShiftLeft,
	">>": ShiftRight,
	"&&": LogicAnd,
	"||": LogicOr,
}

var BinaryOperatorApplies = map[BinaryOperator]map[BasicType]bool{
	Addition:       {Int: true, Float: true},
	Subtraction:    {Int: true, Float: true},
	Multiplication: {Int: true, Float: true},
	Division:       {Int: true, Float: true},
	Equality:       {Int: true, Float: true, Bool: true},
	Inequality:     {Int: true, Float: true, Bool: true},
	Less:           {Int: true, Float: true},
	Greater:        {Int: true, Float: true},
	LessEqual:      {Int: true, Float: true},
	GreaterEqual:   {Int: true, Float: true},
	ShiftLeft:      {Int: true},
	ShiftRight:     {Int: true},
	LogicAnd:       {Bool: true},
	LogicOr:        {Bool: true},
}

var BoolOperators = map[BinaryOperator]bool{
	Equality:     true,
	Inequality:   true,
	Less:         true,
	Greater:      true,
	LessEqual:    true,
	GreaterEqual: true,
}

//go:generate stringer -type=UnaryOperator
type UnaryOperator int

const (
	Inversion UnaryOperator = iota
	LogicNegation
	AddressOf
)

var UnaryOperatorTokens = map[string]UnaryOperator{
	"-": Inversion,
	"!": LogicNegation,
	"^": AddressOf,
}

var UnaryOperatorApplies = map[UnaryOperator]map[BasicType]bool{
	Inversion:     {Int: true, Float: true},
	LogicNegation: {Bool: true},
	AddressOf:     {Bool: true, Int: true, Float: true},
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
		Target Primary
		Value  Value
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
		Name     string
		Resolved *Identifier
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
	Index struct {
		PrimaryBase
		Identifier *Identifier
		Index      Value
	}
	ArrayLiteral struct {
		PrimaryBase
		Values []Value
	}
	Dereference struct {
		PrimaryBase
		Value *Identifier
	}
)

func (l *Literal) Accept(v Visitor) error      { return v.VisitLiteral(l) }
func (i *Identifier) Accept(v Visitor) error   { return v.VisitIdentifier(i) }
func (c *Call) Accept(v Visitor) error         { return v.VisitCall(c) }
func (s *Separated) Accept(v Visitor) error    { return v.VisitSeparated(s) }
func (b *Block) Accept(v Visitor) error        { return v.VisitBlock(b) }
func (c *Condition) Accept(v Visitor) error    { return v.VisitCondition(c) }
func (c *Index) Accept(v Visitor) error        { return v.VisitIndex(c) }
func (a *ArrayLiteral) Accept(v Visitor) error { return v.VisitArrayLiteral(a) }
func (d *Dereference) Accept(v Visitor) error  { return v.VisitDereference(d) }
