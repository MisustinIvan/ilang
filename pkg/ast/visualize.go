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
	fmt.Fprintf(w, `  %s [label="Program",color="lightblue"]`+"\n", rootID)

	for _, decl := range prog.Declarations {
		writeFunctionDeclaration(decl, rootID, w)
	}

	fmt.Fprintln(w, "}")
}

func writeFunctionDeclaration(fd FunctionDeclaration, parent string, w io.Writer) {
	id := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Function Declaration",color="red"]`+"\n", id)
	fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)

	nameID := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Name: %s"]`+"\n", nameID, fd.Name.Value)
	fmt.Fprintf(w, `  %s -> %s`+"\n", id, nameID)

	returnTypeID := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Return Type: %s"]`+"\n", returnTypeID, fd.Type.String())
	fmt.Fprintf(w, `  %s -> %s`+"\n", id, returnTypeID)

	paramsID := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Parameters",color="lightgreen"]`+"\n", paramsID)
	fmt.Fprintf(w, `  %s -> %s`+"\n", id, paramsID)
	for _, p := range fd.Parameters {
		paramID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Parameter: %s %s"]`+"\n", paramID, p.Type.String(), p.Name.Value)
		fmt.Fprintf(w, `  %s -> %s`+"\n", paramsID, paramID)
	}

	bodyID := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Body"]`+"\n", bodyID)
	fmt.Fprintf(w, `  %s -> %s`+"\n", id, bodyID)

	writeBlockExpression(fd.Body, bodyID, w)

}

func writeBlockExpression(expr BlockExpression, parent string, w io.Writer) {
	id := nextNodeID()
	fmt.Fprintf(w, `  %s [label="BlockExpression",color="purple"]`+"\n", id)
	fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)

	for _, expr := range expr.Body {
		writeExpression(expr, id, w)
	}

	if expr.ImplicitReturnExpression != nil {
		retID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Implicit Return Expression",color="lightgreen"]`+"\n", retID)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, retID)
		writeExpression(expr.ImplicitReturnExpression, retID, w)
	}
}

func writeExpression(expr Expression, parent string, w io.Writer) {
	switch e := expr.(type) {
	case *BindExpression:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="BindExpression",color="orange"]`+"\n", id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)

		leftID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Left: %s"]`+"\n", leftID, e.Left.Value)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, leftID)

		if e.Left.LookupValue != "" {
			lookupID := nextNodeID()
			fmt.Fprintf(w, `  %s [label="Lookup Value: %s"]`+"\n", lookupID, e.Left.LookupValue)
			fmt.Fprintf(w, `  %s -> %s`+"\n", leftID, lookupID)
		}

		typeID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Type: %s"]`+"\n", typeID, e.Type.String())
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, typeID)

		rightID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Right"]`+"\n", rightID)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, rightID)
		writeExpression(e.Right, rightID, w)
	case *ReturnExpression:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="ReturnExpression",color="red"]`+"\n", id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
		writeExpression(e.Value, id, w)
	case *AssignmentExpression:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="AssignmentExpression",color="green"]`+"\n", id)
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
		fmt.Fprintf(w, `  %s [label="BinaryExpression",color="lightblue"]`+"\n", id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
		writeExpression(e.Left, id, w)
		opID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Operator: %v"]`+"\n", opID, e.Operator)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, opID)
		writeExpression(e.Right, id, w)
	case *Literal:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Literal",color="green"]`+"\n", id)
		typeID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Type: %s"]`+"\n", typeID, e.Type.String())
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, typeID)
		valID := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Value: %s"]`+"\n", valID, escape(e.Value))
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, valID)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
	case *Identifier:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Identifier: %s",color="lightblue"]`+"\n", id, e.Value)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
		if e.LookupValue != "" {
			lookup_id := nextNodeID()
			fmt.Fprintf(w, `  %s [label="Lookup Value: %s"]`+"\n", lookup_id, e.LookupValue)
			fmt.Fprintf(w, `  %s -> %s`+"\n", id, lookup_id)
		}
	case *FunctionCall:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="FunctionCall",color="red"]`+"\n", id)
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
		fmt.Fprintf(w, `  %s [label="SeparatedExpression",color="lightgreen"]`+"\n", id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
		writeExpression(e.Value, id, w)
	case *ConditionalExpression:
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
		if len(e.ElseBody.Body) != 0 || e.ElseBody.ImplicitReturnExpression != nil {
			else_body_id := nextNodeID()
			fmt.Fprintf(w, `  %s [label="ElseBody"]`+"\n", else_body_id)
			fmt.Fprintf(w, `  %s -> %s`+"\n", id, else_body_id)
			writeBlockExpression(e.ElseBody, else_body_id, w)
		}
	case *BlockExpression:
		writeBlockExpression(*e, parent, w)
	default:
		id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Unknown Expression: %v",color="red"]`+"\n", id, e)
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
	}
}

func escape(val string) string {
	return strings.ReplaceAll(val, "\"", "\\\"")
}
