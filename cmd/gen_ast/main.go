package main

import (
	"fmt"
	"io"
	"lang/pkg/ast"
	"lang/pkg/generator"
	"lang/pkg/lexer"
	"lang/pkg/parser"
	"os"
)

func printUsage() {
	fmt.Printf("Usage: %s <program_file> <output_file>", os.Args[0])
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
		fmt.Printf("Error: %s", err)
		os.Exit(-1)
	}
	defer source_file.Close()

	output_file, err := os.OpenFile(output_file_name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(-1)
	}
	defer output_file.Close()

	source_code, err := io.ReadAll(source_file)
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(-1)
	}

	l := lexer.NewLexer(string(source_code))
	err = l.Lex()
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(-1)
	}

	p := parser.NewParser(l.Tokens())
	prog := p.Parse()
	g := generator.NewGenerator(prog)
	prog = g.Generate()

	ast.ExportASTToGraphviz(&prog, output_file)

	fmt.Printf("Successfully saved output to %s", output_file_name)
}
