package main

import (
	"fmt"
	"io"
	"lang/pkg/ast_visualizer"
	"lang/pkg/lexer"
	"lang/pkg/name_resolver"
	"lang/pkg/parser"
	"lang/pkg/type_checker"
	"lang/pkg/type_resolver"
	"log"
	"os"
	"os/exec"
)

func printUsage() {
	fmt.Printf("Usage: %s <program_file> <output_file>(without extension)", os.Args[0])
}

func main() {
	if len(os.Args) != 3 {
		printUsage()
		os.Exit(-1)
	}

	source_file_name := os.Args[1]
	output_file_name := os.Args[2]

	source_file, err := os.Open(source_file_name)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	defer source_file.Close()

	output_file, err := os.OpenFile(output_file_name+".dot", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	defer output_file.Close()

	source_code, err := io.ReadAll(source_file)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	l := lexer.NewLexer(source_file_name, string(source_code))
	tokens, err := l.Lex()
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	p := parser.NewParser(tokens)
	prog, err := p.Parse()
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	name_resolver := name_resolver.NewNameResolver(prog)
	prog, err = name_resolver.ResolveNames()
	if err != nil {
		fmt.Printf("Name Resolution Errors:\n%s\n", err.Error())
	}

	type_resolver := type_resolver.NewTypeResolver(prog)
	prog, err = type_resolver.ResolveTypes()
	if err != nil {
		fmt.Printf("Type Resolution Errors:\n%s\n", err.Error())
	}

	type_checker := type_checker.NewTypeChecker(prog)
	prog, err = type_checker.CheckTypes()
	if err != nil {
		fmt.Printf("Type Check Errors:\n%s\n", err.Error())
	}

	ast_visualizer.ExportASTToGraphviz(prog, output_file)

	fmt.Printf("Successfully saved graph to %s\n", output_file_name)

	cmd := exec.Command("dot", "-Tpng", output_file_name+".dot", "-o", output_file_name+".png")
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}
