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
	root_id := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Program"]`+"\n", root_id)

	for _, decl := range prog.Declarations {
		writeFunctionDeclaration(decl, root_id, w)
	}

	fmt.Fprintln(w, "}")
}

func writeFunctionDeclaration(fd FunctionDeclaration, parent string, w io.Writer) {
	id := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Function Declaration"]`+"\n", id)
	name_id := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Name: %s"]`+"\n", name_id, fd.Name.Name)
	fmt.Fprintf(w, `  %s -> %s`+"\n", id, name_id)

	return_type_id := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Return Type: %s"]`+"\n", return_type_id, fd.Type.String())
	fmt.Fprintf(w, `  %s -> %s`+"\n", id, return_type_id)

	params_id := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Parameters"]`+"\n", params_id)
	fmt.Fprintf(w, `  %s -> %s`+"\n", id, params_id)
	for _, p := range fd.ParameterTypes {
		n_id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Parameter: %s %s"]`+"\n", n_id, p.Type.String(), p.Name.Name)
		fmt.Fprintf(w, `  %s -> %s`+"\n", params_id, n_id)
	}

	if parent != "" && parent != "root" {
		fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)
	}

	body_id := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Body"]`+"\n", body_id)
	fmt.Fprintf(w, `  %s -> %s`+"\n", id, body_id)

	for _, stmt := range fd.Body {
		writeStatement(stmt, body_id, w)
	}
}

func writeStatement(stmt Statement, parent string, w io.Writer) {
	id := nextNodeID()
	type_id := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Statement"]`+"\n", id)
	fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)

	switch s := stmt.(type) {
	case *AssignmentStatement:
		fmt.Fprintf(w, `  %s [label="Type: Assignment"]`+"\n", type_id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, type_id)

		left_id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Left: %s"]`+"\n", left_id, s.Left.Name)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, left_id)

		right_id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Right"]`+"\n", right_id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, right_id)
		writeExpression(s.Right, right_id, w)
	case *FunctionCallStatement:
		fmt.Fprintf(w, `  %s [label="Type: Function Call"]`+"\n", type_id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, type_id)

		name_id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Function name: %s"]`+"\n", name_id, s.Function.Name)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, name_id)

		args_id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Args"]`+"\n", args_id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, args_id)

		for _, arg := range s.Args {
			writeExpression(arg, args_id, w)
		}
	case *VariableDeclarationStatement:
		fmt.Fprintf(w, `  %s [label="Type: Variable Declaration"]`+"\n", type_id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, type_id)

		var_type_id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Variable Type: %s"]`+"\n", var_type_id, s.Type.String())
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, var_type_id)

		left_id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Left: %s"]`+"\n", left_id, s.Left.Name)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, left_id)

		right_id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Right"]`+"\n", right_id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, right_id)
		writeExpression(s.Right, right_id, w)
	case *ReturnStatement:
		fmt.Fprintf(w, `  %s [label="Type: Return"]`+"\n", type_id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, type_id)

		value_id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Return Value"]`+"\n", value_id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, value_id)

		writeExpression(s.Value, value_id, w)
	}
}

func writeExpression(expr Expression, parent string, w io.Writer) {
	id := nextNodeID()
	fmt.Fprintf(w, `  %s [label="Expression"]`+"\n", id)
	fmt.Fprintf(w, `  %s -> %s`+"\n", parent, id)

	switch e := expr.(type) {
	case *CallExpression:
		type_id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Type: Function Call"]`+"\n", type_id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, type_id)

		name_id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Function name: %s"]`+"\n", name_id, e.Function.Name)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, name_id)

		args_id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Args"]`+"\n", args_id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, args_id)

		for _, arg := range e.Args {
			writeExpression(arg, args_id, w)
		}
	case *IdentifierExpression:
		type_id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Type: Identifier"]`+"\n", type_id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, type_id)

		value_id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Value: %s"]`+"\n", value_id, e.Identifier.Name)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, value_id)

	case *LiteralExpression:
		type_id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Type: Literal"]`+"\n", type_id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, type_id)

		literal_type_id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Literal Type: %s"]`+"\n", literal_type_id, e.Type.String())
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, literal_type_id)

		value_id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Value: %s"]`+"\n", value_id, escape(e.Value))
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, value_id)
	case *EmptyExpression:
		type_id := nextNodeID()
		fmt.Fprintf(w, `  %s [label="Type: Empty Expression"]`+"\n", type_id)
		fmt.Fprintf(w, `  %s -> %s`+"\n", id, type_id)
	}
}

func escape(val string) string {
	return strings.ReplaceAll(val, "\"", "\\\"")
}
