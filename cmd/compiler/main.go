package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/MisustinIvan/ilang/internal/ast_visualizer"
	"github.com/MisustinIvan/ilang/internal/code_generator"
	"github.com/MisustinIvan/ilang/internal/lexer"
	"github.com/MisustinIvan/ilang/internal/name_resolver"
	"github.com/MisustinIvan/ilang/internal/parser"
	"github.com/MisustinIvan/ilang/internal/type_checker"
	"github.com/MisustinIvan/ilang/internal/type_resolver"
)

func fail(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func main() {
	input_path := flag.String("i", "", "input source code file")
	dump_assembly := flag.String("s", "", "dump generated assembly to provided path")
	dump_tokens := flag.String("tk", "", "dump tokens of the source file to provided path")
	dump_ast := flag.String("ast", "", "dump ast of the source file to provided path")

	flag.Parse()

	if !flag.Parsed() || *input_path == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	src, err := lexer.ReadFile(*input_path)
	if err != nil {
		fail(err)
	}

	lexer := lexer.New(*src)
	tokens, err := lexer.Lex()
	if err != nil {
		fail(err)
	}
	if *dump_tokens != "" {
		file, err := os.Create(*dump_tokens)
		if err != nil {
			fail(err)
		}
		defer file.Close()

		for _, token := range tokens {
			_, err := file.WriteString(token.String() + "\n")
			if err != nil {
				fail(err)
			}
		}
	}

	parser := parser.New(tokens)
	ast, err := parser.Parse()
	if err != nil {
		fail(err)
	}

	resolver := name_resolver.NewResolver(ast)
	ast, err = resolver.ResolveNames()
	if err != nil {
		fail(err)
	}

	type_resolver := type_resolver.NewResolver(ast)
	ast, err = type_resolver.ResolveTypes()
	if err != nil {
		fail(err)
	}

	type_checker := type_checker.NewChecker(ast)
	ast, err = type_checker.CheckTypes()
	if err != nil {
		fail(err)
	}

	if *dump_ast != "" {
		visualizer := ast_visualizer.New(ast)
		graph, err := visualizer.Visualize()
		if err != nil {
			fail(err)
		}

		os.WriteFile(*dump_ast, []byte(graph), os.ModePerm)
	}

	codeGenerator := code_generator.New(ast)
	assembly, err := codeGenerator.Generate()
	if err != nil {
		fail(err)
	}

	if *dump_assembly != "" {
		os.WriteFile(*dump_assembly, []byte(assembly), os.ModePerm)
	}
}
