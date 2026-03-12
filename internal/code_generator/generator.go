package code_generator

import (
	"errors"
	"fmt"
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
	f.stackOffset += size
	f.locals[identifier] = f.stackOffset
}

func (f *localFinder) VisitProgram(p *ast.Program) error                         { return nil }
func (f *localFinder) VisitExternalDeclaration(d *ast.ExternalDeclaration) error { return nil }
func (f *localFinder) VisitBasicType(t *ast.BasicType) error                     { return nil }
func (f *localFinder) VisitArrayType(t *ast.ArrayType) error                     { return nil }
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
		size = 16
	}
	f.declareLocal(size, a.Identifier)
	return nil
}

func (f *localFinder) VisitReturn(r *ast.Return) error {
	r.Value.Accept(f)
	return nil
}

func (f *localFinder) VisitBind(b *ast.Bind) error {
	b.Type.Accept(f) // for slice types with a length identifier
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

func (f *localFinder) VisitIndex(i *ast.Index) error {
	i.Index.Accept(f)
	return nil
}

func findLocals(d *ast.Declaration) (map[*ast.Identifier]int, int) {
	l := &localFinder{
		stackOffset: 0,
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

func (g *Generator) isArgument(id *ast.Identifier) bool {
	for _, arg := range g.ctx.currentDecl.Args {
		if arg.Identifier == id {
			return true
		}
	}
	return false
}

// VisitArgument() generates code to move the function argument from its
// location in a register according to the linux x86_64 calling conventions to
// a local variable in the function scope. It currently supports only up to
// 6 arguments.
func (g *Generator) VisitArgument(a *ast.Argument) error {
	offset := g.ctx.locals[a.Identifier]

	switch t := a.Type.(type) {
	case *ast.SliceType, *ast.ArrayType:
		// Slices and arrays are passed as two 8-byte values: length and pointer
		lenLocation, err := g.argLocation(g.ctx.argsGenerated)
		g.ctx.argsGenerated++
		if err != nil {
			return err
		}
		ptrLocation, err := g.argLocation(g.ctx.argsGenerated)
		g.ctx.argsGenerated++
		if err != nil {
			return err
		}

		g.writeln("# move slice/array length and pointer to local")
		g.writefln("mov %s, -%d(%%rbp)", lenLocation, offset-8)
		g.writefln("mov %s, -%d(%%rbp)", ptrLocation, offset)

		if st, ok := t.(*ast.SliceType); ok && st.LengthIdentifier != nil {
			lenOffset := g.ctx.locals[st.LengthIdentifier]
			g.writefln("mov %s, -%d(%%rbp)", lenLocation, lenOffset)
		}
		return nil

	default:
		location, err := g.argLocation(g.ctx.argsGenerated)
		g.ctx.argsGenerated++
		if err != nil {
			return err
		}
		g.writeln("# move function argument to local")
		g.writefln("mov %s, -%d(%%rbp)", location, offset)
		return nil
	}
}

func (g *Generator) VisitBasicType(t *ast.BasicType) error { return nil }
func (g *Generator) VisitArrayType(t *ast.ArrayType) error { return nil }
func (g *Generator) VisitSliceType(t *ast.SliceType) error { return nil }

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
	offset := g.ctx.locals[b.Identifier]

	err := b.Type.Accept(g)
	if err != nil {
		return err
	}

	switch t := b.Type.(type) {
	case *ast.ArrayType:
		// Array zero-initialization
		if literal, isLiteral := b.Value.(*ast.Literal); isLiteral && literal.Value == "0" {
			g.writeln("# array zero-initialization")
			g.writeln("xor %rax, %rax")

			count := (t.Size() + 7) / 8
			g.writefln("mov $%d, %%rcx", count)
			g.writefln("lea -%d(%%rbp), %%rdi", offset)
			g.writeln("rep stosq")
		} else {
			return generatorError(b.Position, "only zero-intialization is allowed for arrays for now")
		}
	case *ast.SliceType:
		// Slice binding
		valueIdentifier := b.Value.(*ast.Identifier).Resolved

		// get slice length
		switch t := b.Value.GetType().(type) {
		case *ast.ArrayType:
			// length from array
			g.writefln("mov $%d, %%rax # length from array", t.Length)
		case *ast.SliceType:
			// length from slice
			valOffset := g.ctx.locals[valueIdentifier]
			g.writefln("mov -%d(%%rbp), %%rax # length from slice", valOffset-8)
		default:
			return generatorError(b.Value.GetPosition(), "expected array or slice type, got %v", b.Value.GetType())
		}

		// store the length for the bound slice
		g.writefln("mov %%rax, -%d(%%rbp)", offset-8)
		// store value if bound slice has a length identifier
		if t.LengthIdentifier != nil {
			lenOffset := g.ctx.locals[t.LengthIdentifier]
			g.writefln("mov %%rax, -%d(%%rbp)", lenOffset)
		}

		// get pointer
		if err := b.Value.Accept(g); err != nil {
			return err
		}
		// store pointer
		g.writefln("mov %%rax, -%d(%%rbp)", offset)

	case *ast.BasicType:
		g.writeln("# scalar value binding")
		if err := b.Value.Accept(g); err != nil {
			return err
		}
		// Expecting the bound value to be stored in %rax
		g.writefln("mov %%rax, -%d(%%rbp)", offset)
	}

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

	switch i.Resolved.GetType().(type) {
	case *ast.ArrayType:
		if g.isArgument(i.Resolved) {
			g.writeln("# move pointer value of array argument to %rax")
			g.writefln("mov -%d(%%rbp), %%rax", offset)
		} else {
			g.writeln("# move resolved address of array type to %rax")
			g.writefln("lea -%d(%%rbp), %%rax", offset)
		}
	case *ast.SliceType:
		g.writeln("# move pointer value of slice type to %rax")
		g.writefln("mov -%d(%%rbp), %%rax", offset)
	case *ast.BasicType:
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

	physicalArgCount := 0

	g.writeln("# arguments")
	for i, arg := range c.Arguments {
		g.writefln("# argument %d", i)
		switch t := arg.GetType().(type) {
		case *ast.BasicType:
			if err := arg.Accept(g); err != nil {
				return err
			}
			g.writeln("push %rax # push argument of basic type")
			physicalArgCount += 1
		case *ast.ArrayType:
			// arrays passed as slices - (length, pointer)
			// get length
			g.writefln("mov $%d, %%rax", t.Length)
			g.writeln("push %rax # push array length")

			// get pointer
			if err := arg.Accept(g); err != nil {
				return err
			}
			g.writeln("push %rax # push array pointer")

			physicalArgCount += 2

		case *ast.SliceType:
			// slices passed as - (length, pointer)
			if identifier, isIdentifier := arg.(*ast.Identifier); isIdentifier {
				offset := g.ctx.locals[identifier.Resolved]
				g.writefln("mov -%d(%%rbp), %%rax # length of slice", offset-8)
				g.writeln("push %rax # push slice length")
			} else {
				return generatorError(arg.GetPosition(), "slices can be only passed as identifiers")
			}

			if err := arg.Accept(g); err != nil {
				return err
			}
			g.writeln("push %rax # push slice pointer")

			physicalArgCount += 2
		}
	}

	g.writeln("# clear %rax for variadic functions")
	g.writeln("xor %rax, %rax")

	for i := physicalArgCount - 1; i >= 0; i-- {
		reg, err := g.argLocation(i)
		if err != nil {
			return err
		}
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

// VisitAssignment() generates code for assigning a value to a target.
// It distinguishes between two types of assignment - to a scalar variable
// and to an array.
func (g *Generator) VisitAssignment(a *ast.Assignment) error {
	g.writeln("# assignment expression")
	if err := a.Value.Accept(g); err != nil {
		return err
	}

	switch target := a.Target.(type) {
	case *ast.Identifier:
		offset, exists := g.ctx.locals[target.Resolved]
		if !exists {
			return generatorError(target.Position, "unresolved identifier - \"%s\"", target.Name)
		}

		if arrayType, isArray := target.Resolved.GetType().(*ast.ArrayType); isArray {
			// assuming that the assignment was already type checked and that an address is in %rax
			count := (arrayType.Size() + 7) / 8
			if literal, ok := a.Value.(*ast.Literal); ok && literal.Value == "0" {
				g.writeln("# array zero-assignment")
				g.writeln("xor %rax, %rax")
				g.writefln("mov $%d, %%rcx", count)
				g.writefln("lea -%d(%%rbp), %%rdi", offset)
				g.writeln("rep stosq")
			} else {
				g.writeln("# array copy assignment")
				g.writeln("mov %rax, %rsi")
				g.writefln("lea -%d(%%rbp), %%rdi", offset)
				g.writefln("mov $%d, %%rcx", count)
				g.writeln("rep movsq")
			}
		} else {
			// scalar assignment
			g.writefln("mov %%rax, -%d(%%rbp)", offset)
		}
	case *ast.Index:
		g.writeln("push %rax") // save value

		if err := target.Index.Accept(g); err != nil {
			return err
		}
		g.writeln("push %rax") // save index

		offset, exists := g.ctx.locals[target.Identifier.Resolved]
		if !exists {
			return generatorError(target.Identifier.Position, "unresolved identifier - \"%s\"", target.Identifier.Name)
		}

		if _, isSlice := target.Identifier.Resolved.GetType().(*ast.SliceType); isSlice {
			g.writefln("mov -%d(%%rbp), %%rcx", offset)
		} else if _, isArray := target.Identifier.Resolved.GetType().(*ast.ArrayType); isArray && g.isArgument(target.Identifier.Resolved) {
			g.writefln("mov -%d(%%rbp), %%rcx", offset)
		} else {
			g.writefln("lea -%d(%%rbp), %%rcx", offset)
		}

		elementSize := target.GetType().Size()
		g.writeln("pop %rdx") // index
		g.writeln("pop %rax") // value
		g.writefln("mov %%rax, (%%rcx, %%rdx, %d)", elementSize)
	default:
		return generatorError(a.Position, "invalid assignment target")
	}

	return nil
}

// VisitIndex() generates code for indexing an array as a value.
// It generates the value for the index, storing the result in %rax
// and using that as an offset for loading a value from the array base
// offset to %rax.
func (g *Generator) VisitIndex(i *ast.Index) error {
	g.writeln("# index expression")
	if err := i.Index.Accept(g); err != nil {
		return err
	}
	g.writeln("push %rax")

	offset, exists := g.ctx.locals[i.Identifier.Resolved]
	if !exists {
		return generatorError(i.Identifier.Position, "unresolved identifier - \"%s\"", i.Identifier.Name)
	}

	if _, isSlice := i.Identifier.Resolved.GetType().(*ast.SliceType); isSlice {
		g.writefln("mov -%d(%%rbp), %%rcx", offset)
	} else if _, isArray := i.Identifier.Resolved.GetType().(*ast.ArrayType); isArray && g.isArgument(i.Identifier.Resolved) {
		g.writefln("mov -%d(%%rbp), %%rcx", offset)
	} else {
		g.writefln("lea -%d(%%rbp), %%rcx", offset)
	}

	elementSize := i.GetType().Size()
	g.writeln("pop %rdx")
	g.writefln("mov (%%rcx, %%rdx, %d), %%rax", elementSize)
	return nil
}
