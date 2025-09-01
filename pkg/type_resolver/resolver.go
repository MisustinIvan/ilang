// The type_resolver package implements a type repolver traverses the program ast
// and resolves types of expressions, reporting any errors along the way.
package type_resolver

import (
	"errors"
	"fmt"
	"lang/pkg/ast"
	"lang/pkg/lexer"
)

type TypeResolver struct {
	program *ast.Program
}

func NewTypeResolver(prog *ast.Program) TypeResolver {
	return TypeResolver{
		program: prog,
	}
}

type TypeResolutionError struct {
	Message  string
	Position lexer.Position
}

func (e TypeResolutionError) Error() string {
	return fmt.Sprintf("%s TypeResolutionError: %s", e.Position.String(), e.Message)
}

func typeResolutionError(msg string, pos lexer.Position) TypeResolutionError {
	return TypeResolutionError{
		Message:  msg,
		Position: pos,
	}
}

func (t *TypeResolver) ResolveTypes() (*ast.Program, error) {
	return t.program, t.program.Accept(t)
}

// implementing ast.AstVisitor

func (t *TypeResolver) VisitProgram(p *ast.Program) error {
	var errs []error
	for _, decl := range p.ExternalDeclarations {
		errs = append(errs, decl.Accept(t))
	}
	for _, decl := range p.Declarations {
		errs = append(errs, decl.Accept(t))
	}
	return errors.Join(errs...)
}

func (t *TypeResolver) VisitFunctionDeclaration(d *ast.FunctionDeclaration) error {
	var errs []error

	return_type, err := ast.ParseType(d.TypeName.Value)
	if err != nil {
		errs = append(errs, err)
	}
	d.Identifier.SetType(return_type)
	d.Type = return_type

	for _, param := range d.Parameters {
		if err := param.Accept(t); err != nil {
			errs = append(errs, err)
		}
	}

	d.Body.SetType(return_type)

	if err := d.Body.Accept(t); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (t *TypeResolver) VisitParameterDefinition(d *ast.ParameterDefinition) error {
	parameter_type, err := ast.ParseType(d.TypeName.Value)
	d.Name.SetType(parameter_type)
	return err
}

func (t *TypeResolver) VisitExternalFunctionDeclaration(d *ast.ExternalFunctionDeclaration) error {
	var errs []error

	return_type, err := ast.ParseType(d.TypeName.Value)
	if err != nil {
		errs = append(errs, err)
	}
	d.Identifier.SetType(return_type)
	d.Type = return_type

	for _, param := range d.Parameters {
		if err := param.Accept(t); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (t *TypeResolver) VisitBind(e *ast.BindExpression) error {
	var errs []error
	variable_type, err := ast.ParseType(e.TypeName.Value)
	if err != nil {
		errs = append(errs, err)
	}
	e.SetType(variable_type)

	if err := e.Value.Accept(t); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (t *TypeResolver) VisitReturn(e *ast.ReturnExpression) error {
	err := e.Value.Accept(t)
	e.SetType(e.Value.GetType())
	return err
}

func (t *TypeResolver) VisitBinary(e *ast.BinaryExpression) error {
	var errs []error

	if err := e.Left.Accept(t); err != nil {
		errs = append(errs, err)
	}
	e.SetType(e.Left.GetType())

	if err := e.Right.Accept(t); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (t *TypeResolver) VisitLiteral(e *ast.LiteralExpression) error {
	literal_type, err := ast.LiteralType(e.Value)
	e.SetType(literal_type)
	return err
}

func (t *TypeResolver) VisitIdentifier(e *ast.IdentifierExpression) error {
	if e.Resolved == nil {
		e.SetType(ast.Undefined)
		return typeResolutionError(fmt.Sprintf("unresolved identifier \"%s\" has undefined type", e.Value), e.Position)
	}
	e.SetType(e.Resolved.Type)
	return nil
}

func (t *TypeResolver) VisitCall(e *ast.CallExpression) error {
	var errs []error

	if err := e.Identifier.Accept(t); err != nil {
		errs = append(errs, err)
	}
	e.SetType(e.Identifier.Type)

	for _, param := range e.Params {
		if err := param.Accept(t); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (t *TypeResolver) VisitBlock(e *ast.BlockExpression) error {
	var errs []error

	for _, expr := range e.Body {
		if err := expr.Accept(t); err != nil {
			errs = append(errs, err)
		}
	}

	if e.ImplicitReturn != nil {
		errs = append(errs, e.ImplicitReturn.Accept(t))
	}

	// If this block expression is a function body, the function will set
	// the expected type, if it is undefined, we have to set it here.
	if e.GetType() == ast.Undefined {
		if e.ImplicitReturn != nil {
			e.SetType(e.ImplicitReturn.GetType())
		} else {
			e.SetType(ast.Unit)
		}
	}

	return errors.Join(errs...)
}

func (t *TypeResolver) VisitSeparated(e *ast.SeparatedExpression) error {
	err := e.Body.Accept(t)
	e.SetType(e.Body.GetType())
	return err
}

func (t *TypeResolver) VisitUnary(e *ast.UnaryExpression) error {
	err := e.Value.Accept(t)
	e.SetType(e.Value.GetType())
	return err
}

func (t *TypeResolver) VisitConditional(e *ast.ConditionalExpression) error {
	var errs []error
	if err := e.Condition.Accept(t); err != nil {
		errs = append(errs, err)
	}
	if err := e.IfBody.Accept(t); err != nil {
		errs = append(errs, err)
	}

	e.SetType(e.IfBody.GetType())

	if e.ElseBody != nil {
		if err := e.ElseBody.Accept(t); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (t *TypeResolver) VisitAssignment(e *ast.AssignmentExpression) error {
	var errs []error

	if err := e.Identifier.Accept(t); err != nil {
		errs = append(errs, err)
	}

	e.SetType(e.Identifier.GetType())

	if err := e.Value.Accept(t); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
