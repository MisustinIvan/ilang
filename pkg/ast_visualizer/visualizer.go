// Contains functions that allow converting an ast.Program to a graphviz graph
package ast_visualizer

import (
	"fmt"
	"io"
	"lang/pkg/ast"
	"strings"
)

var nodeCounter int

func nextNodeID() string {
	nodeCounter++
	return fmt.Sprintf("node%d", nodeCounter)
}

func escape(val string) string {
	return strings.ReplaceAll(val, "\"", "\\\"")
}

func writeNode(w io.Writer, parent, label, color string, args ...any) string {
	id := nextNodeID()
	label = fmt.Sprintf(label, args...)
	if color != "" {
		fmt.Fprintf(w, `  %s [label="%s",color="%s"]`+"\n", id, escape(label), color)
	} else {
		fmt.Fprintf(w, `  %s [label="%s"]`+"\n", id, escape(label))
	}
	if parent != "" {
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
	}
	return id
}

func ExportASTToGraphviz(prog *ast.Program, w io.Writer) {
	fmt.Fprintln(w, "digraph AST {")
	fmt.Fprintln(w, `  node [shape=box];`)
	rootID := writeNode(w, "", "Program", "lightblue")

	for _, decl := range prog.ExternalDeclarations {
		writeExternalFunctionDeclaration(decl, rootID, w)
	}

	for _, decl := range prog.Declarations {
		writeFunctionDeclaration(decl, rootID, w)
	}

	fmt.Fprintln(w, "}")
}

func writeFunctionDeclaration(fd *ast.FunctionDeclaration, parent string, w io.Writer) {
	id := writeNode(w, parent, "Function Declaration", "red")
	writeNode(w, id, "Name: %s", "", fd.Identifier.Value)
	writeNode(w, id, "Return Type Name: %s", "", fd.TypeName.Value)
	writeNode(w, id, "Return Type: %s", "", fd.Type.String())

	paramsID := writeNode(w, id, "Parameters", "lightgreen")
	for _, p := range fd.Parameters {
		writeNode(w, paramsID, "Parameter: %s(%s) %s", "", p.TypeName.Value, p.Name.Type.String(), p.Name.Value)
	}

	bodyID := writeNode(w, id, "Body", "")
	writeBlockExpression(fd.Body, bodyID, w)
}

func writeExternalFunctionDeclaration(fd *ast.ExternalFunctionDeclaration, parent string, w io.Writer) {
	id := writeNode(w, parent, "External Function Declaration", "red")
	writeNode(w, id, "Name: %s", "", fd.Identifier.Value)
	writeNode(w, id, "Return Type Name: %s", "", fd.TypeName.Value)
	writeNode(w, id, "Return Type: %s", "", fd.Type.String())
	writeNode(w, id, "Return Type: %s", "", fd.Identifier.GetType().String())

	paramsID := writeNode(w, id, "Parameters", "lightgreen")
	for _, p := range fd.Parameters {
		writeNode(w, paramsID, "Parameter: %s(%s) %s", "", p.TypeName.Value, p.Name.Type.String(), p.Name.Value)
	}
}

func writeBlockExpression(expr *ast.BlockExpression, parent string, w io.Writer) {
	if expr == nil {
		return
	}

	id := writeNode(w, parent, "BlockExpression", "purple")
	writeNode(w, id, "Type: %s", "", expr.Type.String())

	for _, e := range expr.Body {
		writeExpression(e, id, w)
	}

	if expr.ImplicitReturn != nil {
		retID := writeNode(w, id, "Implicit Return Expression", "lightgreen")
		writeExpression(expr.ImplicitReturn, retID, w)
	}
}

func writeExpression(expr ast.Expression, parent string, w io.Writer) {
	switch e := expr.(type) {
	case *ast.BindExpression:
		id := writeNode(w, parent, "BindExpression", "orange")
		writeNode(w, id, "Identifier: %s", "", e.Identifier.Value)
		writeNode(w, id, "TypeName: %s", "", e.TypeName.Value)
		writeNode(w, id, "Type: %s", "", e.Type.String())
		valID := writeNode(w, id, "Value", "")
		writeExpression(e.Value, valID, w)

	case *ast.ReturnExpression:
		id := writeNode(w, parent, "ReturnExpression", "red")
		writeNode(w, id, "Type: %s", "", e.GetType().String())
		writeExpression(e.Value, id, w)

	case *ast.AssignmentExpression:
		id := writeNode(w, parent, "AssignmentExpression", "green")
		writeNode(w, id, "Identifier: %s", "", e.Identifier.Value)
		writeNode(w, id, "Type: %s", "", e.Identifier.Type.String())
		rightID := writeNode(w, id, "Value", "")
		writeNode(w, rightID, "Type: %s", "", e.Value.GetType().String())
		writeExpression(e.Value, rightID, w)

	case *ast.BinaryExpression:
		id := writeNode(w, parent, "BinaryExpression", "lightblue")
		writeNode(w, id, "Type: %s", "", e.Type.String())
		writeExpression(e.Left, id, w)
		writeNode(w, id, "Operator: %v", "", e.Operator)
		writeExpression(e.Right, id, w)

	case *ast.LiteralExpression:
		id := writeNode(w, parent, "LiteralExpression", "green")
		writeNode(w, id, "Value: %s", "", e.Value)
		writeNode(w, id, "Type: %s", "", e.Type.String())

	case *ast.IdentifierExpression:
		id := writeNode(w, parent, "IdentifierExpression: %s", "lightblue", e.Value)
		writeNode(w, id, "Type: %s", "", e.Type.String())

	case *ast.CallExpression:
		id := writeNode(w, parent, "CallExpression", "red")
		writeNode(w, id, "Identifier: %s", "", e.Identifier.Value)
		writeNode(w, id, "Type: %s", "", e.Identifier.Type.String())
		argsID := writeNode(w, id, "Args", "")
		for _, arg := range e.Params {
			writeExpression(arg, argsID, w)
		}

	case *ast.SeparatedExpression:
		id := writeNode(w, parent, "SeparatedExpression", "lightgreen")
		writeExpression(e.Body, id, w)

	case *ast.ConditionalExpression:
		id := writeNode(w, parent, "ConditionalExpression", "red")
		writeNode(w, id, "Type: %s", "", e.GetType().String())
		condID := writeNode(w, id, "Condition", "")
		writeExpression(e.Condition, condID, w)
		ifID := writeNode(w, id, "IfBody", "")
		writeExpression(e.IfBody, ifID, w)
		if e.ElseBody != nil {
			elseID := writeNode(w, id, "ElseBody", "")
			writeExpression(e.ElseBody, elseID, w)
		}

	case *ast.BlockExpression:
		writeBlockExpression(e, parent, w)

	case *ast.UnaryExpression:
		id := writeNode(w, parent, "UnaryExpression", "red")
		writeNode(w, id, "Operator: %s", "", e.Operator.String())
		valID := writeNode(w, id, "Value", "")
		writeExpression(e.Value, valID, w)

	default:
		writeNode(w, parent, "Unknown Expression: %v", "red", e)
	}
}
