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

alias c := clean
clean:
	rm example.dot
	rm graph.png
