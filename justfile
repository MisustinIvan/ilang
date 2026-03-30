# Run go tests
test:
	go test -v ./...

alias ad := ast-dump
# Dump the ast of a given file in the ./examples directory, generate an image and open it
ast-dump example='test.ilang':
	go run cmd/compiler/main.go -i ./examples/{{example}} -ast example.dot
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
	go run cmd/compiler/main.go -i ./examples/{{example}} -tk example.txt

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
	rm example.dot
	rm graph.png
	rm example
	rm example.s
	rm example.txt

# Generate the treesitter grammar into ./tree-sitter-grammar
generate-ts-parser:
	cd ./tree-sitter-grammar && tree-sitter generate
