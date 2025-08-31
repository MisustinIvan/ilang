package ast_visualizer

import (
	"fmt"
	"io"
	"lang/pkg/ast"
	"strings"
)

// Contains functions that allow converting an ast.Program to a graphviz graph

var nodeCounter int

func nextNodeID() string {
	nodeCounter++
	return fmt.Sprintf("node%d", nodeCounter)
}

func ExportASTToGraphviz(prog *ast.Program, w io.Writer) {
	fmt.Fprintln(w, "digraph AST {")
	fmt.Fprintln(w, `  node [shape=box];`)
	rootID := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Program",color="lightblue"]`+"\n", rootID)

	for _, decl := range prog.Declarations {
		writeFunctionDeclaration(decl, rootID, w)
	}

	fmt.Fprintln(w, "}")
}

func writeFunctionDeclaration(fd ast.FunctionDeclaration, parent string, w io.Writer) {
	id := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Function Declaration",color="red"]`+"\n", id)
	fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)

	nameID := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Name: %s"]`+"\n", nameID, fd.Identifier.Value)
	fmt.Fprintf(w, `  %s -> %s`+"\n", id, nameID)

	returnTypeID := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Return Type Name: %s"]`+"\n", returnTypeID, fd.TypeName.Value)
	fmt.Fprintf(w, `  %s -> %s`+"\n", id, returnTypeID)

	paramsID := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Parameters",color="lightgreen"]`+"\n", paramsID)
	fmt.Fprintf(w, `  %s -> %s`+"\n", id, paramsID)
	for _, p := range fd.Parameters {
		paramID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Parameter: %s %s"]`+"\n", paramID, p.TypeName.Value, p.Name.Value)
		fmt.Fprintf(w, `  %s -> %s`+"\n", paramsID, paramID)
	}

	bodyID := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Body"]`+"\n", bodyID)
	fmt.Fprintf(w, `  %s -> %s`+"\n", id, bodyID)

	writeBlockExpression(fd.Body, bodyID, w)

}

func writeBlockExpression(expr *ast.BlockExpression, parent string, w io.Writer) {
	id := nextNodeID()
	fmt.Fprintf(w, `  %s [label="BlockExpression",color="purple"]`+"\n", id)
	fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)

	if expr == nil {
		return
	}

	for _, expr := range expr.Body {
		writeExpression(expr, id, w)
	}

	if expr.ImplicitReturn != nil {
		retID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Implicit Return Expression",color="lightgreen"]`+"\n", retID)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, retID)
		writeExpression(expr.ImplicitReturn, retID, w)
	}
}

func writeExpression(expr ast.Expression, parent string, w io.Writer) {
	switch e := expr.(type) {
	case *ast.BindExpression:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="BindExpression",color="orange"]`+"\n", id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)

		leftID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Identifier: %s"]`+"\n", leftID, e.Identifier.Value)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, leftID)

		typeID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="TypeName: %s"]`+"\n", typeID, e.TypeName.Value)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, typeID)

		rightID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Value"]`+"\n", rightID)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, rightID)
		writeExpression(e.Value, rightID, w)
	case *ast.ReturnExpression:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="ReturnExpression",color="red"]`+"\n", id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
		writeExpression(e.Value, id, w)
	case *ast.AssignmentExpression:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="AssignmentExpression",color="green"]`+"\n", id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
		leftID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Identifier: %s"]`+"\n", leftID, e.Identifier.Value)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, leftID)
		rightID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Value"]`+"\n", rightID)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, rightID)
		writeExpression(e.Value, rightID, w)
	case *ast.BinaryExpression:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="BinaryExpression",color="lightblue"]`+"\n", id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
		writeExpression(e.Left, id, w)
		opID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Operator: %v"]`+"\n", opID, e.Operator)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, opID)
		writeExpression(e.Right, id, w)
	case *ast.LiteralExpression:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="LiteralExpression",color="green"]`+"\n", id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
		valID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Value: %s"]`+"\n", valID, escape(e.Value))
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, valID)

	case *ast.IdentifierExpression:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="IdentifierExpression: %s",color="lightblue"]`+"\n", id, e.Value)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
	case *ast.CallExpression:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="CallExpression",color="red"]`+"\n", id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
		fnID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Identifier: %s"]`+"\n", fnID, e.Identifier.Value)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, fnID)
		argsID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Args"]`+"\n", argsID)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, argsID)
		for _, arg := range e.Params {
			writeExpression(arg, argsID, w)
		}
	case *ast.SeparatedExpression:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="SeparatedExpression",color="lightgreen"]`+"\n", id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
		writeExpression(e.Body, id, w)
	case *ast.ConditionalExpression:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="ConditionalExpression",color="red"]`+"\n", id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
		condition_id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Condition"]`+"\n", condition_id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, condition_id)
		writeExpression(e.Condition, condition_id, w)
		if_body_id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="IfBody"]`+"\n", if_body_id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, if_body_id)
		writeBlockExpression(e.IfBody, if_body_id, w)
		if e.ElseBody != nil {
			else_body_id := nextNodeID()
			fmt.Fprintf(w, `  %s [label="ElseBody"]`+"\n", else_body_id)
			fmt.Fprintf(w, `  %s -> %s`+"\n", id, else_body_id)
			writeBlockExpression(e.ElseBody, else_body_id, w)
		}
	case *ast.BlockExpression:
		writeBlockExpression(e, parent, w)
	case *ast.UnaryExpression:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Unary Expression",color="red"]`+"\n", id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
		operator_id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Operator: %s"]`+"\n", operator_id, e.Operator.String())
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, operator_id)
		value_id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Value""]`+"\n", value_id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, value_id)
		writeExpression(e.Value, value_id, w)

	default:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Unknown Expression: %v",color="red"]`+"\n", id, e)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
	}
}

func escape(val string) string {
	return strings.ReplaceAll(val, "\"", "\\\"")
}
