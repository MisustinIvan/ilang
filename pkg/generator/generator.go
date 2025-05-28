package generator

import (
	"fmt"
	"lang/pkg/ast"
	"os"
)

type Generator struct {
	Program      ast.Prog
	Output       string
	Declarations map[string]FunctionContext
	// constants
	Constants       map[string]ConstantContext
	constantCounter int
	// conditionals
	conditionalLabelCounter int
}

func (g *Generator) NextConditionalLabel() string {
	v := g.conditionalLabelCounter
	g.conditionalLabelCounter++
	return fmt.Sprintf(".conditional_label_%d", v)
}

type ConstantContext struct {
	Type        ast.Type
	Value       string
	LookupValue string
}

func (g *Generator) NewUniqueConstantId() string {
	c := g.constantCounter
	g.constantCounter++
	return fmt.Sprintf("constant#%d", c)
}

func NewGenerator(prog ast.Prog) Generator {
	return Generator{
		Program:      prog,
		Output:       "",
		Declarations: map[string]FunctionContext{},
		Constants:    map[string]ConstantContext{},
	}
}

type FunctionContext struct {
	Generator *Generator
	// metadata for ast validation
	Name           string
	ReturnType     ast.Type
	ParameterTypes []ast.Type
	// local variables
	LocalVariables LocalVariableScope
	StackOffset    int
	// helpers for scope management
	localScopes   []LocalVariableScope
	uniqueCounter int
}

type LocalVariableScope map[string]LocalVariableContext

func (ctx *FunctionContext) NewUniqueId() int {
	i := ctx.uniqueCounter
	ctx.uniqueCounter++
	return i
}

func (ctx *FunctionContext) PushScope() {
	ctx.localScopes = append(ctx.localScopes, LocalVariableScope{})
}

func (ctx *FunctionContext) PopScope() {
	if len(ctx.localScopes) == 0 {
		fmt.Printf("No scope to pop\n")
		os.Exit(-1)
	}
	ctx.localScopes = ctx.localScopes[:len(ctx.localScopes)-1]
}

func (ctx *FunctionContext) CurrentScope() *LocalVariableScope {
	return &ctx.localScopes[len(ctx.localScopes)-1]
}

func (ctx *FunctionContext) RegisterLocal(name string, l_type ast.Type) LocalVariableContext {
	current_scope := ctx.CurrentScope()

	// check if variable name does not collide with global function name
	if _, collides := ctx.Generator.Declarations[name]; collides {
		fmt.Printf("Local variable name: %s collides with global function name\n", name)
		os.Exit(-1)
	}

	if _, exists := (*current_scope)[name]; exists {
		fmt.Printf("Redefinition of variable %s\n", name)
		os.Exit(-1)
	}

	uid := fmt.Sprintf("%s#%d", name, ctx.NewUniqueId())

	v := LocalVariableContext{
		Name:     name,
		Type:     l_type,
		Offset:   ctx.IncrementStackOffset(l_type),
		UniqueId: uid,
	}

	(*current_scope)[name] = v
	ctx.LocalVariables[uid] = v

	return v
}

func (ctx *FunctionContext) IncrementStackOffset(v_type ast.Type) int {
	ctx.StackOffset += v_type.Size()
	return -ctx.StackOffset
}

func (ctx *FunctionContext) Resolve(name string) LocalVariableContext {
	for i := len(ctx.localScopes) - 1; i >= 0; i-- {
		if val, ok := ctx.localScopes[i][name]; ok {
			return val
		}
	}
	fmt.Printf("Undefined local variable %s\n", name)
	os.Exit(-1)
	return LocalVariableContext{}
}

type LocalVariableContext struct {
	Name     string
	Type     ast.Type
	Offset   int
	UniqueId string
}

func (g *Generator) AllocateConstant(value string, v_type ast.Type) ConstantContext {
	id := g.NewUniqueConstantId()
	ctx := ConstantContext{
		Type:        v_type,
		Value:       value,
		LookupValue: id,
	}
	g.Constants[id] = ctx
	return ctx
}

func (g *Generator) RegisterGlobalFunction(f ast.FunctionDeclaration) {
	_, exists := g.Declarations[f.Name.Value]
	if exists {
		fmt.Printf("Redefinition of function %s\n", f.Name.Value)
		os.Exit(-1)
	}

	var param_types []ast.Type
	for _, param := range f.Parameters {
		param_types = append(param_types, param.Type)
	}

	ctx := FunctionContext{
		Generator:      g,
		Name:           f.Name.Value,
		ReturnType:     f.Type,
		ParameterTypes: param_types,
		LocalVariables: LocalVariableScope{},
		StackOffset:    0,
		localScopes:    []LocalVariableScope{},
		uniqueCounter:  0,
	}

	g.Declarations[f.Name.Value] = ctx
}

func (g *Generator) RegisterGlobalFunctions() {
	for _, d := range g.Program.Declarations {
		g.RegisterGlobalFunction(d)
	}
}

func (g *Generator) ResolveIdentifiers(expr ast.Expression, ctx *FunctionContext, top_level bool) {
	switch e := expr.(type) {
	case *ast.Identifier:
		// resolve the identifier
		resolved := ctx.Resolve(e.Value)
		e.LookupValue = resolved.UniqueId
	case *ast.SeparatedExpression:
		g.ResolveIdentifiers(e.Value, ctx, false)
	case *ast.ReturnExpression:
		g.ResolveIdentifiers(e.Value, ctx, false)
	case *ast.FunctionCall:
		for _, parm := range e.Params {
			g.ResolveIdentifiers(parm, ctx, false)
		}
	case *ast.ConditionalExpression:
		g.ResolveIdentifiers(e.Condition, ctx, false)
		g.ResolveIdentifiers(&e.IfBody, ctx, false)
		g.ResolveIdentifiers(&e.ElseBody, ctx, false)
	case *ast.BlockExpression:
		if top_level {
			for _, b_expr := range e.Body {
				g.ResolveIdentifiers(b_expr, ctx, false)
			}
			if e.ImplicitReturnExpression != nil {
				g.ResolveIdentifiers(e.ImplicitReturnExpression, ctx, false)
			}
		}
	case *ast.BinaryExpression:
		g.ResolveIdentifiers(e.Left, ctx, false)
		g.ResolveIdentifiers(e.Right, ctx, false)
	case *ast.AssignmentExpression:
		g.ResolveIdentifiers(&e.Left, ctx, false)
		g.ResolveIdentifiers(e.Right, ctx, false)
	case *ast.BindExpression:
		g.ResolveIdentifiers(e.Right, ctx, false)
	}
}

func (g *Generator) FindLocals(expr ast.Expression, ctx *FunctionContext) {
	switch e := expr.(type) {
	case *ast.BindExpression:
		// register the local variable
		c := ctx.RegisterLocal(e.Left.Value, e.Type)
		e.Left.LookupValue = c.UniqueId
		g.FindLocals(e.Right, ctx)
	case *ast.AssignmentExpression:
		g.FindLocals(e.Right, ctx)
	case *ast.BinaryExpression:
		g.FindLocals(e.Left, ctx)
		g.FindLocals(e.Right, ctx)
	case *ast.BlockExpression:
		ctx.PushScope()
		for _, be := range e.Body {
			g.FindLocals(be, ctx)
		}
		if e.ImplicitReturnExpression != nil {
			g.FindLocals(e.ImplicitReturnExpression, ctx)
		}
		g.ResolveIdentifiers(e, ctx, true)
		ctx.PopScope()
	case *ast.ConditionalExpression:
		g.FindLocals(e.Condition, ctx)
		g.FindLocals(&e.IfBody, ctx)
		g.FindLocals(&e.ElseBody, ctx)
	case *ast.FunctionCall:
		for _, parm := range e.Params {
			g.FindLocals(parm, ctx)
		}
	case *ast.ReturnExpression:
		g.FindLocals(e.Value, ctx)
	case *ast.SeparatedExpression:
		g.FindLocals(e.Value, ctx)
	case *ast.Literal:
		if e.Type == ast.String {
			c := g.AllocateConstant(e.Value, ast.String)
			e.LookupValue = c.LookupValue
		}
	case *ast.Identifier:
	default:
		fmt.Printf("Unknown expression type\n")
		os.Exit(-1)
	}
}

func (g *Generator) ProcessFunctionLocals(d ast.FunctionDeclaration) {
	ctx := g.Declarations[d.Name.Value]
	ctx.PushScope()

	for _, parm := range d.Parameters {
		ctx.RegisterLocal(parm.Name.Value, parm.Type)
	}

	g.FindLocals(&d.Body, &ctx)

	g.Declarations[d.Name.Value] = ctx
}

func (g *Generator) ProcessFunctionsLocals() {
	for _, d := range g.Program.Declarations {
		g.ProcessFunctionLocals(d)
	}
}

func (g *Generator) EmitPrologue(ctx FunctionContext) string {
	out := ""
	out += ctx.Name + ":\n"
	out += "#   prologue\n"
	out += "    push %rbp\n"
	out += "    mov %rsp, %rbp\n"
	out += "#   stack allocation\n"
	out += fmt.Sprintf("    sub $%d, %%rsp\n", ctx.StackOffset)

	return out
}

func (g *Generator) EmitEpilogue() string {
	out := ""
	out += "#   epilogue\n"
	out += "    leave\n"
	out += "    ret\n"
	return out
}

func Register(n int) string {
	regs := []string{"%rdi", "%rsi", "%rdx", "%rcx", "%r8", "%r9"}
	if n < len(regs) {
		return regs[n]
	}
	fmt.Printf("Invalid register, allowed range is 0..5\n")
	os.Exit(-1)
	return ""
}

func (g *Generator) GenerateMoveFunctionParameters(decl ast.FunctionDeclaration) string {
	out := "#   move function parameters into local stack space\n"

	offset := -8
	for i, param := range decl.Parameters {
		reg := Register(i)
		out += fmt.Sprintf("    mov %s, %d(%%rbp)\n", reg, offset)
		offset -= param.Type.Size()
	}

	return out
}

func (g *Generator) GenerateLiteralExpression(expr *ast.Literal, ctx *FunctionContext) string {
	out := ""
	out += "#   literal expression\n"
	switch expr.Type {
	case ast.Integer:
		out += "    mov $" + expr.Value + ", %rax\n"
	case ast.Boolean:
		val := ""
		if expr.Value == "true" {
			val = "1"
		} else if expr.Value == "false" {
			val = "0"
		} else {
			fmt.Printf("invalid boolean value: %s\n", expr.Value)
			os.Exit(-1)
		}
		out += "    mov $" + val + ", %rax\n"
	case ast.Float:
		fmt.Printf("floats not implemented yet\n")
		os.Exit(-1)
	case ast.String:
		out += "    lea ." + expr.LookupValue + "(%rip), %rax\n"
	case ast.Unit:
	default:
		fmt.Printf("invalid literal type\n")
		os.Exit(-1)
	}
	return out
}

func (g *Generator) GenerateBindExpression(expr *ast.BindExpression, ctx *FunctionContext) string {
	out := ""

	// generate right value and assume its in %rax
	out += g.GenerateExpression(expr.Right, ctx)

	out += "#   bind expression\n"

	out += fmt.Sprintf("    mov %%rax, %d(%%rbp)\n", ctx.LocalVariables[expr.Left.LookupValue].Offset)
	// result stays in %rax, as intended

	return out
}

func (g *Generator) GenerateAssignmentExpression(expr *ast.AssignmentExpression, ctx *FunctionContext) string {
	out := ""

	// generate right value and assume its in %rax
	out += g.GenerateExpression(expr.Right, ctx)

	out += "#   assignment expression\n"
	out += fmt.Sprintf("    mov %%rax, %d(%%rbp)\n", ctx.LocalVariables[expr.Left.LookupValue].Offset)
	// result stays in %rax, as intended

	return out
}

func (g *Generator) GenerateBlockExpression(expr *ast.BlockExpression, ctx *FunctionContext) string {
	out := ""

	out += "#   block expression\n"

	for _, e := range expr.Body {
		out += g.GenerateExpression(e, ctx)
	}

	if expr.ImplicitReturnExpression != nil {
		out += "#   implicit return expression\n"
		out += g.GenerateExpression(expr.ImplicitReturnExpression, ctx)
	} else {
		out += "    mov $0, %rax\n"
	}

	return out
}

// assumes op1 in %rax and op2 in %rbx
func (g *Generator) GenerateBinaryOperator(op ast.BinaryOperator) string {
	switch op {
	case ast.Addition:
		out := ""
		out += "    add %rbx, %rax\n"
		return out
	case ast.Subtraction:
		out := ""
		out += "    sub %rbx, %rax\n"
		return out
	case ast.Multiplication:
		out := ""
		out += "    imul %rbx, %rax\n"
		return out
	case ast.Division:
		out := ""
		out += "    cltd\n"
		out += "    idiv %rbx\n"
		return out
	case ast.Equality:
		out := ""
		out += "    cmp %rax, %rbx\n"
		out += "    sete %al\n"
		out += "    movzbq %al, %rax\n"
		return out
	case ast.Inequality:
		out := ""
		out += "    cmp %rax, %rbx\n"
		out += "    setne %al\n"
		out += "    movzbq %al, %rax\n"
		return out
	case ast.LesserThan:
		out := ""
		out += "    cmp %rax, %rbx\n"
		out += "    setl %al\n"
		out += "    movzbq %al, %rax\n"
		return out
	case ast.GreaterThan:
		out := ""
		out += "    cmp %rax, %rbx\n"
		out += "    setg %al\n"
		out += "    movzbq %al, %rax\n"
		return out
	case ast.LesserOrEqualThan:
		out := ""
		out += "    cmp %rax, %rbx\n"
		out += "    setle %al\n"
		out += "    movzbq %al, %rax\n"
		return out
	case ast.GreaterOrEqualThan:
		out := ""
		out += "    cmp %rax, %rbx\n"
		out += "    setge %al\n"
		out += "    movzbq %al, %rax\n"
		return out
	case ast.LeftShift:
		out := ""
		out += "    shl %rbx, %rax\n"
		return out
	case ast.RightShift:
		out := ""
		out += "    shr %rbx, %rax\n"
		return out
	case ast.LogicAnd:
		out := ""
		out += "    and %rbx, %rax\n"
		return out
	case ast.LogicOr:
		out := ""
		out += "    or %rbx, %rax\n"
	}
	fmt.Printf("operator %s not implemented\n", op.String())
	os.Exit(-1)
	return ""
}

func (g *Generator) GenerateBinaryExpression(expr *ast.BinaryExpression, ctx *FunctionContext) string {
	out := ""
	out += "#   binary expression\n"

	// generate right value
	out += g.GenerateExpression(expr.Right, ctx)
	out += "    push %rax\n"
	out += g.GenerateExpression(expr.Left, ctx)
	out += "    push %rax\n"

	out += "    pop %rax\n"
	out += "    pop %rbx\n"

	out += g.GenerateBinaryOperator(expr.Operator)

	return out
}

func (g *Generator) GenerateIdentifierExpression(expr *ast.Identifier, ctx *FunctionContext) string {
	out := ""
	out += "#   identifier expression\n"

	out += fmt.Sprintf("    mov  %d(%%rbp), %%rax\n", ctx.LocalVariables[expr.LookupValue].Offset)

	return out
}

func (g *Generator) GenerateConditionalExpression(expr *ast.ConditionalExpression, ctx *FunctionContext) string {
	out := ""
	out += "#   conditional expression\n"

	else_label := g.NextConditionalLabel()
	end_label := g.NextConditionalLabel()

	out += g.GenerateExpression(expr.Condition, ctx)
	// this line is very important, as its not possible to do a conditional jump
	// based on a register, instead we must set the flag using the `cmp` instruction
	out += "    cmp $0, %rax\n"
	out += "    je " + else_label + "\n"
	out += g.GenerateExpression(&expr.IfBody, ctx)
	out += "    jmp " + end_label + "\n"
	out += else_label + ":\n"
	out += g.GenerateExpression(&expr.ElseBody, ctx)
	out += end_label + ":\n"

	return out
}

func (g *Generator) GenerateExpression(expr ast.Expression, ctx *FunctionContext) string {
	switch e := expr.(type) {
	case *ast.Literal:
		return g.GenerateLiteralExpression(e, ctx)
	case *ast.BindExpression:
		return g.GenerateBindExpression(e, ctx)
	case *ast.AssignmentExpression:
		return g.GenerateAssignmentExpression(e, ctx)
	case *ast.BlockExpression:
		return g.GenerateBlockExpression(e, ctx)
	case *ast.BinaryExpression:
		return g.GenerateBinaryExpression(e, ctx)
	case *ast.Identifier:
		return g.GenerateIdentifierExpression(e, ctx)
	case *ast.ConditionalExpression:
		return g.GenerateConditionalExpression(e, ctx)
	}
	fmt.Printf("expression not implemented\n")
	os.Exit(-1)
	return ""
}

func (g *Generator) GenerateFunction(decl ast.FunctionDeclaration) {
	ctx := g.Declarations[decl.Name.Value]
	out := ""

	out += g.EmitPrologue(ctx)

	out += g.GenerateMoveFunctionParameters(decl)

	out += "#   function body\n"

	out += g.GenerateExpression(&decl.Body, &ctx)

	out += g.EmitEpilogue()

	g.Output += out
}

func (g *Generator) GenerateFunctions() {
	for _, d := range g.Program.Declarations {
		g.GenerateFunction(d)
	}
}

func (g *Generator) AssertEntryPoint() {
	for _, d := range g.Declarations {
		if d.Name == "main" {
			return
		}
	}
	fmt.Printf("no entry point \"main\"\n")
	os.Exit(-1)
}

func (g *Generator) GenerateHeaders() {
	out := ""
	out += "# headers\n"
	out += ".globl main\n"
	g.Output += out
}

func (g *Generator) Generate() ast.Prog {
	g.RegisterGlobalFunctions()

	g.AssertEntryPoint()

	g.ProcessFunctionsLocals()

	g.GenerateHeaders()

	g.GenerateFunctions()

	//g.GenerateDataSection()

	fmt.Printf("g.Output:\n%v\n", g.Output)

	return g.Program
}
