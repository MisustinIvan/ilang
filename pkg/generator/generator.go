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
}

func NewGenerator(prog ast.Prog) Generator {
	return Generator{
		Program:      prog,
		Output:       "",
		Declarations: map[string]FunctionContext{},
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
	localScopes        []LocalVariableScope
	uniqueCounter      int
	localVariableCount int
}

type LocalVariableScope map[string]LocalVariableContext

func (ctx *FunctionContext) NewLocalVariableIndex() int {
	v := ctx.localVariableCount
	ctx.localVariableCount++
	return v
}

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
		Index:    ctx.NewLocalVariableIndex(),
		UniqueId: uid,
	}

	(*current_scope)[name] = v
	ctx.LocalVariables[uid] = v

	return v
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
	Index    int
	UniqueId string
}

func (g *Generator) RegisterGlobalFunction(f ast.FunctionDeclaration) {
	_, exists := g.Declarations[f.Name.Value]
	if exists {
		fmt.Printf("Redefinition of function %s\n", f.Name.Value)
	}

	var param_types []ast.Type
	for _, param := range f.Parameters {
		param_types = append(param_types, param.Type)
	}

	ctx := FunctionContext{
		Generator:          g,
		Name:               f.Name.Value,
		ReturnType:         f.Type,
		ParameterTypes:     param_types,
		LocalVariables:     LocalVariableScope{},
		StackOffset:        0,
		localScopes:        []LocalVariableScope{},
		uniqueCounter:      0,
		localVariableCount: 0,
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

func (g *Generator) CalculateStackOffset(d ast.FunctionDeclaration) {
	ctx := g.Declarations[d.Name.Value]

	total_size := 0
	for _, l := range ctx.LocalVariables {
		total_size += l.Type.Size()
	}
	ctx.StackOffset = total_size

	g.Declarations[d.Name.Value] = ctx
}

func (g *Generator) CalculateStackOffsets() {
	for _, d := range g.Program.Declarations {
		g.CalculateStackOffset(d)
	}
}

func (g *Generator) Generate() ast.Prog {
	g.RegisterGlobalFunctions()
	g.ProcessFunctionsLocals()
	g.CalculateStackOffsets()

	return g.Program
}
