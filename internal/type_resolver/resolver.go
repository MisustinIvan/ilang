package type_resolver

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/MisustinIvan/ilang/internal/ast"
)

type Resolver struct {
	prog *ast.Program
}

func NewResolver(prog *ast.Program) *Resolver {
	return &Resolver{
		prog: prog,
	}
}

func (r *Resolver) ResolveTypes() (*ast.Program, error) { return r.prog, r.VisitProgram(r.prog) }
func (r *Resolver) VisitProgram(p *ast.Program) error {
	var err error
	for _, decl := range p.ExternalDeclarations {
		err = errors.Join(decl.Accept(r))
	}

	for _, decl := range p.Declarations {
		err = errors.Join(decl.Accept(r))
	}

	return err
}

func (r *Resolver) VisitDeclaration(d *ast.Declaration) error {
	var err error

	d.Identifier.SetType(d.Type)

	for _, param := range d.Params {
		err = errors.Join(param.Accept(r))
	}

	err = errors.Join(err, d.Body.Accept(r))

	return err
}

func (r *Resolver) VisitExternalDeclaration(d *ast.ExternalDeclaration) error {
	var err error
	d.Identifier.SetType(d.Type)

	for _, param := range d.Params {
		err = errors.Join(param.Accept(r))
	}

	return err
}

func (r *Resolver) VisitParameter(p *ast.Parameter) error {
	p.Identifier.SetType(p.Type)
	return nil
}

func (r *Resolver) VisitType(t *ast.Type) error { return nil }
func (r *Resolver) VisitReturn(e *ast.Return) error {
	var err error
	err = errors.Join(err, e.Value.Accept(r))
	e.SetType(e.Value.GetType())
	return err
}

func (r *Resolver) VisitBind(b *ast.Bind) error {
	b.Identifier.SetType(b.Type)
	b.SetType(b.Type)
	return b.Value.Accept(r)
}

func literalType(val string) ast.Type {
	t := ast.Undefined

	if val == "unit" {
		t = ast.Unit
	} else if val == "true" || val == "false" {
		t = ast.Bool
	} else if strings.HasPrefix(val, "\"") {
		t = ast.String
	} else if _, err := strconv.ParseInt(val, 10, 64); err == nil {
		t = ast.Int
	} else if _, err := strconv.ParseFloat(val, 64); err == nil {
		t = ast.Float
	}

	return t
}

func (r *Resolver) VisitLiteral(l *ast.Literal) error {
	t := literalType(l.Value)
	l.SetType(t)
	if t == ast.Undefined {
		return fmt.Errorf("literal of undefined type at %v", l.GetPosition())
	}
	return nil
}

func (r *Resolver) VisitIdentifier(i *ast.Identifier) error {
	if i.Resolved == nil {
		i.SetType(ast.Undefined)
		return fmt.Errorf("unresolved identifier at %s has undefined type", i.GetPosition())
	}
	i.SetType(i.Resolved.Type)
	return nil
}

func (r *Resolver) VisitCall(c *ast.Call) error {
	var err error
	err = errors.Join(err, c.Identifier.Accept(r))
	if c.Identifier.Resolved == nil {
		c.SetType(ast.Undefined)
		err = errors.Join(err, fmt.Errorf("unresolved function call at %s has undefined type", c.GetPosition()))
	} else {
		c.SetType(c.Identifier.Resolved.Type)
	}

	for _, arg := range c.Arguments {
		err = errors.Join(err, arg.Accept(r))
	}

	return err
}

func (r *Resolver) VisitSeparated(s *ast.Separated) error {
	err := s.Value.Accept(r)
	s.SetType(s.Value.GetType())
	return err
}

func (r *Resolver) VisitUnary(u *ast.Unary) error {
	err := u.Value.Accept(r)
	u.SetType(u.Value.GetType())
	return err
}

func (r *Resolver) VisitBinary(u *ast.Binary) error {
	var err error
	err = errors.Join(err, u.Left.Accept(r))
	err = errors.Join(err, u.Right.Accept(r))
	u.SetType(u.Left.GetType()) // assume this type and check the right side later
	return err
}

func (r *Resolver) VisitBlock(b *ast.Block) error {
	var err error

	for _, expr := range b.Body {
		err = errors.Join(err, expr.Accept(r))
	}

	if b.ImplicitReturn != nil {
		err = errors.Join(err, b.ImplicitReturn.Accept(r))
		b.SetType(b.ImplicitReturn.GetType())
	} else {
		b.SetType(ast.Undefined)
	}

	return err
}

func (r *Resolver) VisitCondition(c *ast.Condition) error {
	var err error

	err = errors.Join(err, c.Condition.Accept(r))
	err = errors.Join(err, c.Body.Accept(r))
	c.SetType(c.Body.GetType())
	if c.Else != nil {
		err = errors.Join(err, c.Else.Accept(r))
	}

	return err
}

func (r *Resolver) VisitAssignment(a *ast.Assignment) error {
	var err error

	err = errors.Join(err, a.Identifier.Accept(r))
	err = errors.Join(err, a.Value.Accept(r))
	a.SetType(a.Identifier.Type)

	return err
}
