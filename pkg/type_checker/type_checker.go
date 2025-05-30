package type_checker

import (
	"fmt"
	"lang/pkg/ast"
	"lang/pkg/lexer"
)

type TypeError struct {
	Position lexer.TokenPosition
	Message  string
}

func (e TypeError) Error() string {
	return fmt.Sprintf("%s: %s", e.Position.String(), e.Message)
}

func typeError(msg string, pos lexer.TokenPosition) TypeError {
	return TypeError{
		Position: pos,
		Message:  msg,
	}
}

type TypeChecker struct {
	program ast.Prog
	errors  []TypeError
	decls   map[string]ast.FunctionDeclaration
}

func NewTypeChecker(in ast.Prog) TypeChecker {
	decls := map[string]ast.FunctionDeclaration{}
	for _, d := range in.Declarations {
		decls[d.Name.Value] = d
	}
	return TypeChecker{
		program: in,
		errors:  []TypeError{},
		decls:   decls,
	}
}

func (c *TypeChecker) CheckTypes() bool {
	for _, decl := range c.program.Declarations {
		errs := c.CheckExpressionType(&decl.Body)
		c.errors = append(c.errors, errs...)

		if decl.Body.Type != decl.Type {
			c.errors = append(c.errors, TypeError{
				Message:  fmt.Sprintf("type of body (%s) does not match declared return type (%s) in function %s", decl.Body.Type.String(), decl.Type.String(), decl.Name.Value),
				Position: decl.Position,
			})
		}
	}
	return len(c.errors) == 0
}

func (c *TypeChecker) Report() {
	for _, err := range c.errors {
		fmt.Printf("Error: %s\n", err.Error())
	}
}

func (c *TypeChecker) CheckExpressionType(e ast.Expression) []TypeError {
	switch e := e.(type) {
	case *ast.BlockExpression:
		ret_errs := []TypeError{}
		for _, expr := range e.Body {
			errs := c.CheckExpressionType(expr)
			if len(errs) != 0 {
				ret_errs = append(ret_errs, errs...)
			}
		}

		if e.ImplicitReturnExpression != nil {
			errs := c.CheckExpressionType(e.ImplicitReturnExpression)
			ret_errs = append(ret_errs, errs...)
		}
		return ret_errs
	case *ast.Literal:
		return nil
	case *ast.Identifier:
		return nil
	case *ast.ReturnExpression:
		return c.CheckExpressionType(e.Value)
	case *ast.UnaryExpression:
		errs := c.CheckExpressionType(e.Value)
		t := e.Expression_i.Type
		op := e.Operator
		if !ast.UnaryOperatorApplies[op][t] {
			errs = append(errs, typeError(fmt.Sprintf("operator %s not compatible with type %s", op.String(), t.String()), e.TokenPosition))
		}
		return errs
	case *ast.BindExpression:
		lt := e.Left.Type
		errs := c.CheckExpressionType(e.Right)
		if lt != e.Right.GetType() {
			errs = append(errs, typeError(fmt.Sprintf("can't bind value of type %s to identifier of type %s", e.Right.GetType().String(), lt.String()), e.TokenPosition))
		}
		return errs
	case *ast.AssignmentExpression:
		lt := e.Left.GetType()
		errs := c.CheckExpressionType(e.Right)
		if lt != e.Right.GetType() {
			errs = append(errs, typeError(fmt.Sprintf("invalid type on right side of assignment, expected %s, got %s", lt.String(), e.Right.GetType().String()), e.TokenPosition))
		}
		return errs
	case *ast.BinaryExpression:
		lt := e.Left.GetType()
		rt := e.Right.GetType()
		op := e.Operator
		errs := c.CheckExpressionType(e.Left)
		errs = append(errs, c.CheckExpressionType(e.Right)...)
		if lt != rt {
			errs = append(errs, typeError(fmt.Sprintf("incompatible types %s and %s in binary expression", lt.String(), rt.String()), e.TokenPosition))
		}
		if !ast.BinaryOperatorApplies[op][lt] {
			errs = append(errs, typeError(fmt.Sprintf("operator %s not compatible with type %s", op.String(), lt.String()), e.TokenPosition))
		}
		if !ast.BinaryOperatorApplies[op][rt] {
			errs = append(errs, typeError(fmt.Sprintf("operator %s not compatible with type %s", op.String(), rt.String()), e.TokenPosition))
		}
		return errs
	case *ast.SeparatedExpression:
		return c.CheckExpressionType(e.Value)
	case *ast.ConditionalExpression:
		errs := c.CheckExpressionType(e.Condition)
		errs = append(errs, c.CheckExpressionType(&e.IfBody)...)
		errs = append(errs, c.CheckExpressionType(&e.ElseBody)...)
		it := e.IfBody.GetType()
		et := e.ElseBody.GetType()
		if it != et {
			errs = append(errs, typeError(fmt.Sprintf("conditional branches dont return the same type: %s vs %s", it.String(), et.String()), e.TokenPosition))
		}
		return errs
	case *ast.ForExpression:
		errs := c.CheckExpressionType(e.Condition)
		errs = append(errs, c.CheckExpressionType(e.Body)...)
		return errs
	case *ast.BreakExpression:
		return nil
	case *ast.FunctionCall:
		d, ok := c.decls[e.Function.Value]
		if !ok {
			return []TypeError{typeError(fmt.Sprintf("call to undeclared function %s", e.Function.Value), e.TokenPosition)}
		}
		errs := []TypeError{}
		if len(d.Parameters) != len(e.Params) {
			errs = append(errs, typeError(fmt.Sprintf("not matching amount of parameters in function %s call", e.Function.Value), e.TokenPosition))
		}

		for i, arg := range e.Params {
			errs = append(errs, c.CheckExpressionType(arg)...)
			if i >= len(d.Parameters) {
				errs = append(errs, typeError(fmt.Sprintf("unexpected parameter"), e.TokenPosition))
			} else {
				at := arg.GetType()
				if at != d.Parameters[i].Type {
					errs = append(errs, typeError(fmt.Sprintf("invalid type of argument, expected %s got %s", d.Parameters[i].Type.String(), at.String()), e.TokenPosition))
				}
			}
		}
		return errs
	default:
		return nil
	}
}
