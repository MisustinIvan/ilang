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
	return fmt.Errorf("%s %s\n%s", position.String(), fmt.Sprintf(msg, args...), position.Snippet(1))
}

type functionContext struct {
	currentDecl         *ast.Declaration // current function declaration ast node
	locals              map[any]int      // maps a local identifier or node to a stack offset
	stackOffset         int              // total stack offset of the function to allocate memory on the stack
	intArgsGenerated    int              // how many integer function arguments had their code generated
	floatArgsGenerated  int              // how many float function arguments had their code generated
	stackSlotsGenerated int              // how many stack slots had their local-move code generated
}

type Generator struct {
	source     strings.Builder
	prog       *ast.Program
	ctx        *functionContext
	externals  map[*ast.Identifier]bool
	constants  map[string]*ast.Literal
	labelCount int
}

func (g *Generator) label() string {
	g.labelCount++
	return fmt.Sprintf(".label_%d", g.labelCount)
}

// loadScalar moves the local at offset into %rax.
func (g *Generator) loadScalar(offset int) {
	g.writefln("mov -%d(%%rbp), %%rax", offset)
}

// storeScalar writes %rax to the local at offset.
func (g *Generator) storeScalar(offset int) {
	g.writefln("mov %%rax, -%d(%%rbp)", offset)
}

// loadFloatScalar loads the float scalar from offset to %xmm0
func (g *Generator) loadFloatScalar(offset int) {
	g.writefln("movsd -%d(%%rbp), %%xmm0", offset)
}

// storeFloatScalar stores the float scalar from %xmm0 to local at offset
func (g *Generator) storeFloatScalar(offset int) {
	g.writefln("movsd %%xmm0, -%d(%%rbp)", offset)
}

// loadArrayAddr loads the address of the local array at offset into %rax.
func (g *Generator) loadArrayAddr(offset int) {
	g.writefln("lea -%d(%%rbp), %%rax", offset)
}

// loadSlice loads pointer to %rax and length to %rbx from the slice slot at offset.
func (g *Generator) loadSlice(offset int) {
	g.writefln("mov -%d(%%rbp), %%rax", offset)   // pointer
	g.writefln("mov -%d(%%rbp), %%rbx", offset-8) // length
}

// storeSlice writes pointer in %rax and length in %rbx into the slice slot at offset.
func (g *Generator) storeSlice(offset int) {
	g.writefln("mov %%rax, -%d(%%rbp)", offset)   // pointer
	g.writefln("mov %%rbx, -%d(%%rbp)", offset-8) // length
}

func (g *Generator) pushValue(t ast.Type) {
	if t.Equals(ast.BasicTypePtr(ast.Float)) {
		g.writeln("sub $8, %rsp")
		g.writeln("movq %xmm0, (%rsp)")
	} else {
		g.writeln("push %rax")
	}
}

func (g *Generator) popValue(t ast.Type) {
	if t.Equals(ast.BasicTypePtr(ast.Float)) {
		g.writeln("movsd (%rsp), %xmm0")
		g.writeln("add $8, %rsp")
	} else {
		g.writeln("pop %rax")
	}
}

// zeroArray zeroes qwords*8 bytes starting at -offset(%rbp) using rep stosq.
func (g *Generator) zeroArray(offset, qwords int) {
	g.writeln("xor %rax, %rax")
	g.writefln("mov $%d, %%rcx", qwords)
	g.writefln("lea -%d(%%rbp), %%rdi", offset)
	g.writeln("rep stosq")
}

// copyArray copies qwords*8 bytes from the address in %rax to -offset(%rbp).
func (g *Generator) copyArray(offset, qwords int) {
	g.writeln("mov %rax, %rsi")
	g.writefln("lea -%d(%%rbp), %%rdi", offset)
	g.writefln("mov $%d, %%rcx", qwords)
	g.writeln("rep movsq")
}

func (g *Generator) writeln(s string)               { g.source.WriteString(s + "\n") }
func (g *Generator) writefln(f string, args ...any) { g.writeln(fmt.Sprintf(f, args...)) }

func New(prog *ast.Program) *Generator {
	return &Generator{
		prog:      prog,
		externals: map[*ast.Identifier]bool{},
		constants: map[string]*ast.Literal{},
	}
}

// newContext creates a function context with pre-computed stack offsets.
func (g *Generator) newContext(d *ast.Declaration) {
	locals, stackOffset := findLocals(d)
	if stackOffset%16 != 0 {
		stackOffset += 16 - (stackOffset % 16)
	}
	g.ctx = &functionContext{
		currentDecl: d,
		locals:      locals,
		stackOffset: stackOffset,
	}
}

func (g *Generator) Generate() (string, error) {
	err := g.prog.Accept(g)
	g.writeln("")
	g.writeln("# data section")
	g.writeln(".data")
	g.writeln(".const_neg_one:")
	g.writeln(".double -1.0")
	for id, l := range g.constants {
		switch {
		case l.GetType().Equals(ast.BasicTypePtr(ast.String)):
			g.writeln(id + ":")
			g.writeln(".asciz " + l.Value)
		case l.GetType().Equals(ast.BasicTypePtr(ast.Float)):
			g.writeln(id + ":")
			g.writeln(".double " + l.Value)
		default:
			err = errors.Join(err, generatorError(l.Position, "can't generate constant of type %s", l.GetType().String()))
		}
	}
	return g.source.String(), err
}

func (g *Generator) programHeaders() {
	g.writeln("# program headers")
	g.writeln(".text")
	g.writeln(".globl main\n")
	g.generateBuiltinExterns()
}

func (g *Generator) generateBuiltinExterns() {
	g.writeln(".extern malloc")
	g.writeln(".extern free")
}

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

func (g *Generator) VisitExternalDeclaration(d *ast.ExternalDeclaration) error {
	g.externals[d.Identifier] = true
	g.writefln(".extern %s", d.Identifier.Name)
	return nil
}

func (g *Generator) generatePrologue() {
	g.writeln("# function prologue")
	g.writefln("%s:", g.ctx.currentDecl.Identifier.Name)
	g.writeln("push %rbp")
	g.writeln("mov %rsp, %rbp")
	g.writefln("sub $%d, %%rsp", g.ctx.stackOffset)
	g.writeln("")
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

// storeNthIntArg expects the stored value in %rax. It stores it for according to
// the System V AMD64 ABI.
func (g *Generator) storeNthIntArg(n int) {
	switch n {
	case 0:
		g.writeln("mov %rax, %rdi")
	case 1:
		g.writeln("mov %rax, %rsi")
	case 2:
		g.writeln("mov %rax, %rdx")
	case 3:
		g.writeln("mov %rax, %rcx")
	case 4:
		g.writeln("mov %rax, %r8")
	case 5:
		g.writeln("mov %rax, %r9")
	default:
		panic(fmt.Sprintf("storeNthIntArg called with stack slot %d", n))
	}
}

// loadNthArg loads the n-th int argument to a given local offset according to
// the System V AMD64 ABI
func (g *Generator) loadNthIntArg(n, offset int) {
	switch n {
	case 0:
		g.writefln("mov %%rdi, -%d(%%rbp)", offset)
	case 1:
		g.writefln("mov %%rsi, -%d(%%rbp)", offset)
	case 2:
		g.writefln("mov %%rdx, -%d(%%rbp)", offset)
	case 3:
		g.writefln("mov %%rcx, -%d(%%rbp)", offset)
	case 4:
		g.writefln("mov %%r8, -%d(%%rbp)", offset)
	case 5:
		g.writefln("mov %%r9, -%d(%%rbp)", offset)
	default:
		panic("unexpected int register")
	}
}

func (g *Generator) storeNthFloatArg(n int) {
	if n > 7 {
		panic(fmt.Sprintf("storeNthFloatArg called with slot %d", n))
	}
	g.writefln("movq %%rax, %%xmm%d", n)
}

func (g *Generator) loadNthFloatArg(n, offset int) {
	if n <= 7 {
		g.writefln("movsd %%xmm%d, -%d(%%rbp)", n, offset)
	} else {
		panic("unexpected float register")
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

func (g *Generator) loadStackSlot(slotIndex, localOffset int, isFloat bool) {
	stackOffset := 16 + (slotIndex * 8)

	if isFloat {
		g.writefln("movq %d(%%rbp), %%xmm0", stackOffset)
		g.writefln("movq %%xmm0, -%d(%%rbp)", localOffset)
	} else {
		g.writefln("mov %d(%%rbp), %%rax", stackOffset)
		g.writefln("mov %%rax, -%d(%%rbp)", localOffset)
	}
}

// VisitArgument moves an incoming argument from its ABI register(s) to the stack.
// Arrays and slices occupy two consecutive registers/locations: (length, pointer).
func (g *Generator) VisitArgument(a *ast.Argument) error {
	offset := g.ctx.locals[a.Identifier]

	switch t := a.Type.(type) {
	case *ast.BasicType, *ast.PointerType:
		if t.Equals(ast.BasicTypePtr(ast.Float)) {
			if g.ctx.floatArgsGenerated < 8 {
				g.loadNthFloatArg(g.ctx.floatArgsGenerated, offset)
				g.ctx.floatArgsGenerated++
			} else {
				g.loadStackSlot(g.ctx.stackSlotsGenerated, offset, true)
				g.ctx.stackSlotsGenerated++
			}
		} else {
			if g.ctx.intArgsGenerated < 6 {
				g.loadNthIntArg(g.ctx.intArgsGenerated, offset)
				g.ctx.intArgsGenerated++
			} else {
				g.loadStackSlot(g.ctx.stackSlotsGenerated, offset, false)
				g.ctx.stackSlotsGenerated++
			}
		}

	case *ast.SliceType, *ast.ArrayType:
		if g.ctx.intArgsGenerated < 6 {
			g.loadNthIntArg(g.ctx.intArgsGenerated, offset)
			g.ctx.intArgsGenerated++
		} else {
			g.loadStackSlot(g.ctx.stackSlotsGenerated, offset, false)
			g.ctx.stackSlotsGenerated++
		}

		if g.ctx.intArgsGenerated < 6 {
			g.loadNthIntArg(g.ctx.intArgsGenerated, offset-8)
			g.ctx.intArgsGenerated++

			if st, ok := t.(*ast.SliceType); ok && st.LengthIdentifier != nil {
				lenOffset := g.ctx.locals[st.LengthIdentifier]
				g.loadNthIntArg(g.ctx.intArgsGenerated-1, lenOffset)
			}
		} else {
			g.loadStackSlot(g.ctx.stackSlotsGenerated, offset-8, false)
			g.ctx.stackSlotsGenerated++
			if st, ok := t.(*ast.SliceType); ok && st.LengthIdentifier != nil {
				lenOffset := g.ctx.locals[st.LengthIdentifier]
				g.loadStackSlot(g.ctx.stackSlotsGenerated-1, lenOffset, false)
			}
		}

	}
	return nil
}

func (g *Generator) VisitBasicType(t *ast.BasicType) error     { return nil }
func (g *Generator) VisitArrayType(t *ast.ArrayType) error     { return nil }
func (g *Generator) VisitSliceType(t *ast.SliceType) error     { return nil }
func (g *Generator) VisitPointerType(t *ast.PointerType) error { return nil }

// VisitReturn moves the return value to %rax/%xmm0 (and %rbx for slices/arrays) and returns.
func (g *Generator) VisitReturn(r *ast.Return) error {
	g.writeln("# return")
	if r.Value != nil {
		if err := r.Value.Accept(g); err != nil {
			return err
		}
	} else {
		g.writeln("mov $0, %rax")
	}
	g.writeln("ret")
	g.writeln("")
	return nil
}

// VisitBind evaluates the right-hand value and stores it into the declared local.
// The value protocol means slices and arrays always arrive as (%rax, %rbx),
// so there is no need to inspect the value's type - only the declared binding type.
func (g *Generator) VisitBind(b *ast.Bind) error {
	g.writeln("# bind")
	offset := g.ctx.locals[b.Identifier]
	switch t := b.Type.(type) {
	case *ast.ArrayType:
		if lit, ok := b.Value.(*ast.Literal); ok && lit.Value == "0" {
			g.writeln("# array zero-init")
			g.zeroArray(offset, (t.Size()+7)/8)
		} else if arrLit, ok := b.Value.(*ast.ArrayLiteral); ok {
			return g.generateArrayLiteralInit(arrLit, offset)
		} else {
			// any expression yielding a pointer in %rax
			if err := b.Value.Accept(g); err != nil {
				return err
			}
			g.writeln("# array copy")
			g.copyArray(offset, (t.Size()+7)/8)
		}
	case *ast.SliceType:
		if err := b.Value.Accept(g); err != nil {
			return err
		}
		g.writeln("# slice bind")
		g.storeSlice(offset)
		if t.LengthIdentifier != nil {
			lenOffset := g.ctx.locals[t.LengthIdentifier]
			g.writefln("mov %%rbx, -%d(%%rbp)", lenOffset)
		}
	case *ast.BasicType, *ast.PointerType:
		if err := b.Value.Accept(g); err != nil {
			return err
		}
		if t.Equals(ast.BasicTypePtr(ast.Float)) {
			g.storeFloatScalar(offset)
		} else {
			g.storeScalar(offset)
		}
	default:
		return generatorError(b.GetPosition(), "unexpected type %s", b.Identifier.Resolved.GetType().String())
	}
	return nil
}

func (g *Generator) constLabel() string {
	return fmt.Sprintf(".const_%d", len(g.constants))
}

// VisitLiteral generates a scalar literal, leaving the value in %rax.
func (g *Generator) VisitLiteral(l *ast.Literal) error {
	g.writefln("# literal (%s)", l.Type.String())
	t, ok := l.GetType().(*ast.BasicType)
	if !ok {
		return generatorError(l.Position, "literals of non-basic type are not supported")
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
			return generatorError(l.Position, "unknown boolean literal %q", l.Value)
		}
	case ast.String:
		label := g.constLabel()
		g.constants[label] = l
		g.writefln("lea %s(%%rip), %%rax", label)
	case ast.Float:
		label := g.constLabel()
		g.constants[label] = l
		g.writefln("movsd %s(%%rip), %%xmm0", label)
	case ast.Unit:
		g.writeln("mov $0, %rax")
	case ast.Undefined:
		return generatorError(l.Position, "literal of undefined type")
	}
	return nil
}

// VisitIdentifier evaluates an identifier per the value protocol:
//
//	BasicType -> scalar in %rax
//			  -> Float -> scalar in %xmm0
//	ArrayType -> pointer in %rax, length in %rbx
//	SliceType -> pointer in %rax, length in %rbx
func (g *Generator) VisitIdentifier(i *ast.Identifier) error {
	g.writeln("# identifier")
	offset, exists := g.ctx.locals[i.Resolved]
	if !exists {
		return generatorError(i.Position, "unresolved identifier %q", i.Name)
	}
	switch t := i.Resolved.GetType().(type) {
	case *ast.ArrayType:
		if g.isArgument(i.Resolved) {
			g.loadScalar(offset) // stored as a pointer when passed in
		} else {
			g.loadArrayAddr(offset)
		}
		g.writefln("mov $%d, %%rbx", t.Length)
	case *ast.SliceType:
		g.loadSlice(offset)
	case *ast.BasicType, *ast.PointerType:
		if t.Equals(ast.BasicTypePtr(ast.Float)) {
			g.loadFloatScalar(offset)
		} else {
			g.loadScalar(offset)
		}
	default:
		return generatorError(i.GetPosition(), "unexpected type %s", i.Resolved.GetType().String())
	}
	return nil
}

// VisitCall generates a function call using the System V AMD64 ABI.
//
// Arguments are evaluated and pushed to the stack so that register
// allocations within each argument do not interfere with each other.
// Arrays and slices each occupy two physical arguments: (length, pointer).
// Once all values are on the stack they are popped into the correct registers
// in reverse push order.
func (g *Generator) VisitCall(c *ast.Call) error {
	g.writeln("# call")

	// pre-pass - map every argument to destination
	type argDest struct {
		isFloat bool
		isStack bool
		regNum  int
	}

	slots := []argDest{}

	intRegsUsed := 0
	floatRegsUsed := 0
	stackSlotsUsed := 0

	for _, arg := range c.Arguments {
		switch t := arg.GetType().(type) {
		case *ast.BasicType, *ast.PointerType:
			isFloat := t.Equals(ast.BasicTypePtr(ast.Float))

			if isFloat {
				if floatRegsUsed < 8 {
					slots = append(slots, argDest{isFloat: true, isStack: false, regNum: floatRegsUsed})
					floatRegsUsed++
				} else {
					slots = append(slots, argDest{isFloat: true, isStack: true})
					stackSlotsUsed++
				}
			} else {
				if intRegsUsed < 6 {
					slots = append(slots, argDest{isFloat: false, isStack: false, regNum: intRegsUsed})
					intRegsUsed++
				} else {
					slots = append(slots, argDest{isFloat: false, isStack: true})
					stackSlotsUsed++
				}
			}
		case *ast.ArrayType, *ast.SliceType:
			for range 2 {
				if intRegsUsed < 6 {
					slots = append(slots, argDest{isFloat: false, isStack: false, regNum: intRegsUsed})
					intRegsUsed++
				} else {
					slots = append(slots, argDest{isFloat: false, isStack: true})
					stackSlotsUsed++
				}
			}
		default:
			panic("unexpected argument type")
		}
	}

	// evaluate everything right-to-left, pushing everything onto the stack
	for _, arg := range slices.Backward(c.Arguments) {
		if err := arg.Accept(g); err != nil {
			return err
		}
		switch t := arg.GetType().(type) {
		case *ast.BasicType, *ast.PointerType:
			if t.Equals(ast.BasicTypePtr(ast.Float)) {
				g.writeln("movq %xmm0, %rax")
			}
			g.writeln("push %rax")
		case *ast.SliceType, *ast.ArrayType:
			g.writeln("push %rbx # length")
			g.writeln("push %rax # pointer")
		default:
			panic("unexpected argument type")
		}
	}

	// move register arguments to registers
	for i, dest := range slots {
		if !dest.isStack {
			offset := i * 8
			g.writefln("mov %d(%%rsp), %%rax", offset)
			if dest.isFloat {
				g.storeNthFloatArg(dest.regNum)
			} else {
				g.storeNthIntArg(dest.regNum)
			}
		}
	}

	// compact the rest of the stack slots
	totalSlots := len(slots)
	adjustment := (totalSlots - stackSlotsUsed) * 8
	stackArgIndex := stackSlotsUsed - 1

	for i := totalSlots - 1; i >= 0; i-- {
		dest := slots[i]
		if dest.isStack {
			srcOffset := i * 8
			destOffset := adjustment + (stackArgIndex * 8)

			if srcOffset != destOffset { // optimize out unnecessary moves
				g.writefln("mov %d(%%rsp), %%rax", srcOffset)
				g.writefln("mov %%rax, %d(%%rsp)", destOffset)
			}
			stackArgIndex--
		}
	}

	// discard register argument slots
	if adjustment > 0 {
		g.writefln("add $%d, %%rsp", adjustment)
	}

	if g.externals[c.Identifier.Resolved] {
		g.writefln("mov $%d, %%rax", floatRegsUsed)
		g.writefln("call %s@PLT", c.Identifier.Name)
	} else {
		g.writeln("xor %rax, %rax")
		g.writefln("call %s", c.Identifier.Name)
	}

	// clean up the remaining stack arguments
	if stackSlotsUsed > 0 {
		g.writefln("add $%d, %%rsp", stackSlotsUsed*8)
	}

	return nil
}

func (g *Generator) VisitSeparated(s *ast.Separated) error {
	g.writeln("# separated")
	return s.Value.Accept(g)
}

func (g *Generator) VisitUnary(u *ast.Unary) error {
	g.writeln("# unary")

	if u.Operator == ast.AddressOf {
		id, ok := u.Value.(*ast.Identifier)
		if !ok {
			return generatorError(u.GetPosition(), "can only take address of identifiers")
		}
		offset := g.ctx.locals[id.Resolved] // identifier already resolved
		g.writefln("lea -%d(%%rbp), %%rax", offset)
		return nil
	}

	if err := u.Value.Accept(g); err != nil {
		return err
	}
	switch u.Operator {
	case ast.Inversion:
		if u.Value.GetType().Equals(ast.BasicTypePtr(ast.Float)) {
			g.writeln("movsd .const_neg_one(%rip), %xmm1")
			g.writeln("mulsd %xmm1, %xmm0")
		} else {
			g.writeln("imul $-1, %rax")
		}
	case ast.LogicNegation:
		g.writeln("cmp $0, %rax")
		g.writeln("sete %al")
		g.writeln("movzbq %al, %rax")
	default:
		return generatorError(u.Position, "unknown unary operator")
	}
	return nil
}

// generateBinaryOperator emits the instruction(s) for the given operator.
// Expects: left operand in %rax, right operand in %rbx.
func (g *Generator) generateBinaryOperator(o ast.BinaryOperator) error {
	g.writeln("# non-float binary operator")
	switch o {
	case ast.Addition:
		g.writeln("add %rbx, %rax")
	case ast.Subtraction:
		g.writeln("sub %rbx, %rax")
	case ast.Multiplication:
		g.writeln("imul %rbx, %rax")
	case ast.Division:
		g.writeln("cltd")
		g.writeln("idiv %rbx")
	case ast.Modulo:
		g.writeln("cqto")
		g.writeln("idiv %rbx")
		g.writeln("mov %rdx, %rax")
	case ast.Equality:
		g.writeln("cmp %rbx, %rax")
		g.writeln("sete %al")
		g.writeln("movzbq %al, %rax")
	case ast.Inequality:
		g.writeln("cmp %rbx, %rax")
		g.writeln("setne %al")
		g.writeln("movzbq %al, %rax")
	case ast.Less:
		g.writeln("cmp %rbx, %rax")
		g.writeln("setl %al")
		g.writeln("movzbq %al, %rax")
	case ast.Greater:
		g.writeln("cmp %rbx, %rax")
		g.writeln("setg %al")
		g.writeln("movzbq %al, %rax")
	case ast.LessEqual:
		g.writeln("cmp %rbx, %rax")
		g.writeln("setle %al")
		g.writeln("movzbq %al, %rax")
	case ast.GreaterEqual:
		g.writeln("cmp %rbx, %rax")
		g.writeln("setge %al")
		g.writeln("movzbq %al, %rax")
	case ast.ShiftLeft:
		g.writeln("mov %rbx, %rcx")
		g.writeln("shl %cl, %rax")
	case ast.ShiftRight:
		g.writeln("mov %rbx, %rcx")
		g.writeln("sar %cl, %rax")
	case ast.LogicAnd:
		g.writeln("and %rbx, %rax")
	case ast.LogicOr:
		g.writeln("or %rbx, %rax")
	default:
		return fmt.Errorf("operator %s not implemented", o.String())
	}
	return nil
}

// the same as with non-float types but with %xmm0 and %xmm1
func (g *Generator) generateFloatBinaryOperator(o ast.BinaryOperator) error {
	g.writeln("# float binary operator")
	switch o {
	case ast.Addition:
		g.writeln("addsd %xmm1, %xmm0")
	case ast.Subtraction:
		g.writeln("subsd %xmm1, %xmm0")
	case ast.Multiplication:
		g.writeln("mulsd %xmm1, %xmm0")
	case ast.Division:
		g.writeln("divsd %xmm1, %xmm0")
	case ast.Equality:
		g.writeln("ucomisd %xmm1, %xmm0")
		g.writeln("sete %al")
		g.writeln("movzbq %al, %rax")
	case ast.Inequality:
		g.writeln("ucomisd %xmm1, %xmm0")
		g.writeln("setne %al")
		g.writeln("movzbq %al, %rax")
	case ast.Less:
		g.writeln("ucomisd %xmm1, %xmm0")
		g.writeln("setb %al")
		g.writeln("movzbq %al, %rax")
	case ast.Greater:
		g.writeln("ucomisd %xmm1, %xmm0")
		g.writeln("seta %al")
		g.writeln("movzbq %al, %rax")
	case ast.LessEqual:
		g.writeln("ucomisd %xmm1, %xmm0")
		g.writeln("setbe %al")
		g.writeln("movzbq %al, %rax")
	case ast.GreaterEqual:
		g.writeln("ucomisd %xmm1, %xmm0")
		g.writeln("setae %al")
		g.writeln("movzbq %al, %rax")
	default:
		panic(fmt.Sprintf("float operator %s not implemented", o.String()))
	}
	return nil
}

// VisitBinary evaluates left to %rax (pushed), right to %rax (moved to %rbx),
// pops left back to %rax, then applies the operator. (it does the same for floats
// by storing the intermediate value on the stack and using %xmm0 and %xmm1)
func (g *Generator) VisitBinary(u *ast.Binary) error {
	g.writeln("# binary")
	if u.Left.GetType().Equals(ast.BasicTypePtr(ast.Float)) {
		if err := u.Left.Accept(g); err != nil {
			return err
		}
		g.writeln("sub $8, %rsp")
		g.writeln("movsd %xmm0, (%rsp)") // save left
		if err := u.Right.Accept(g); err != nil {
			return err
		}
		g.writeln("movsd %xmm0, %xmm1")  // right to %xmm1
		g.writeln("movsd (%rsp), %xmm0") // left to %xmm0
		g.writeln("add $8, %rsp")
		return g.generateFloatBinaryOperator(u.Operator)
	} else {
		if err := u.Left.Accept(g); err != nil {
			return err
		}
		g.writeln("push %rax")
		if err := u.Right.Accept(g); err != nil {
			return err
		}
		g.writeln("mov %rax, %rbx")
		g.writeln("pop %rax")
		return g.generateBinaryOperator(u.Operator)
	}
}

func (g *Generator) VisitBlock(b *ast.Block) error {
	g.writeln("# block")
	for _, expr := range b.Body {
		if err := expr.Accept(g); err != nil {
			return err
		}
	}
	if b.ImplicitReturn != nil {
		return b.ImplicitReturn.Accept(g)
	}
	g.writeln("mov $0, %rax")
	return nil
}

func (g *Generator) VisitCondition(c *ast.Condition) error {
	g.writeln("# condition")
	ifLabel := g.label()
	elseLabel := g.label()
	endLabel := g.label()

	if err := c.Condition.Accept(g); err != nil {
		return err
	}
	g.writeln("cmp $1, %rax")
	g.writefln("je %s", ifLabel)
	if c.Else != nil {
		g.writefln("jmp %s", elseLabel)
	} else {
		g.writefln("jmp %s", endLabel)
	}
	g.writefln("%s:", ifLabel)
	if err := c.Body.Accept(g); err != nil {
		return err
	}
	if c.Else != nil {
		g.writefln("jmp %s", endLabel)
		g.writefln("%s:", elseLabel)
		if err := c.Else.Accept(g); err != nil {
			return err
		}
	}
	g.writefln("%s:", endLabel)
	return nil
}

// containerLoad loads %rcx with the base address of the indexed container:
// a pointer value for slices and array arguments, or an effective address
// for locally-allocated arrays.
func (g *Generator) containerLoad(id *ast.Identifier) error {
	offset, exists := g.ctx.locals[id.Resolved]
	if !exists {
		return generatorError(id.Position, "unresolved identifier %q", id.Name)
	}
	_, isSlice := id.Resolved.GetType().(*ast.SliceType)
	_, isArray := id.Resolved.GetType().(*ast.ArrayType)
	if isSlice || (isArray && g.isArgument(id.Resolved)) {
		g.writefln("mov -%d(%%rbp), %%rcx", offset)
	} else {
		g.writefln("lea -%d(%%rbp), %%rcx", offset)
	}
	return nil
}

// VisitAssignment generates code for assigning a value to a scalar or indexed target.
func (g *Generator) VisitAssignment(a *ast.Assignment) error {
	g.writeln("# assignment")
	switch target := a.Target.(type) {
	case *ast.Identifier:
		offset, exists := g.ctx.locals[target.Resolved]
		if !exists {
			return generatorError(target.Position, "unresolved identifier %q", target.Name)
		}
		switch t := target.Resolved.GetType().(type) {
		case *ast.ArrayType:
			if lit, ok := a.Value.(*ast.Literal); ok && lit.Value == "0" {
				g.writeln("# array zero-assignment")
				g.zeroArray(offset, (t.Size()+7)/8)
			} else if arrLit, ok := a.Value.(*ast.ArrayLiteral); ok {
				return g.generateArrayLiteralInit(arrLit, offset)
			} else {
				if err := a.Value.Accept(g); err != nil {
					return err
				}
				g.writeln("# array copy assignment")
				g.copyArray(offset, (t.Size()+7)/8)
			}
		case *ast.SliceType:
			if err := a.Value.Accept(g); err != nil {
				return err
			}
			g.storeSlice(offset)
		default: // scalar type
			if err := a.Value.Accept(g); err != nil {
				return err
			}
			if t.Equals(ast.BasicTypePtr(ast.Float)) {
				g.storeFloatScalar(offset)
			} else {
				g.storeScalar(offset)
			}
		}
	case *ast.Index:
		if err := a.Value.Accept(g); err != nil {
			return err
		}
		g.pushValue(a.Value.GetType()) // save value

		if err := target.Index.Accept(g); err != nil {
			return err
		}
		g.pushValue(target.Index.GetType()) // save index

		if err := g.containerLoad(target.Identifier); err != nil {
			return err
		}

		g.writeln("pop %rdx")         // index
		g.popValue(a.Value.GetType()) // value

		if a.Value.GetType().Equals(ast.BasicTypePtr(ast.Float)) {
			g.writefln("mov %%xmm0, (%%rcx, %%rdx, %d)", target.GetType().Size())
		} else {
			g.writefln("mov %%rax, (%%rcx, %%rdx, %d)", target.GetType().Size())
		}

	case *ast.Dereference:
		// evaluate value
		if err := a.Value.Accept(g); err != nil {
			return err
		}
		// save scalar value
		g.pushValue(a.Value.GetType())
		// evaluate pointer to %rax
		if err := target.Value.Accept(g); err != nil {
			return err
		}

		if a.Value.GetType().Equals(ast.BasicTypePtr(ast.Float)) {
			g.writeln("movsd (%rsp), %xmm0")
			g.writeln("add $8, %rsp")
			g.writeln("movq %xmm0, (%rax)")
		} else {
			g.writeln("pop %rbx")
			g.writeln("mov %rbx, (%rax)")
		}

	default:
		return generatorError(a.Position, "invalid assignment target")
	}
	return nil
}

func (g *Generator) VisitDereference(d *ast.Dereference) error {
	if err := d.Value.Accept(g); err != nil {
		return err
	}

	if d.GetType().Equals(ast.BasicTypePtr(ast.Float)) {
		g.writeln("movsd (%rax), %xmm0")
	} else {
		g.writeln("mov (%rax), %rax")
	}

	return nil
}

func (g *Generator) VisitLoop(l *ast.Loop) error {
	g.writeln("# for loop")
	startLabel := g.label()
	endLabel := g.label()

	g.writeln("mov $0, %rax") // default if body never runs
	g.writefln("%s:", startLabel)
	g.writeln("push %rax") // save last result before condition clobbers it
	if err := l.Condition.Accept(g); err != nil {
		return err
	}
	g.writeln("cmp $1, %rax")
	g.writefln("jne %s", endLabel) // exit with saved value on stack
	g.writeln("pop %rax")          // discard saved - body will produce new value
	if err := l.Body.Accept(g); err != nil {
		return err
	}
	g.writefln("jmp %s", startLabel)
	g.writefln("%s:", endLabel)
	g.writeln("pop %rax") // restore last body result (or 0 for no iterations)
	return nil
}

func (g *Generator) VisitMake(m *ast.Make) error {
	if err := m.Length.Accept(g); err != nil {
		return err
	}
	// save length
	g.writeln("push %rax")
	// multiply by element size
	g.writefln("imul $%d, %%rax", m.Type.Size())
	// call malloc
	g.writeln("mov %rax, %rdi")
	g.writeln("call malloc@PLT")
	g.writeln("pop %rbx") // length to %rbx
	return nil
}

func (g *Generator) VisitRelease(r *ast.Release) error {
	offset := g.ctx.locals[r.Value.Resolved]
	g.loadSlice(offset) // ptr in %rax, length in %rbx
	g.writeln("mov %rax, %rdi")
	g.writeln("call free@PLT")
	return nil
}

func (g *Generator) generateArrayLiteralInit(a *ast.ArrayLiteral, baseOffset int) error {
	g.writefln("# array literal init at -%d(%%rbp)", baseOffset)
	if len(a.Values) == 0 {
		return generatorError(a.GetPosition(), "unexpected empty array literal ")
	}
	elementSize := a.Values[0].GetType().Size()
	moveInstruction := "mov %rax"
	if a.Values[0].GetType().Equals(ast.BasicTypePtr(ast.Float)) {
		moveInstruction = "movq %xmm0"
	}

	for i, val := range a.Values {
		if err := val.Accept(g); err != nil {
			return err
		}
		g.writefln("%s, -%d(%%rbp)", moveInstruction, baseOffset-(i*elementSize))
	}
	return nil
}

// VisitArrayLiteral writes the literal into its pre-allocated stack space and
// returns pointer to %rax, element count to %rbx.
func (g *Generator) VisitArrayLiteral(a *ast.ArrayLiteral) error {
	offset, exists := g.ctx.locals[a]
	if !exists {
		return generatorError(a.Position, "anonymous array literal has no stack space")
	}
	if err := g.generateArrayLiteralInit(a, offset); err != nil {
		return err
	}
	g.loadArrayAddr(offset)
	g.writefln("mov $%d, %%rbx", len(a.Values))
	return nil
}

// VisitIndex generates an indexed load, leaving the element value in %rax or %xmm0.
func (g *Generator) VisitIndex(i *ast.Index) error {
	g.writeln("# index")
	if err := i.Index.Accept(g); err != nil {
		return err
	}
	g.writeln("push %rax")
	if err := g.containerLoad(i.Identifier); err != nil {
		return err
	}
	g.writeln("pop %rdx")
	if i.GetType().Equals(ast.BasicTypePtr(ast.Float)) {
		g.writefln("movsd (%%rcx, %%rdx, %d), %%xmm0", i.GetType().Size())
	} else {
		g.writefln("mov (%%rcx, %%rdx, %d), %%rax", i.GetType().Size())
	}
	return nil
}
