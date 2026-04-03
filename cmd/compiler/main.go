package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

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

func writeFile(path, content string) {
	if err := os.WriteFile(path, []byte(content), os.ModePerm); err != nil {
		fail(fmt.Errorf("could not write file %q: %v", path, err))
	}
}

func main() {
	inputPath := flag.String("i", "", "input source file (required)")
	execFile := flag.String("o", "", "output executable")
	run := flag.Bool("r", false, "run after compilation")
	help := flag.Bool("h", false, "show help")
	dumpAssembly := flag.String("s", "", "write generated assembly to file")
	dumpTokens := flag.String("t", "", "write token dump to file")
	dumpAst := flag.String("a", "", "write AST dot graph to file")
	flag.Parse()

	if *inputPath == "" || *help {
		flag.Usage()
		os.Exit(1)
	}

	src, err := lexer.ReadFile(*inputPath)
	if err != nil {
		fail(fmt.Errorf("could not read %q: %v", *inputPath, err))
	}

	l := lexer.New(*src)
	tokens, err := l.Lex()
	if err != nil {
		fail(fmt.Errorf("lex: %v", err))
	}

	if *dumpTokens != "" {
		var out strings.Builder
		for _, token := range tokens {
			out.WriteString(token.String() + "\n")
		}
		writeFile(*dumpTokens, out.String())
		fmt.Printf("tokens written to %q\n", *dumpTokens)
	}

	program, err := parser.New(tokens).Parse()
	if err != nil {
		fail(err)
	}

	program, err = name_resolver.NewResolver(program).ResolveNames()
	if err != nil {
		fail(err)
	}

	program, err = type_resolver.NewResolver(program).ResolveTypes()
	if err != nil {
		fail(err)
	}

	program, err = type_checker.NewChecker(program).CheckTypes()
	if err != nil {
		fail(err)
	}

	if *dumpAst != "" {
		graph, err := ast_visualizer.New(program).Visualize()
		if err != nil {
			fail(err)
		}
		writeFile(*dumpAst, graph)
		fmt.Printf("AST written to %q\n", *dumpAst)
	}

	assembly, err := code_generator.New(program).Generate()
	if err != nil {
		fail(err)
	}

	if *dumpAssembly != "" {
		writeFile(*dumpAssembly, assembly)
		fmt.Printf("assembly written to %q\n", *dumpAssembly)
	}

	if *execFile != "" || *run {
		asmFile, err := os.CreateTemp("", "ilang-*.s")
		if err != nil {
			fail(fmt.Errorf("could not create temp file: %v", err))
		}
		defer os.Remove(asmFile.Name())

		if _, err := asmFile.WriteString(assembly); err != nil {
			fail(fmt.Errorf("could not write assembly: %v", err))
		}
		asmFile.Close()

		outFile := *execFile
		if outFile == "" {
			outFile = "a.out"
		}

		gcc := exec.Command("gcc", "-no-pie", "-o", outFile, asmFile.Name(), "-lm")
		gcc.Stdout = os.Stdout
		gcc.Stderr = os.Stderr
		if err := gcc.Run(); err != nil {
			fail(fmt.Errorf("gcc: %v", err))
		}
		fmt.Printf("compiled to %q\n", outFile)

		if *run {
			cmd := exec.Command("./" + outFile)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fail(fmt.Errorf("run: %v", err))
			}
		}
	}
}
