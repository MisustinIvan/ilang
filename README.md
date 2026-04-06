# Ilang

A simple, statically typed programming language implemented in Go targeted for x86_64.

## Build
```bash
go build -o ilang-compiler cmd/compiler/main.go
```

## Usage
Run a program directly:
```bash
./ilang-compiler -i examples/game_of_life.ilang -r
```

Generate assembly:
```bash
./ilang-compiler -i examples/mandelbrot.ilang -s mandelbrot.s
```

Further documentation available in [`docs/docs.pdf`](./docs/docs.pdf)
