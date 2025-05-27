package main

import (
	"fmt"
	"io"
	"lang/pkg/generator"
	"lang/pkg/lexer"
	"lang/pkg/parser"
	"os"
)

func printUsage() {
	fmt.Printf("Usage: %s <source> <output>\n", os.Args[0])
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
		fmt.Printf("Error: %s\n", err)
		os.Exit(-1)
	}
	defer source_file.Close()

	input, err := io.ReadAll(source_file)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(-1)
	}

	l := lexer.NewLexer(string(input))
	if err := l.Lex(); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(-1)
	}

	p := parser.NewParser(l.Tokens())

	g := generator.NewGenerator(p.Parse())

	//g.Generate()

	err = os.WriteFile(output_file_name, []byte(g.Output), 0644)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(-1)
	}

	fmt.Printf("Successfully generated assembly")
}
