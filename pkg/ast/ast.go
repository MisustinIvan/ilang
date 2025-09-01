package ast

import "lang/pkg/lexer"

// Contains definitions for ast nodes as defined by the grammar

type AstVisitor interface {
	VisitProgram(p *Program) error
	VisitFunctionDeclaration(d *FunctionDeclaration) error
	VisitExternalFunctionDeclaration(e *ExternalFunctionDeclaration) error
	VisitParameterDefinition(d *ParameterDefinition) error
	VisitBind(e *BindExpression) error
	VisitReturn(e *ReturnExpression) error
	VisitBinary(e *BinaryExpression) error
	VisitLiteral(e *LiteralExpression) error
	VisitIdentifier(e *IdentifierExpression) error
	VisitCall(e *CallExpression) error
	VisitBlock(e *BlockExpression) error
	VisitSeparated(e *SeparatedExpression) error
	VisitUnary(e *UnaryExpression) error
	VisitConditional(e *ConditionalExpression) error
	VisitAssignment(e *AssignmentExpression) error
}

type Node interface {
	Accept(AstVisitor) error
}

type Program struct {
	Declarations         []*FunctionDeclaration
	ExternalDeclarations []*ExternalFunctionDeclaration
}

func (p *Program) Accept(v AstVisitor) error {
	return v.VisitProgram(p)
}

type FunctionDeclaration struct {
	Type       Type
	Identifier *IdentifierExpression
	TypeName   *IdentifierExpression
	Parameters []ParameterDefinition
	Body       *BlockExpression
}

func (d *FunctionDeclaration) Accept(v AstVisitor) error {
	return v.VisitFunctionDeclaration(d)
}

type ParameterDefinition struct {
	Name     *IdentifierExpression
	TypeName *IdentifierExpression
}

func (d *ParameterDefinition) Accept(v AstVisitor) error {
	return v.VisitParameterDefinition(d)
}

type ExternalFunctionDeclaration struct {
	Type       Type
	Identifier *IdentifierExpression
	TypeName   *IdentifierExpression
	Parameters []ParameterDefinition
}

func (d *ExternalFunctionDeclaration) Accept(v AstVisitor) error {
	return v.VisitExternalFunctionDeclaration(d)
}

type Expression interface {
	Node
	expression_mark()
	GetType() Type
	SetType(Type)
	GetPosition() lexer.Position
	SetPosition(lexer.Position)
}

type Expression_i struct {
	Type     Type
	Position lexer.Position
}

func (e *Expression_i) expression_mark()             {}
func (e *Expression_i) GetType() Type                { return e.Type }
func (e *Expression_i) GetPosition() lexer.Position  { return e.Position }
func (e *Expression_i) SetType(t Type)               { e.Type = t }
func (e *Expression_i) SetPosition(p lexer.Position) { e.Position = p }

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

func (e *LiteralExpression) Accept(v AstVisitor) error {
	return v.VisitLiteral(e)
}

type IdentifierExpression struct {
	PrimaryExpression_i
	Value    string
	Resolved *IdentifierExpression
}

func (e *IdentifierExpression) Accept(v AstVisitor) error {
	return v.VisitIdentifier(e)
}

type CallExpression struct {
	PrimaryExpression_i
	Identifier *IdentifierExpression
	Params     []SimpleExpression
}

func (e *CallExpression) Accept(v AstVisitor) error {
	return v.VisitCall(e)
}

type BlockExpression struct {
	PrimaryExpression_i
	Body           []Expression
	ImplicitReturn Expression
}

func (e *BlockExpression) Accept(v AstVisitor) error {
	return v.VisitBlock(e)
}

type SeparatedExpression struct {
	PrimaryExpression_i
	Body SimpleExpression
}

func (e *SeparatedExpression) Accept(v AstVisitor) error {
	return v.VisitSeparated(e)
}

type ConditionalExpression struct {
	PrimaryExpression_i
	Condition SimpleExpression
	IfBody    SimpleExpression
	ElseBody  SimpleExpression
}

func (e *ConditionalExpression) Accept(v AstVisitor) error {
	return v.VisitConditional(e)
}

// next are simple expressions

type BinaryExpression struct {
	SimpleExpression_i
	Left     PrimaryExpression
	Operator BinaryOperator
	Right    SimpleExpression
}

func (e *BinaryExpression) Accept(v AstVisitor) error {
	return v.VisitBinary(e)
}

type UnaryExpression struct {
	SimpleExpression_i
	Operator UnaryOperator
	Value    PrimaryExpression
}

func (e *UnaryExpression) Accept(v AstVisitor) error {
	return v.VisitUnary(e)
}

// next are highest level expressions

type BindExpression struct {
	Expression_i
	Identifier *IdentifierExpression
	TypeName   *IdentifierExpression
	Value      SimpleExpression
}

func (e *BindExpression) Accept(v AstVisitor) error {
	return v.VisitBind(e)
}

type ReturnExpression struct {
	Expression_i
	Value SimpleExpression
}

func (e *ReturnExpression) Accept(v AstVisitor) error {
	return v.VisitReturn(e)
}

type AssignmentExpression struct {
	Expression_i
	Identifier *IdentifierExpression
	Value      SimpleExpression
}

func (e *AssignmentExpression) Accept(v AstVisitor) error {
	return v.VisitAssignment(e)
}
