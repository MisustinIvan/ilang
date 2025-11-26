/*
Implements a lexer for the language according to the grammar.

The lexer consumes a SourceFile and emits a slice of Tokens.
*/
package lexer

import (
	"fmt"
	"io"
	"os"
)

type SourceFile struct {
	filename string
	content  string
}

func ReadFile(path string) (*SourceFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return &SourceFile{
		filename: path,
		content:  string(content),
	}, nil
}

type Lexer struct {
	source     SourceFile
	source_len int
	head       int
	line       int
	column     int
	output     []Token
}

func New(source SourceFile) *Lexer {
	return &Lexer{
		source:     source,
		source_len: len(source.content),
		head:       0,
		line:       1,
		column:     1,
		output:     []Token{},
	}
}

func (l *Lexer) currentPos() Position {
	return Position{
		File:   l.source.filename,
		Line:   l.line,
		Column: l.column,
	}
}

func (l *Lexer) headInBounds() bool {
	return l.head < l.source_len
}

func (l *Lexer) current() byte {
	if !l.headInBounds() {
		return 0
	}
	return l.source.content[l.head]
}

func (l *Lexer) next() {
	l.head++
	l.column++
}

func (l *Lexer) newLine() {
	l.column = 1
	l.line++
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\n' || c == '\t' || c == '\r'
}

func isDigit(x byte) bool {
	return x >= '0' && x <= '9'
}

func isLetter(x byte) bool {
	return (x >= 'a' && x <= 'z') || (x >= 'A' && x <= 'Z') || x == '_'
}

func (l *Lexer) Lex() ([]Token, error) {
	for l.headInBounds() {
		c := l.current()

		// whitespaces
		if isWhitespace(c) {
			l.next()
			if c == '\n' {
				l.newLine()
			}
			continue
		}

		// integer literals
		if isDigit(c) {
			startPos := l.currentPos()
			start := l.head
			for l.head < l.source_len && isDigit(l.current()) {
				l.next()
			}
			value := l.source.content[start:l.head]
			l.output = append(l.output, Token{
				Kind:     Literal,
				Value:    value,
				Position: startPos,
			})
			continue
		}

		// string literals
		if c == '"' {
			startPos := l.currentPos()
			start := l.head
			l.next() // consume opening quote

			for l.headInBounds() && l.current() != '"' {
				if l.current() == '\n' {
					l.newLine()
				}
				l.next()
			}

			if !l.headInBounds() {
				return nil, fmt.Errorf("Unterminated string literal beginning at %v", startPos)
			}

			l.next() // consume closing quote

			value := l.source.content[start:l.head]
			l.output = append(l.output, Token{
				Kind:     Literal,
				Value:    value,
				Position: startPos,
			})
			continue
		}

		// keywords and identifiers (and two literals for booleans)
		if isLetter(c) {
			startPos := l.currentPos()
			start := l.head
			for l.headInBounds() && (isLetter(l.current()) || isDigit(l.current())) {
				l.next()
			}

			value := l.source.content[start:l.head]
			kind := Identifier
			if KeywordTokens[value] {
				kind = Keyword
			}

			if value == "true" || value == "false" {
				kind = Literal
			}

			l.output = append(l.output, Token{
				Kind:     kind,
				Value:    value,
				Position: startPos,
			})
			continue
		}

		// punctuators
		if PunctuatorTokens[string(c)] {
			l.output = append(l.output, Token{
				Kind:     Punctuator,
				Value:    string(c),
				Position: l.currentPos(),
			})
			l.next()
			continue
		}

		// operators
		{
			startPos := l.currentPos()
			start := l.head

			for l.headInBounds() && !isDigit(l.current()) && !isWhitespace(l.current()) && !isLetter(l.current()) {
				l.next()
			}

			value := l.source.content[start:l.head]

			if OperatorTokens[value] {
				l.output = append(l.output, Token{
					Kind:     Operator,
					Value:    value,
					Position: startPos,
				})
				continue
			}

			return nil, fmt.Errorf("Unexpected token %q at %v", value, startPos)
		}
	}

	return l.output, nil
}
