package main

import (
	"flag"
	"log"
	"os"

	"github.com/MisustinIvan/ilang/internal/ast_visualizer"
	"github.com/MisustinIvan/ilang/internal/lexer"
	"github.com/MisustinIvan/ilang/internal/name_resolver"
	"github.com/MisustinIvan/ilang/internal/parser"
)

func main() {
	input_path := flag.String("i", "", "input source code file")
	dump_tokens := flag.String("tk", "", "dump tokens of the source file to provided path")
	dump_ast := flag.String("ast", "", "dump ast of the source file to provided path")

	flag.Parse()

	if !flag.Parsed() || *input_path == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	src, err := lexer.ReadFile(*input_path)
	if err != nil {
		log.Fatal(err)
	}

	lexer := lexer.New(*src)
	tokens, err := lexer.Lex()
	if err != nil {
		log.Fatal(err)
	}
	if *dump_tokens != "" {
		file, err := os.Create(*dump_tokens)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		for _, token := range tokens {
			_, err := file.WriteString(token.String() + "\n")
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	parser := parser.New(tokens)
	ast, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	resolver := name_resolver.NewResolver(ast)
	ast, err = resolver.ResolveNames()
	if err != nil {
		log.Fatal(err)
	}

	if *dump_ast != "" {
		visualizer := ast_visualizer.New(ast)
		graph, err := visualizer.Visualize()
		if err != nil {
			log.Fatal(err)
		}

		os.WriteFile(*dump_ast, []byte(graph), os.ModePerm)
	}
}
