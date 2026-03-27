package code_generator

import (
	"errors"

	"github.com/MisustinIvan/ilang/internal/ast"
)

type localFinder struct {
	stackOffset int
	locals      map[any]int
}

func (f *localFinder) declareLocal(size int, key any) {
	f.stackOffset += size
	f.locals[key] = f.stackOffset
}

func (f *localFinder) VisitProgram(p *ast.Program) error                         { return nil }
func (f *localFinder) VisitExternalDeclaration(d *ast.ExternalDeclaration) error { return nil }
func (f *localFinder) VisitBasicType(t *ast.BasicType) error                     { return nil }
func (f *localFinder) VisitArrayType(t *ast.ArrayType) error                     { return nil }
func (f *localFinder) VisitPointerType(t *ast.PointerType) error                 { return nil }
func (f *localFinder) VisitSliceType(t *ast.SliceType) error {
	if t.LengthIdentifier != nil {
		f.declareLocal(8, t.LengthIdentifier)
	}
	return nil
}
func (f *localFinder) VisitLiteral(l *ast.Literal) error       { return nil }
func (f *localFinder) VisitIdentifier(i *ast.Identifier) error { return nil }
func (f *localFinder) VisitDeclaration(d *ast.Declaration) error {
	for _, arg := range d.Args {
		arg.Accept(f)
	}
	d.Body.Accept(f)
	return nil
}
func (f *localFinder) VisitArgument(a *ast.Argument) error {
	a.Type.Accept(f)
	size := a.Type.Size()
	_, isArray := a.Type.(*ast.ArrayType)
	_, isSlice := a.Type.(*ast.SliceType)
	if isArray || isSlice {
		size = 16 // (pointer, length)
	}
	f.declareLocal(size, a.Identifier)
	return nil
}
func (f *localFinder) VisitReturn(r *ast.Return) error {
	if r.Value != nil {
		return r.Value.Accept(f)
	}
	return nil
}
func (f *localFinder) VisitBind(b *ast.Bind) error {
	b.Type.Accept(f) // handles slice LengthIdentifier
	f.declareLocal(b.Type.Size(), b.Identifier)
	b.Value.Accept(f)
	return nil
}
func (f *localFinder) VisitCall(c *ast.Call) error {
	for _, arg := range c.Arguments {
		arg.Accept(f)
	}
	return nil
}
func (f *localFinder) VisitSeparated(s *ast.Separated) error { return s.Value.Accept(f) }
func (f *localFinder) VisitUnary(u *ast.Unary) error         { return u.Value.Accept(f) }
func (f *localFinder) VisitBinary(u *ast.Binary) error {
	u.Left.Accept(f)
	u.Right.Accept(f)
	return nil
}
func (f *localFinder) VisitBlock(b *ast.Block) error {
	for _, expr := range b.Body {
		expr.Accept(f)
	}
	if b.ImplicitReturn != nil {
		b.ImplicitReturn.Accept(f)
	}
	return nil
}
func (f *localFinder) VisitCondition(c *ast.Condition) error {
	c.Condition.Accept(f)
	c.Body.Accept(f)
	if c.Else != nil {
		c.Else.Accept(f)
	}
	return nil
}
func (f *localFinder) VisitAssignment(a *ast.Assignment) error {
	a.Value.Accept(f)
	return nil
}
func (f *localFinder) VisitDereference(d *ast.Dereference) error {
	return d.Value.Accept(f)
}
func (f *localFinder) VisitLoop(l *ast.Loop) error {
	return errors.Join(l.Condition.Accept(f), l.Body.Accept(f))
}
func (f *localFinder) VisitMake(m *ast.Make) error {
	return m.Length.Accept(f)
}
func (f *localFinder) VisitRelease(r *ast.Release) error {
	return r.Value.Accept(f)
}
func (f *localFinder) VisitIndex(i *ast.Index) error {
	i.Index.Accept(f)
	return nil
}
func (f *localFinder) VisitArrayLiteral(a *ast.ArrayLiteral) error {
	for _, val := range a.Values {
		val.Accept(f)
	}
	f.declareLocal(a.GetType().Size(), a)
	return nil
}

func findLocals(d *ast.Declaration) (map[any]int, int) {
	l := &localFinder{locals: map[any]int{}}
	d.Accept(l)
	return l.locals, l.stackOffset
}
