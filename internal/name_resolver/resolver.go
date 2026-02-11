/*
Implements an ast.Visitor that resolves identifier names.
*/
package name_resolver

import (
	"errors"
	"fmt"

	"github.com/MisustinIvan/ilang/internal/ast"
)

type scope struct {
	parent *scope
	locals map[string]*ast.Identifier
}

type Resolver struct {
	program *ast.Program
	scope   *scope
}

func NewResolver(p *ast.Program) *Resolver {
	return &Resolver{
		program: p,
		scope:   nil,
	}
}

// PushScope creates a new empty scope with the current scope as the parent.
func (r *Resolver) PushScope() {
	r.scope = &scope{
		parent: r.scope,
		locals: map[string]*ast.Identifier{},
	}
}

// PopScope discards the innermost scope if it is not the root scope, in that
// case it simply does nothing.
func (r *Resolver) PopScope() {
	if r.scope != nil {
		r.scope = r.scope.parent
	}
}

// Declare declares a new local name from the identifier. In case the name is
// already declared Declare returns an error reporting the details. The Declare
// function allow for name shadowing and so it checks for conflicts only in the
// innermost scope.
func (r *Resolver) Declare(identifier *ast.Identifier) error {
	if val, exists := r.scope.locals[identifier.Name]; exists {
		return fmt.Errorf("identifier %s already declared at %s", identifier.Name, val.Position)
	}
	r.scope.locals[identifier.Name] = identifier
	return nil
}

// Lookup tries to find an identifier in the scopes starting from the innermost
// scope going upwards.
func (r *Resolver) Lookup(id string) *ast.Identifier {
	for scope := r.scope; scope != nil; scope = scope.parent {
		if val, ok := scope.locals[id]; ok {
			return val
		}
	}
	return nil
}

func (r *Resolver) ResolveNames() (*ast.Program, error) {
	return r.program, r.program.Accept(r)
}

// Implementing ast.Visitor interface...
func (r *Resolver) VisitProgram(p *ast.Program) error {
	r.PushScope() // global scope
	var err error
	for _, decl := range p.ExternalDeclarations {
		err = errors.Join(err, decl.Accept(r))
	}

	for _, decl := range p.Declarations {
		err = errors.Join(err, r.Declare(decl.Identifier))
	}

	r.PopScope() // global scope
	return err
}

func (r *Resolver) VisitDeclaration(d *ast.Declaration) error {
	err := r.Declare(d.Identifier) // declare in global scope
	r.PushScope()                  // function scope

	for _, param := range d.Params {
		err = errors.Join(param.Accept(r))
	}

	err = errors.Join(d.Body.Accept(r))

	r.PopScope() // function scope
	return err
}
func (r *Resolver) VisitExternalDeclaration(d *ast.ExternalDeclaration) error {
	var err error
	err = errors.Join(r.Declare(d.Identifier))
	r.PushScope() // function scope

	for _, param := range d.Params {
		err = errors.Join(param.Accept(r))
	}

	r.PopScope() // function scope
	return err
}

func (r *Resolver) VisitParameter(p *ast.Parameter) error {
	return r.Declare(p.Identifier)
}

func (r *Resolver) VisitReturn(e *ast.Return) error {
	return e.Value.Accept(r)
}

func (r *Resolver) VisitBind(b *ast.Bind) error {
	return errors.Join(b.Value.Accept(r), r.Declare(b.Identifier))
}

func (r *Resolver) VisitIdentifier(i *ast.Identifier) error {
	ref := r.Lookup(i.Name)
	if ref == nil {
		return fmt.Errorf("undeclared identifier %s at %s", i.Name, i.Position)
	}
	i.Resolved = ref
	return nil
}

func (r *Resolver) VisitCall(c *ast.Call) error {
	var err error

	err = errors.Join(err, c.Identifier.Accept(r))
	for _, arg := range c.Arguments {
		err = errors.Join(err, arg.Accept(r))
	}

	return err
}

func (r *Resolver) VisitSeparated(s *ast.Separated) error {
	return s.Value.Accept(r)
}

func (r *Resolver) VisitUnary(u *ast.Unary) error {
	return u.Value.Accept(r)
}

func (r *Resolver) VisitBinary(b *ast.Binary) error {
	return errors.Join(b.Left.Accept(r), b.Right.Accept(r))
}

func (r *Resolver) VisitBlock(b *ast.Block) error {
	var err error
	r.PushScope() // block scope

	for _, expr := range b.Body {
		err = errors.Join(err, expr.Accept(r))
	}

	if b.ImplicitReturn != nil {
		err = errors.Join(err, b.ImplicitReturn.Accept(r))
	}

	r.PopScope() // block scope
	return err
}

func (r *Resolver) VisitCondition(c *ast.Condition) error {
	var err error

	err = errors.Join(err, c.Condition.Accept(r))
	err = errors.Join(err, c.Body.Accept(r))
	if c.Else != nil {
		err = errors.Join(err, c.Else.Accept(r))
	}

	return err
}

func (r *Resolver) VisitAssignment(a *ast.Assignment) error {
	return errors.Join(a.Identifier.Accept(r), a.Value.Accept(r))
}

func (r *Resolver) VisitType(t *ast.Type) error       { return nil }
func (r *Resolver) VisitLiteral(l *ast.Literal) error { return nil }
