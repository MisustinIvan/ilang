package type_checker

import (
	"errors"
	"fmt"

	"github.com/MisustinIvan/ilang/internal/ast"
	"github.com/MisustinIvan/ilang/internal/lexer"
)

func typeError(position lexer.Position, format string, args ...any) error {
	return fmt.Errorf("%v at %s", fmt.Errorf(format, args...), position)
}

type Checker struct {
	prog         *ast.Program
	declarations map[*ast.Identifier][]ast.Parameter
}

func NewChecker(prog *ast.Program) *Checker {
	c := &Checker{
		prog:         prog,
		declarations: make(map[*ast.Identifier][]ast.Parameter),
	}

	for _, decl := range prog.Declarations {
		c.declarations[decl.Identifier] = decl.Params
	}

	for _, decl := range prog.ExternalDeclarations {
		c.declarations[decl.Identifier] = decl.Params
	}

	return c
}

func (c *Checker) ResolveTypes() (*ast.Program, error) { return c.prog, c.VisitProgram(c.prog) }
func (c *Checker) VisitProgram(p *ast.Program) error {
	var err error

	for _, decl := range p.ExternalDeclarations {
		err = errors.Join(err, decl.Accept(c))
	}

	for _, decl := range p.Declarations {
		err = errors.Join(err, decl.Accept(c))
	}

	return err
}

func (c *Checker) VisitDeclaration(d *ast.Declaration) error {
	var err error

	err = errors.Join(err, d.Body.Accept(c))
	if d.Body.GetType() != d.Type {
		err = errors.Join(err, typeError(d.Body.Position, "body type: %v does not match function type: %v", d.Body.GetType(), d.Type))
	}

	return err
}

func (c *Checker) VisitExternalDeclaration(d *ast.ExternalDeclaration) error { return nil } // here everything should be fine
func (c *Checker) VisitParameter(p *ast.Parameter) error                     { return nil }
func (c *Checker) VisitType(t *ast.Type) error                               { return nil }
func (c *Checker) VisitReturn(e *ast.Return) error                           { return e.Value.Accept(c) }
func (c *Checker) VisitBind(b *ast.Bind) error {
	var err error

	if b.Type != b.Value.GetType() {
		err = errors.Join(err, typeError(b.Position, "bound value type: %v does not match expected type: %v", b.Value.GetType(), b.Type))
	}
	err = errors.Join(err, b.Value.Accept(c))

	return err
}

func (c *Checker) VisitLiteral(l *ast.Literal) error       { return nil }
func (c *Checker) VisitIdentifier(i *ast.Identifier) error { return nil }
func (c *Checker) VisitCall(cl *ast.Call) error {
	var err error

	declared_args, ok := c.declarations[cl.Identifier.Resolved]
	if !ok {
		return typeError(cl.Position, "calling unresolved function %s", cl.Identifier.Name)
	}

	param_len := max(len(cl.Arguments), len(declared_args))
	for i := range param_len {
		if i < len(cl.Arguments) {
			err = errors.Join(err, cl.Arguments[i].Accept(c))
		}
		if i >= len(declared_args) {
			err = errors.Join(err, typeError(cl.Position, "unexpected function call argument"))
			continue
		}
		if i >= len(cl.Arguments) {
			err = errors.Join(err, typeError(cl.Position, "missing function call argument"))
			continue
		}
		expected := declared_args[i].Type
		got := cl.Arguments[i].GetType()

		if expected != got {
			err = errors.Join(err, typeError(cl.Position, "parameter types dont match - %v vs %v - at index %d", expected, got, i))
		}
	}

	return err
}

func (c *Checker) VisitSeparated(s *ast.Separated) error { return s.Value.Accept(c) }
func (c *Checker) VisitUnary(u *ast.Unary) error {
	var err error

	err = errors.Join(err, u.Value.Accept(c))
	if !ast.UnaryOperatorApplies[u.Operator][u.Value.GetType()] {
		err = errors.Join(err, typeError(u.Position, "unary operator does not apply to type %v", u.Value.GetType()))
	}

	return err
}

func (c *Checker) VisitBinary(u *ast.Binary) error {
	var err error

	err = errors.Join(err, u.Left.Accept(c))
	err = errors.Join(err, u.Right.Accept(c))

	if u.Left.GetType() != u.Right.GetType() {
		err = errors.Join(err, typeError(u.Position, "binary expression types dont match - %v vs %v", u.Left.GetType(), u.Right.GetType()))
	}
	if !ast.BinaryOperatorApplies[u.Operator][u.Left.GetType()] {
		err = errors.Join(err, typeError(u.Position, "binary operator %v does not apply to type %v", u.Operator.String(), u.Left.GetType()))
	}

	return err
}

func (c *Checker) VisitBlock(b *ast.Block) error {
	var err error

	for _, expr := range b.Body {
		err = errors.Join(err, expr.Accept(c))
	}
	if b.ImplicitReturn != nil {
		err = errors.Join(err, b.ImplicitReturn.Accept(c))
	}

	return err
}

func (c *Checker) VisitCondition(cd *ast.Condition) error {
	var err error

	err = errors.Join(err, cd.Condition.Accept(c))
	err = errors.Join(err, cd.Body.Accept(c))
	if cd.Else != nil {
		err = errors.Join(err, cd.Else.Accept(c))
	}

	if cd.Condition.GetType() != ast.Bool {
		err = errors.Join(err, typeError(cd.Condition.GetPosition(), "condition of type %v must be of type bool", cd.Condition.GetType()))
	}

	if cd.Else != nil {
		if cd.Body.GetType() != cd.Else.GetType() {
			err = errors.Join(err, typeError(cd.GetPosition(), "both branches of condition dont have the same type - %v vs %v", cd.Body.GetType(), cd.Else.GetType()))
		}
	}

	return err
}
func (c *Checker) VisitAssignment(a *ast.Assignment) error {
	var err error

	err = errors.Join(err, a.Value.Accept(c))
	if a.Identifier.Resolved != nil {
		if a.Value.GetType() != a.Identifier.Resolved.GetType() {
			err = errors.Join(err, typeError(a.GetPosition(), "assigning value of type %v to variable of type %v", a.Value.GetType(), a.Identifier.Resolved.GetType()))
		}
	}

	return err
}
