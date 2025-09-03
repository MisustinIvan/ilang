// The name_resolver package contains the definition for the NameResolver struct
// which implements ast.AstVisitor and uses the visitor pattern to resolve names
// of identifiers in the AST provided in the form of ast.Program.
package name_resolver

import (
	"errors"
	"fmt"
	"lang/pkg/ast"
	"lang/pkg/lexer"
)

// NameResolutionError represents an error that happended during the name
// resolution step along with its position.
type NameResolutionError struct {
	Message  string
	Position lexer.Position
}

func (e NameResolutionError) Error() string {
	return fmt.Sprintf("%s NameResolutionError: %s", e.Position.String(), e.Message)
}

func nameResolutionError(m string, p lexer.Position) NameResolutionError {
	return NameResolutionError{
		Message:  m,
		Position: p,
	}
}

type scope struct {
	parent *scope
	locals map[string]*ast.IdentifierExpression
}

type NameResolver struct {
	program *ast.Program
	scope   *scope
}

func NewNameResolver(program *ast.Program) NameResolver {
	r := NameResolver{
		program: program,
	}
	r.PushScope()
	return r
}

func (r *NameResolver) ResolveNames() (*ast.Program, error) {
	return r.program, r.program.Accept(r)
}

// PushScope creates a new empty scope with the current scope as the parent.
func (r *NameResolver) PushScope() {
	r.scope = &scope{
		parent: r.scope,
		locals: make(map[string]*ast.IdentifierExpression),
	}
}

// PopScope discards the innermost scope if it is not the root scope, in that
// case it simply does nothing.
func (r *NameResolver) PopScope() {
	if r.scope != nil {
		if r.scope.parent != nil {
			r.scope = r.scope.parent
		}
	}
}

// Lookup tries to find an identifier in the scopes starting from the innermost
// scope going upwards.
func (r *NameResolver) Lookup(id string) *ast.IdentifierExpression {
	for scope := r.scope; scope != nil; scope = scope.parent {
		if val, ok := scope.locals[id]; ok {
			return val
		}
	}
	return nil
}

// Declare declares a new local name from the identifier. In case the name is
// already declared Declare returns an error reporting the details. The Declare
// function allow for name shadowing and so it checks for conflicts only in the
// innermost scope.
func (r *NameResolver) Declare(id *ast.IdentifierExpression) error {
	if val, ok := r.scope.locals[id.Value]; ok {
		return nameResolutionError(fmt.Sprintf("identifier \"%s\" already declared in this scope at %s", id.Value, val.Position.String()), id.Position)
	}
	r.scope.locals[id.Value] = id
	return nil
}

// implementing ast.AstVisitor

func (r *NameResolver) VisitProgram(p *ast.Program) error {
	var errs []error
	for _, decl := range p.ExternalDeclarations {
		errs = append(errs, decl.Accept(r))
	}
	for _, decl := range p.Declarations {
		errs = append(errs, decl.Accept(r))
	}
	return errors.Join(errs...)
}

func (r *NameResolver) VisitFunctionDeclaration(d *ast.FunctionDeclaration) error {
	var errs []error
	if err := r.Declare(d.Identifier); err != nil {
		errs = append(errs, err)
	}
	r.PushScope()
	defer r.PopScope()
	for _, param := range d.Parameters {
		if err := param.Accept(r); err != nil {
			errs = append(errs, err)
		}
	}
	if err := d.Body.Accept(r); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (r *NameResolver) VisitParameterDefinition(d *ast.ParameterDefinition) error {
	return r.Declare(d.Name)

}

func (r *NameResolver) VisitExternalFunctionDeclaration(d *ast.ExternalFunctionDeclaration) error {
	var errs []error
	if err := r.Declare(d.Identifier); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (r *NameResolver) VisitBind(e *ast.BindExpression) error {
	var errs []error
	if err := e.Value.Accept(r); err != nil {
		errs = append(errs, err)
	}
	if err := r.Declare(e.Identifier); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (r *NameResolver) VisitReturn(e *ast.ReturnExpression) error {
	return e.Value.Accept(r)
}

func (r *NameResolver) VisitBinary(e *ast.BinaryExpression) error {
	var errs []error
	if err := e.Left.Accept(r); err != nil {
		errs = append(errs, err)
	}
	if err := e.Right.Accept(r); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (r *NameResolver) VisitLiteral(e *ast.LiteralExpression) error {
	return nil
}

func (r *NameResolver) VisitIdentifier(e *ast.IdentifierExpression) error {
	decl := r.Lookup(e.Value)
	if decl == nil {
		return nameResolutionError(fmt.Sprintf("undeclared identifier: %s", e.Value), e.Position)
	}
	e.Resolved = decl
	return nil
}

func (r *NameResolver) VisitCall(e *ast.CallExpression) error {
	var errs []error
	if err := e.Identifier.Accept(r); err != nil {
		errs = append(errs, err)
	}
	for _, param := range e.Params {
		if err := param.Accept(r); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (r *NameResolver) VisitBlock(e *ast.BlockExpression) error {
	var errs []error
	r.PushScope()
	defer r.PopScope()

	for _, expr := range e.Body {
		if err := expr.Accept(r); err != nil {
			errs = append(errs, err)
		}
	}

	if e.ImplicitReturn != nil {
		errs = append(errs, e.ImplicitReturn.Accept(r))
	}

	return errors.Join(errs...)
}

func (r *NameResolver) VisitSeparated(e *ast.SeparatedExpression) error {
	return e.Body.Accept(r)
}

func (r *NameResolver) VisitUnary(e *ast.UnaryExpression) error {
	return e.Value.Accept(r)
}

func (r *NameResolver) VisitConditional(e *ast.ConditionalExpression) error {
	var errs []error
	if err := e.Condition.Accept(r); err != nil {
		errs = append(errs, err)
	}
	if err := e.IfBody.Accept(r); err != nil {
		errs = append(errs, err)
	}
	if e.ElseBody != nil {
		if err := e.ElseBody.Accept(r); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (r *NameResolver) VisitAssignment(e *ast.AssignmentExpression) error {
	var errs []error
	if err := e.Identifier.Accept(r); err != nil {
		errs = append(errs, err)
	}
	if err := e.Value.Accept(r); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}
