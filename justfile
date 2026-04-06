# Run go tests
test:
	go test -v ./...

alias ad := ast-dump
# Dump the ast of a given file in the ./examples directory, generate an image and open it
ast-dump example='test.ilang':
	go run cmd/compiler/main.go -i ./examples/{{example}} -a example.dot
	dot -Tpng example.dot -o graph.png
	niri msg action focus-workspace 'media'
	sxiv graph.png
	niri msg action focus-workspace 'terminal'

alias as := assembly-dump
# Dump the assembly of a given file to ./example.s
assembly-dump example='test.ilang':
	go run cmd/compiler/main.go -i ./examples/{{example}} -s example.s

alias tk := token-dump
# Dump the tokens of a given file to ./example.txt
token-dump example='test.ilang':
	go run cmd/compiler/main.go -i ./examples/{{example}} -t example.txt

alias r := run
# Compile and run the given source code file from the ./examples directory
run example='test.ilang':
	go run cmd/compiler/main.go -i ./examples/{{example}} -s example.s
	gcc -g -no-pie -o example example.s -lm
	chmod +x example
	./example

alias c := clean
# Clean up generated files
clean:
	rm -f example.dot
	rm -f graph.png
	rm -f example
	rm -f example.s
	rm -f example.txt

alias b := build
build:
	go build -o compiler ./cmd/compiler/main.go

# Generate the treesitter grammar into ./tree-sitter-grammar
generate-ts-parser:
	cd ./tree-sitter-grammar && tree-sitter generate


# Run a brainfuck program from the ./examples/brainfuck directory using an interpreter written in ilang
brainfuck_example program="sierpinski":
	go run cmd/compiler/main.go -i ./examples/brainfuck/brainfuck.ilang -s brainfuck.s
	gcc -g -no-pie -o brainfuck brainfuck.s -lm
	chmod +x brainfuck
	cat ./examples/brainfuck/{{program}}.bf | ./brainfuck

# Run the brainfuck interpreter
brainfuck:
	go run cmd/compiler/main.go -i ./examples/brainfuck/brainfuck.ilang -s brainfuck.s
	gcc -g -no-pie -o brainfuck brainfuck.s -lm
	chmod +x brainfuck
	./brainfuck
