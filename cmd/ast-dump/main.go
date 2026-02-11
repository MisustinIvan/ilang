package main

import (
	"fmt"
	"log"
	"os"

	"github.com/MisustinIvan/ilang/internal/ast_visualizer"
	"github.com/MisustinIvan/ilang/internal/lexer"
	"github.com/MisustinIvan/ilang/internal/parser"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: ast-dump <file>")
		return
	}

	filepath := os.Args[1]
	source, err := lexer.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	l := lexer.New(*source)
	tokens, err := l.Lex()
	if err != nil {
		log.Fatalf("Lexer error: %s\n", err)
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		log.Fatalf("Parser error: %s\n", err)
	}

	visualizer := ast_visualizer.New(program)
	output, err := visualizer.Visualize()
	if err != nil {
		log.Fatalf("Visualizer error: %s\n", err)
	}

	fmt.Println(output)
}
