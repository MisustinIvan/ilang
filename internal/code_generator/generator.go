package code_generator

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/MisustinIvan/ilang/internal/ast"
	"github.com/MisustinIvan/ilang/internal/lexer"
)

func generatorError(position lexer.Position, msg string, args ...any) error {
	return fmt.Errorf("%s %s", position.String(), fmt.Sprintf(msg, args...))
}

type functionContext struct {
	currentDecl   *ast.Declaration        // current function declaration ast node
	locals        map[*ast.Identifier]int // maps a local identifier to a stack offset
	stackOffset   int                     // total stack offset of the function to allocate memory on the stack
	argsGenerated int                     // how many function arguments had their code generated
}

type Generator struct {
	source      strings.Builder
	prog        *ast.Program
	ctx         *functionContext
	externals   map[*ast.Identifier]bool
	constants   map[string]*ast.Literal
	label_count int
}

func (g *Generator) label() string {
	g.label_count++
	return fmt.Sprintf(".label_%d", g.label_count)
}

type localFinder struct {
	stackOffset int
	locals      map[*ast.Identifier]int
}

func (f *localFinder) declareLocal(size int, identifier *ast.Identifier) {
	f.locals[identifier] = f.stackOffset
	f.stackOffset += size
}

func (f *localFinder) VisitProgram(p *ast.Program) error                         { return nil }
func (f *localFinder) VisitExternalDeclaration(d *ast.ExternalDeclaration) error { return nil }
func (f *localFinder) VisitBasicType(t *ast.BasicType) error                     { return nil }
func (f *localFinder) VisitArrayType(t *ast.ArrayType) error                     { return nil }
func (f *localFinder) VisitLiteral(l *ast.Literal) error                         { return nil }
func (f *localFinder) VisitIdentifier(i *ast.Identifier) error                   { return nil }

func (f *localFinder) VisitDeclaration(d *ast.Declaration) error {
	for _, arg := range d.Args {
		arg.Accept(f)
	}
	d.Body.Accept(f)
	return nil
}

func (f *localFinder) VisitArgument(a *ast.Argument) error {
	f.declareLocal(a.Type.Size(), a.Identifier)
	return nil
}

func (f *localFinder) VisitReturn(r *ast.Return) error {
	r.Value.Accept(f)
	return nil
}

func (f *localFinder) VisitBind(b *ast.Bind) error {
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

func (f *localFinder) VisitSeparated(s *ast.Separated) error {
	return s.Value.Accept(f)
}

func (f *localFinder) VisitUnary(u *ast.Unary) error {
	return u.Value.Accept(f)
}

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

func findLocals(d *ast.Declaration) (map[*ast.Identifier]int, int) {
	l := &localFinder{
		stackOffset: 8, // start stack offset at 8 to prevent stack corruption
		locals:      map[*ast.Identifier]int{},
	}

	d.Accept(l)

	return l.locals, l.stackOffset
}

// newContext() creates a new function context containing all of the stack
// offsets for the local variables and also the total stack offset of the function.
func (g *Generator) newContext(d *ast.Declaration) {
	locals, stackOffset := findLocals(d)
	// align the stack to 16 bytes
	if stackOffset%16 != 0 {
		stackOffset += 16 - (stackOffset % 16)
	}
	g.ctx = &functionContext{
		currentDecl: d,
		locals:      locals,
		stackOffset: stackOffset,
	}
}

func (g *Generator) writeln(s string)               { g.source.WriteString(s + "\n") }
func (g *Generator) writefln(f string, args ...any) { g.writeln(fmt.Sprintf(f, args...)) }

func New(prog *ast.Program) *Generator {
	return &Generator{
		source:      strings.Builder{},
		prog:        prog,
		ctx:         nil,
		externals:   map[*ast.Identifier]bool{},
		constants:   map[string]*ast.Literal{},
		label_count: 0,
	}
}

func (g *Generator) Generate() (string, error) {
	err := g.prog.Accept(g)

	g.writeln("")
	g.writeln("# data section")
	g.writeln(".data")

	for id, l := range g.constants {
		switch {
		case l.GetType().Equals(ast.BasicTypePtr(ast.String)):
			g.writeln(id + ":")
			g.writeln(".asciz " + l.Value)
		default:
			err = errors.Join(err, generatorError(l.Position, "can't generate constant of type %s", l.GetType().String()))
		}
	}

	return g.source.String(), err
}

// programHeaders() generates program headers, expecting a there to be a `main`
// function that is the entry point.
func (g *Generator) programHeaders() {
	g.writeln("# program headers")
	g.writeln(".text")
	g.writeln(".globl main\n")
}

// VisitProgram() generates code for the whole program, reporting any arrors
// encountered along the way.
func (g *Generator) VisitProgram(p *ast.Program) error {
	var err error
	g.programHeaders()

	g.writeln("# external functions")
	for _, decl := range p.ExternalDeclarations {
		err = errors.Join(err, decl.Accept(g))
	}
	g.writeln("")

	g.writeln("# function declarations")
	for _, decl := range p.Declarations {
		err = errors.Join(err, decl.Accept(g))
	}

	return err
}

// VisitExternalDeclaration() generates code for external function declarations
// using the .extern directive.
func (g *Generator) VisitExternalDeclaration(d *ast.ExternalDeclaration) error {
	g.externals[d.Identifier] = true
	g.writefln(".extern %s", d.Identifier.Name)
	return nil
}

// generatePrologue() generates the function prologue. It creates a new stack
// frame, allocating the required memory for local variables. It expectes
// the g.ctx context to be non-nil and containing valid values.
func (g *Generator) generatePrologue() error {
	g.writeln("# function prologue")
	g.writefln("%s:", g.ctx.currentDecl.Identifier.Name)
	g.writeln("push %rbp")
	g.writeln("mov %rsp, %rbp")
	g.writeln("# locals stack allocation")
	g.writefln("sub $%d, %%rsp", g.ctx.stackOffset)
	g.writeln("")
	return nil
}

func (g *Generator) generateEpilogue() {
	g.writeln("# function epilogue")
	g.writeln("leave")
	g.writeln("ret")
	g.writeln("")
}

func (g *Generator) VisitDeclaration(d *ast.Declaration) error {
	var err error
	g.newContext(d)
	g.generatePrologue()

	for _, arg := range d.Args {
		err = errors.Join(err, arg.Accept(g))
	}
	g.writeln("")

	err = errors.Join(err, d.Body.Accept(g))

	g.generateEpilogue()
	return err
}

// argLocation() returns the location of the nth function argument using the
// x86_64 linux c calling conventions. In case of more than 6 arguments it
// returns an error for now.
func (g *Generator) argLocation(n int) (string, error) {
	switch n {
	case 0:
		return "%rdi", nil
	case 1:
		return "%rsi", nil
	case 2:
		return "%rdx", nil
	case 3:
		return "%rcx", nil
	case 4:
		return "%r8", nil
	case 5:
		return "%r9", nil
	default:
		return "", generatorError(g.ctx.currentDecl.Identifier.Position, "can't use more than 6 arguments for now")
	}
}

// VisitArgument() generates code to move the function argument from its
// location in a register according to the linux x86_64 calling conventions to
// a local variable in the function scope. It currently supports only up to
// 6 arguments.
func (g *Generator) VisitArgument(a *ast.Argument) error {
	location, err := g.argLocation(g.ctx.argsGenerated)
	g.ctx.argsGenerated++
	if err != nil {
		return err
	}

	offset := g.ctx.locals[a.Identifier]
	g.writeln("# move function argument to local")
	g.writefln("mov %s, -%d(%%rbp)", location, offset)
	return nil
}

func (g *Generator) VisitBasicType(t *ast.BasicType) error { return nil }
func (g *Generator) VisitArrayType(t *ast.ArrayType) error { return nil }

// VisitReturn() generates code to move the return value to %rax and then return
func (g *Generator) VisitReturn(r *ast.Return) error {
	g.writeln("# return expression")
	if r.Value != nil {
		g.writeln("# value")
		err := r.Value.Accept(g)
		if err != nil {
			return err
		}
	} else {
		g.writeln("mov $0, %rax")
	}
	g.writeln("ret")
	g.writeln("")
	return nil
}

// VisitBind() generates code to move the bound value to a local variable on
// the stack.
func (g *Generator) VisitBind(b *ast.Bind) error {
	g.writeln("# bind expression")
	g.writeln("# value")
	err := b.Value.Accept(g)
	if err != nil {
		return err
	}

	// expecting the bound value to be stored in %rax
	offset := g.ctx.locals[b.Identifier]
	g.writeln("# move value to local")
	g.writefln("mov %%rax, -%d(%%rbp)", offset)
	return nil
}

func (g *Generator) constLabel() string {
	return fmt.Sprintf(".const_%d", len(g.constants))
}

// VisitLiteral() generates code of a literal value to be stored in %rax.
func (g *Generator) VisitLiteral(l *ast.Literal) error {
	g.writefln("# literal expression of type %s", l.Type.String())
	t, basicType := l.GetType().(*ast.BasicType)
	if !basicType {
		return generatorError(l.Position, "literals of non-basic type are not supported yet")
	}
	switch *t {
	case ast.Int:
		g.writefln("mov $%s, %%rax", l.Value)
	case ast.Bool:
		switch l.Value {
		case "false":
			g.writeln("mov $0, %rax")
		case "true":
			g.writeln("mov $1, %rax")
		default:
			return generatorError(l.Position, "unknown boolean value")
		}

	case ast.String:
		label := g.constLabel()
		g.constants[label] = l
		g.writefln("lea %s(%%rip), %%rax", label)
	case ast.Float:
		return generatorError(l.Position, "floats are not supported yet")

	case ast.Unit:
		g.writeln("mov $0, %rax")

	case ast.Undefined:
		return generatorError(l.Position, "literal of undefined type")
	}
	return nil
}

// VisitIdentifier() generates code to move the value stored in a local variable
// on the stack to %rax.
func (g *Generator) VisitIdentifier(i *ast.Identifier) error {
	g.writeln("# identifier expression")

	offset, exists := g.ctx.locals[i.Resolved]
	if !exists {
		return generatorError(i.Position, "unresolved identifier - \"%s\"", i.Name)
	}

	if _, isArray := i.Resolved.GetType().(*ast.ArrayType); isArray {
		g.writeln("# move resolved address of array type to %rax")
		g.writefln("lea -%d(%%rbp), %%rax", offset)
	} else {
		g.writeln("# move resolved value of simple type to %rax")
		g.writefln("mov -%d(%%rbp), %%rax", offset)
	}

	return nil
}

// VisitCall() generates the code for calling functions. It first iterates through
// all the function arguments, generates the code for their values and pushes them
// to the stack to not be mangled in registers by other generated code...
// Then it pops all the values to their registers according to the linux x86_64
// calling conditions and calls the function either from the PLT for externals
// or normally for user-defined functions.
func (g *Generator) VisitCall(c *ast.Call) error {
	g.writeln("# call expression")
	g.writeln("# arguments")
	for i, arg := range c.Arguments {
		g.writefln("# argument %d", i)
		if err := arg.Accept(g); err != nil {
			return err
		}
		g.writefln("push %%rax # temporarily store argument %d", i)
	}

	g.writeln("# clear %rax for variadic functions")
	g.writeln("xor %rax, %rax")

	for i := range slices.Backward(c.Arguments) {
		reg, _ := g.argLocation(i)
		g.writefln("pop %s # pop argument %d", reg, i)
	}

	g.writeln("# call the function")
	if g.externals[c.Identifier.Resolved] {
		g.writefln("call %s @PLT", c.Identifier.Name)
	} else {
		g.writefln("call %s", c.Identifier.Name)
	}

	return nil
}

func (g *Generator) VisitSeparated(s *ast.Separated) error {
	g.writeln("# separated expression")
	return s.Value.Accept(g)
}

// VisitUnary() generates code for unary expression. It first generates code for
// the value(stored in %rax) and then the code for the unary operator.
func (g *Generator) VisitUnary(u *ast.Unary) error {
	g.writeln("# unary expression")
	if err := u.Value.Accept(g); err != nil {
		return err
	}

	switch u.Operator {
	case ast.Inversion:
		g.writeln("imul $-1, %rax")
	case ast.LogicNegation:
		g.writeln("cmp $0, %rax")
		g.writeln("sete %al")
		g.writeln("movzbq %al, %rax")
	default:
		return generatorError(u.Position, "unknown unary operator")
	}

	return nil
}

// generateBinaryOperator() Generates code for the given binary operator.
// It expects the first operand to be in %rax and the second in %rbx.
func (g *Generator) generateBinaryOperator(o ast.BinaryOperator) error {
	switch o {
	case ast.Addition:
		g.writeln("add %rbx, %rax")
		return nil
	case ast.Subtraction:
		g.writeln("sub %rbx, %rax")
		return nil
	case ast.Multiplication:
		g.writeln("imul %rbx, %rax")
		return nil
	case ast.Division:
		g.writeln("cltd")
		g.writeln("idiv %rbx")
		return nil
	case ast.Equality:
		g.writeln("cmp %rbx, %rax")
		g.writeln("sete %al")
		g.writeln("movzbq %al, %rax")
		return nil
	case ast.Inequality:
		g.writeln("cmp %rbx, %rax")
		g.writeln("setne %al")
		g.writeln("movzbq %al, %rax")
		return nil
	case ast.Less:
		g.writeln("cmp %rbx, %rax")
		g.writeln("setl %al")
		g.writeln("movzbq %al, %rax")
		return nil
	case ast.Greater:
		g.writeln("cmp %rbx, %rax")
		g.writeln("setg %al")
		g.writeln("movzbq %al, %rax")
		return nil
	case ast.LessEqual:
		g.writeln("cmp %rbx, %rax")
		g.writeln("setle %al")
		g.writeln("movzbq %al, %rax")
		return nil
	case ast.GreaterEqual:
		g.writeln("cmp %rbx, %rax")
		g.writeln("setge %al")
		g.writeln("movzbq %al, %rax")
		return nil
	case ast.ShiftLeft:
		g.writeln("shl %rbx, %rax")
		return nil
	case ast.ShiftRight:
		g.writeln("shr %rbx, %rax")
		return nil
	case ast.LogicAnd:
		g.writeln("and %rbx, %rax")
		return nil
	case ast.LogicOr:
		g.writeln("or %rbx, %rax")
		return nil
	default:
		return fmt.Errorf("operator %s not implemented", o.String())
	}
}

// VisitBinary() generates code for a binary expression. It generates code for
// the two operands first, preserving expected order of execution by first
// storing the first value on the stack and only then storing the second.
func (g *Generator) VisitBinary(u *ast.Binary) error {
	g.writeln("# binary expression")
	g.writeln("# left")
	if err := u.Left.Accept(g); err != nil {
		return err
	}
	g.writeln("push %rax")
	g.writeln("# right")
	if err := u.Right.Accept(g); err != nil {
		return err
	}
	g.writeln("mov %rax, %rbx")
	g.writeln("pop %rax")

	g.writefln("# binary operator %s", u.Operator.String())
	return g.generateBinaryOperator(u.Operator)
}

// VisitBlock() generates code for a block expression. It iterates through all
// of the body expressions and generates their code. It also generates code for
// the implicit return expression.
func (g *Generator) VisitBlock(b *ast.Block) error {
	g.writeln("# block expression")
	for _, expr := range b.Body {
		if err := expr.Accept(g); err != nil {
			return err
		}
	}
	g.writeln("# implicit return")
	if b.ImplicitReturn != nil {
		if err := b.ImplicitReturn.Accept(g); err != nil {
			return err
		}
	} else {
		g.writeln("mov $0, %rax")
	}
	return nil
}

// VisitCondition() generates code for a conditional expression.
func (g *Generator) VisitCondition(c *ast.Condition) error {
	g.writeln("# conditional expression")
	if_label := g.label()
	else_label := g.label()
	end_label := g.label()

	g.writeln("# condition")
	c.Condition.Accept(g)
	g.writeln("cmp $1, %rax")
	g.writefln("je %s", if_label)
	if c.Else != nil {
		g.writefln("jmp %s", else_label)
	} else {
		g.writefln("jmp %s", end_label)
	}

	g.writeln("# if branch")
	g.writeln(if_label + ":")
	c.Body.Accept(g)

	if c.Else != nil {
		g.writefln("jmp %s", end_label)
		g.writeln("# else branch")
		g.writeln(else_label + ":")
		c.Else.Accept(g)
	}

	g.writeln("# if end")
	g.writeln(end_label + ":")

	return nil
}

func (g *Generator) VisitAssignment(a *ast.Assignment) error {
	g.writeln("# assignment expression")
	if err := a.Value.Accept(g); err != nil {
		return err
	}

	offset := g.ctx.locals[a.Identifier.Resolved]
	g.writefln("mov %%rax, -%d(%%rbp)", offset)

	return nil
}
