package ast

import (
	"fmt"
	"io"
	"strings"
)

var nodeCounter int

func nextNodeID() string {
	nodeCounter++
	return fmt.Sprintf("node%d", nodeCounter)
}

func ExportASTToGraphviz(prog *Prog, w io.Writer) {
	fmt.Fprintln(w, "digraph AST {")
	fmt.Fprintln(w, `  node [shape=box];`)
	rootID := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Program"]`+"\n", rootID)

	for _, decl := range prog.Declarations {
		writeFunctionDeclaration(decl, rootID, w)
	}

	fmt.Fprintln(w, "}")
}

func writeFunctionDeclaration(fd FunctionDeclaration, parent string, w io.Writer) {
	id := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Function Declaration"]`+"\n", id)
	fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)

	nameID := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Name: %s"]`+"\n", nameID, fd.Name.Value)
	fmt.Fprintf(w, `  %s -> %s`+"\n", id, nameID)

	returnTypeID := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Return Type: %s"]`+"\n", returnTypeID, fd.Type.String())
	fmt.Fprintf(w, `  %s -> %s`+"\n", id, returnTypeID)

	paramsID := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Parameters"]`+"\n", paramsID)
	fmt.Fprintf(w, `  %s -> %s`+"\n", id, paramsID)
	for _, p := range fd.Parameters {
		paramID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Parameter: %s %s"]`+"\n", paramID, p.Type.String(), p.Name.Value)
		fmt.Fprintf(w, `  %s -> %s`+"\n", paramsID, paramID)
	}

	bodyID := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Body"]`+"\n", bodyID)
	fmt.Fprintf(w, `  %s -> %s`+"\n", id, bodyID)

	for _, expr := range fd.Body.Body {
		writeExpression(expr, bodyID, w)
	}

	if fd.Body.ReturnExpression != nil {
		retID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Return Expression"]`+"\n", retID)
		fmt.Fprintf(w, `  %s -> %s`+"\n", bodyID, retID)
		writeExpression(fd.Body.ReturnExpression, retID, w)
	}
}

func writeExpression(expr Expression, parent string, w io.Writer) {
	switch e := expr.(type) {
	case *BindExpression:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="BindExpression"]`+"\n", id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)

		leftID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Left: %s"]`+"\n", leftID, e.Left.Value)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, leftID)

		typeID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Type: %s"]`+"\n", typeID, e.Type.String())
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, typeID)

		rightID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Right"]`+"\n", rightID)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, rightID)
		writeExpression(e.Right, rightID, w)
	case *ReturnExpression:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="ReturnExpression"]`+"\n", id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
		writeExpression(e.Value, id, w)
	case *AssignmentExpression:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="AssignmentExpression"]`+"\n", id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
		leftID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Left: %s"]`+"\n", leftID, e.Left.Value)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, leftID)
		rightID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Right"]`+"\n", rightID)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, rightID)
		writeExpression(e.Right, rightID, w)
	case *BinaryExpression:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="BinaryExpression"]`+"\n", id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
		writeExpression(e.Left, id, w)
		opID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Operator: %v"]`+"\n", opID, e.Operator)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, opID)
		writeExpression(e.Right, id, w)
	case *Literal:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Literal"]`+"\n", id)
		typeID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Type: %s"]`+"\n", typeID, e.Type.String())
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, typeID)
		valID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Value: %s"]`+"\n", valID, escape(e.Value))
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, valID)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
	case *Identifier:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Identifier: %s"]`+"\n", id, e.Value)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
	case *FunctionCall:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="FunctionCall"]`+"\n", id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
		fnID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Function: %s"]`+"\n", fnID, e.Function.Value)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, fnID)
		argsID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Args"]`+"\n", argsID)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, argsID)
		for _, arg := range e.Params {
			writeExpression(arg, argsID, w)
		}
	case *SeparatedExpression:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="SeparatedExpression"]`+"\n", id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
		writeExpression(e.Value, id, w)
	default:
		fmt.Fprintf(w, `  // Unknown expression type`+"\n")
	}
}

func escape(val string) string {
	return strings.ReplaceAll(val, "\"", "\\\"")
}
