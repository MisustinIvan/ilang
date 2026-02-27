package ast_visualizer

import (
	"fmt"
	"strings"

	"github.com/MisustinIvan/ilang/internal/ast"
)

type color string

const (
	none       color = ""
	purple     color = "#ae02e2"
	blue       color = "#2685d3"
	red        color = "#d1062e"
	green      color = "#1dd159"
	orange     color = "#c18607"
	light_blue color = "#8bd6b1"
)

func escape(val string) string {
	return strings.ReplaceAll(val, "\"", "\\\"")
}

type AstVisualizer struct {
	program   *ast.Program
	output    *strings.Builder
	nodeStack []int
	nodeID    int
}

func New(p *ast.Program) *AstVisualizer {
	return &AstVisualizer{
		program:   p,
		output:    &strings.Builder{},
		nodeStack: []int{},
		nodeID:    0,
	}
}

func (v *AstVisualizer) Pop() {
	if len(v.nodeStack) > 0 {
		v.nodeStack = v.nodeStack[:len(v.nodeStack)-1]
	}
}

func (v *AstVisualizer) NewID() int {
	v.nodeID++
	return v.nodeID
}

func (v *AstVisualizer) ParentID() int {
	return v.nodeStack[len(v.nodeStack)-1]
}

func (v *AstVisualizer) WriteNode(label string, color color, args ...any) {
	id := v.NewID()
	label = escape(fmt.Sprintf(label, args...))
	if color != none {
		fmt.Fprintf(v.output, "%d [label=\"%s\", style=filled, fillcolor=\"%s\"]\n", id, label, color)
	} else {
		fmt.Fprintf(v.output, "%d [label=\"%s\"]\n", id, label)
	}
	if len(v.nodeStack) > 0 {
		fmt.Fprintf(v.output, "  %d -> %d\n", v.ParentID(), id)
	}
	v.nodeStack = append(v.nodeStack, id)
}

func (v *AstVisualizer) Visualize() (string, error) {
	fmt.Fprintln(v.output, "digraph AST {")
	fmt.Fprintln(v.output, "  node [shape=box];")

	err := v.VisitProgram(v.program)
	if err != nil {
		return "", err
	}

	fmt.Fprintln(v.output, "}")
	return v.output.String(), nil
}

func (v *AstVisualizer) VisitProgram(p *ast.Program) error {
	v.WriteNode("Program", none)
	defer v.Pop()

	for _, extrn := range p.ExternalDeclarations {
		if err := extrn.Accept(v); err != nil {
			return err
		}
	}
	for _, decl := range p.Declarations {
		if err := decl.Accept(v); err != nil {
			return err
		}
	}
	return nil
}

func (v *AstVisualizer) VisitDeclaration(d *ast.Declaration) error {
	v.WriteNode("Declaration", none)
	defer v.Pop()

	if err := d.Type.Accept(v); err != nil {
		return err
	}
	if err := d.Identifier.Accept(v); err != nil {
		return err
	}

	v.WriteNode("Parameters", none)
	for _, param := range d.Params {
		if err := param.Accept(v); err != nil {
			return err
		}
	}
	v.Pop()

	return d.Body.Accept(v)
}

func (v *AstVisualizer) VisitExternalDeclaration(d *ast.ExternalDeclaration) error {
	v.WriteNode("External Declaration", none)
	defer v.Pop()
	if err := d.Type.Accept(v); err != nil {
		return err
	}
	if err := d.Identifier.Accept(v); err != nil {
		return err
	}
	v.WriteNode("Parameters", none)
	for _, param := range d.Params {
		if err := param.Accept(v); err != nil {
			return err
		}
	}
	v.Pop()

	return nil
}

func (v *AstVisualizer) VisitParameter(p *ast.Parameter) error {
	v.WriteNode("Parameter", none)
	defer v.Pop()
	if err := p.Type.Accept(v); err != nil {
		return err
	}
	return p.Identifier.Accept(v)
}

func (v *AstVisualizer) writeType(t ast.Type) {
	var color color

	switch t {
	case ast.Undefined:
		color = purple
	case ast.Int:
		color = blue
	case ast.Bool:
		color = red
	case ast.Float:
		color = green
	case ast.String:
		color = orange
	case ast.Unit:
		color = light_blue
	}

	v.WriteNode("Type: %s", color, t.String())
	v.Pop()
}

func (v *AstVisualizer) VisitType(t *ast.Type) error {
	v.writeType(*t)
	return nil
}

func (v *AstVisualizer) VisitReturn(r *ast.Return) error {
	v.WriteNode("Return", none)
	defer v.Pop()
	return r.Value.Accept(v)
}

func (v *AstVisualizer) VisitBind(b *ast.Bind) error {
	v.WriteNode("Bind", none)
	defer v.Pop()
	defer v.Pop()

	if err := b.Identifier.Accept(v); err != nil {
		return err
	}
	if err := b.Type.Accept(v); err != nil {
		return err
	}
	v.WriteNode("Value", none)
	return b.Value.Accept(v)
}

func (v *AstVisualizer) VisitLiteral(l *ast.Literal) error {
	v.WriteNode("Literal: %s", none, l.Value)
	v.writeType(l.Type)
	v.Pop()
	return nil
}

func (v *AstVisualizer) VisitIdentifier(i *ast.Identifier) error {
	v.WriteNode("Identifier: %s", none, i.Name)
	v.WriteNode("Resolved: %v", none, i.Resolved != nil)
	v.Pop()
	v.writeType(i.Type)
	v.Pop()
	return nil
}

func (v *AstVisualizer) VisitCall(c *ast.Call) error {
	v.WriteNode("Call", none)
	defer v.Pop()

	if err := c.Identifier.Accept(v); err != nil {
		return err
	}
	v.WriteNode("Arguments", none)
	for _, arg := range c.Arguments {
		if err := arg.Accept(v); err != nil {
			return err
		}
	}
	v.Pop()
	return nil
}

func (v *AstVisualizer) VisitSeparated(s *ast.Separated) error {
	v.WriteNode("Separated", none)
	defer v.Pop()
	return s.Value.Accept(v)
}

func (v *AstVisualizer) VisitUnary(u *ast.Unary) error {
	v.WriteNode("Unary: %s", none, u.Operator.String())
	defer v.Pop()
	return u.Value.Accept(v)
}

func (v *AstVisualizer) VisitBinary(b *ast.Binary) error {
	v.WriteNode("Binary: %s", none, b.Operator.String())
	defer v.Pop()

	if err := b.Left.Accept(v); err != nil {
		return err
	}
	return b.Right.Accept(v)
}

func (v *AstVisualizer) VisitBlock(b *ast.Block) error {
	v.WriteNode("Block", none)
	defer v.Pop()

	v.WriteNode("Body", none)
	for _, expr := range b.Body {
		if err := expr.Accept(v); err != nil {
			return err
		}
	}
	v.Pop()

	if b.ImplicitReturn != nil {
		v.WriteNode("Implicit Return", none)
		if err := b.ImplicitReturn.Accept(v); err != nil {
			return err
		}
		v.Pop()
	}

	return nil
}

func (v *AstVisualizer) VisitCondition(c *ast.Condition) error {
	v.WriteNode("Condition", none)
	defer v.Pop()

	v.WriteNode("If", none)
	if err := c.Condition.Accept(v); err != nil {
		return err
	}
	v.Pop()

	v.WriteNode("Then", none)
	if err := c.Body.Accept(v); err != nil {
		return err
	}
	v.Pop()

	if c.Else != nil {
		v.WriteNode("Else", none)
		if err := c.Else.Accept(v); err != nil {
			return err
		}
		v.Pop()
	}
	return nil
}

func (v *AstVisualizer) VisitAssignment(a *ast.Assignment) error {
	v.WriteNode("Assignment", none)
	defer v.Pop()

	if err := a.Identifier.Accept(v); err != nil {
		return err
	}
	return a.Value.Accept(v)
}
