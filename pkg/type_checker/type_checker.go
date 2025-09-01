// The type_checker package implements a type checker that traverses the program
// ast using the visitor pattern, reporting any type mismatches along the way.
package type_checker

import (
	"errors"
	"fmt"
	"lang/pkg/ast"
	"lang/pkg/lexer"
)

type TypeChecker struct {
	program              *ast.Program
	function_return_type ast.Type
	functions            map[*ast.IdentifierExpression][]ast.ParameterDefinition
}

type TypeCheckError struct {
	Message  string
	Position lexer.Position
}

func (e TypeCheckError) Error() string {
	return fmt.Sprintf("%s TypeCheckError: %s", e.Position.String(), e.Message)
}

func typeCheckError(msg string, pos lexer.Position) TypeCheckError {
	return TypeCheckError{
		Message:  msg,
		Position: pos,
	}
}

func NewTypeChecker(program *ast.Program) *TypeChecker {
	return &TypeChecker{
		program:              program,
		function_return_type: ast.Undefined,
		functions:            make(map[*ast.IdentifierExpression][]ast.ParameterDefinition),
	}
}

func (c *TypeChecker) CheckTypes() (*ast.Program, error) {
	return c.program, c.program.Accept(c)
}

func (c *TypeChecker) VisitProgram(p *ast.Program) error {
	var errs []error

	for _, decl := range p.Declarations {
		errs = append(errs, decl.Accept(c))
	}

	return errors.Join(errs...)
}

func (c *TypeChecker) VisitFunctionDeclaration(d *ast.FunctionDeclaration) error {
	c.function_return_type = d.Type
	c.functions[d.Identifier] = d.Parameters
	return d.Body.Accept(c)
}

func (c *TypeChecker) VisitParameterDefinition(d *ast.ParameterDefinition) error {
	return nil
}

func (c *TypeChecker) VisitBind(e *ast.BindExpression) error {
	var errs []error
	errs = append(errs, e.Value.Accept(c))
	if e.Type != e.Value.GetType() {
		errs = append(errs, typeCheckError(fmt.Sprintf("type mismatch in bind expression: %s vs %s", e.Type.String(), e.Value.GetType().String()), e.Position))
	}
	return errors.Join(errs...)
}

func (c *TypeChecker) VisitReturn(e *ast.ReturnExpression) error {
	var errs []error
	errs = append(errs, e.Value.Accept(c))
	if e.Value.GetType() != c.function_return_type {
		errs = append(errs, typeCheckError(fmt.Sprintf("type mismatch in return expression: %s vs %s", e.Value.GetType().String(), c.function_return_type.String()), e.Position))
	}
	return errors.Join(errs...)
}

func (c *TypeChecker) VisitBinary(e *ast.BinaryExpression) error {
	var errs []error
	if e.Left.GetType() != e.Right.GetType() {
		errs = append(errs, typeCheckError(fmt.Sprintf("type mismatch in binary expression: %s vs %s", e.Left.GetType().String(), e.Right.GetType().String()), e.Position))
	}
	if !ast.BinaryOperatorApplies[e.Operator][e.Left.GetType()] {
		errs = append(errs, typeCheckError(fmt.Sprintf("operator %s does not apply to type %s", e.Operator.String(), e.Left.GetType().String()), e.Left.GetPosition()))
	}
	if !ast.BinaryOperatorApplies[e.Operator][e.Right.GetType()] {
		errs = append(errs, typeCheckError(fmt.Sprintf("operator %s does not apply to type %s", e.Operator.String(), e.Right.GetType().String()), e.Right.GetPosition()))
	}
	return errors.Join(errs...)
}

func (c *TypeChecker) VisitLiteral(e *ast.LiteralExpression) error {
	return nil
}

func (c *TypeChecker) VisitIdentifier(e *ast.IdentifierExpression) error {
	return nil
}

func (c *TypeChecker) VisitCall(e *ast.CallExpression) error {
	var errs []error
	if e.Identifier.Resolved == nil {
		return typeCheckError(fmt.Sprintf("undeclared function %s has unknown type", e.Identifier.Value), e.Position)
	}
	if params, ok := c.functions[e.Identifier.Resolved]; !ok {
		return typeCheckError(fmt.Sprintf("resolved function %s with unknown parameter types, should never happen", e.Identifier.Value), e.Position)
	} else {
		if len(e.Params) != len(params) {
			errs = append(errs, typeCheckError(fmt.Sprintf("function call with wrong amount of parameters, got %d, expecting %d", len(e.Params), len(params)), e.Position))
		}
		max_len := min(len(e.Params), len(params))

		for i := range max_len {
			if e.Params[i].GetType() != params[i].Name.GetType() {
				errs = append(errs, typeCheckError(fmt.Sprintf("wrong type of parameter %d, got %s, expecting %s", i, e.Params[i].GetType().String(), params[i].Name.GetType().String()), e.Position))
			}
		}
	}
	return errors.Join(errs...)
}

func (c *TypeChecker) VisitBlock(e *ast.BlockExpression) error {
	var errs []error
	for _, expr := range e.Body {
		errs = append(errs, expr.Accept(c))
	}
	if e.ImplicitReturn != nil {
		errs = append(errs, e.ImplicitReturn.Accept(c))
		// check correct return type
		if e.GetType() != e.ImplicitReturn.GetType() {
			errs = append(errs, typeCheckError(fmt.Sprintf("implicit return type mismatch, got %s, expected %s", e.ImplicitReturn.GetType().String(), e.GetType().String()), e.ImplicitReturn.GetPosition()))
		}
	}

	return errors.Join(errs...)
}

func (c *TypeChecker) VisitSeparated(e *ast.SeparatedExpression) error {
	return e.Body.Accept(c)
}

func (c *TypeChecker) VisitUnary(e *ast.UnaryExpression) error {
	var errs []error
	errs = append(errs, e.Value.Accept(c))
	if !ast.UnaryOperatorApplies[e.Operator][e.Value.GetType()] {
		errs = append(errs, typeCheckError(fmt.Sprintf("operator %s does not apply to type %s", e.Operator.String(), e.Value.GetType()), e.Position))
	}
	return errors.Join(errs...)
}

func (c *TypeChecker) VisitConditional(e *ast.ConditionalExpression) error {
	var errs []error
	errs = append(errs, e.Condition.Accept(c))
	if e.Condition.GetType() != ast.Boolean {
		errs = append(errs, typeCheckError(fmt.Sprintf("expected boolean condition value, got %s", e.Condition.GetType().String()), e.Position))
	}
	errs = append(errs, e.IfBody.Accept(c))
	if e.ElseBody != nil {
		errs = append(errs, e.ElseBody.Accept(c))
		if e.ElseBody.GetType() != e.IfBody.GetType() {
			errs = append(errs, typeCheckError(fmt.Sprintf("conditional branches have different types: %s vs %s", e.IfBody.GetType().String(), e.ElseBody.GetType().String()), e.Position))
		}
	}
	return errors.Join(errs...)
}

func (c *TypeChecker) VisitAssignment(e *ast.AssignmentExpression) error {
	var errs []error
	errs = append(errs, e.Value.Accept(c))
	if e.Value.GetType() != e.Identifier.GetType() {
		errs = append(errs, typeCheckError(fmt.Sprintf("type mismatch in assignment, expected %s, got %s", e.Identifier.GetType().String(), e.Value.GetType().String()), e.Position))
	}
	return errors.Join(errs...)
}
