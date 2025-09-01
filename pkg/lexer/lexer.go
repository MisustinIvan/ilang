package lexer

import (
	"fmt"
)

type TokenKind int

func (k *TokenKind) String() string {
	s := "UNKNOWN"

	switch *k {
	case Literal:
		s = "Literal"
	case Keyword:
		s = "Keyword"
	case Operator:
		s = "Operator"
	case Identifier:
		s = "Identifier"
	case Punctuator:
		s = "Punctuator"
	}

	return s
}

const (
	Literal TokenKind = iota
	Keyword
	Operator
	Identifier
	Punctuator
)

type Position struct {
	File   string
	Line   int
	Column int
}

func (p *Position) String() string {
	return fmt.Sprintf("%s:%d:%d:", p.File, p.Line+1, p.Column)
}

type Token struct {
	Kind     TokenKind
	Position Position
	Value    string
}

func (t Token) String() string {
	return fmt.Sprintf("<%s> <%s>", t.Kind.String(), t.Value)
}

type Lexer struct {
	input_filename string
	input          string
	head           int
	output         []Token
}

func NewLexer(input_filename string, input string) Lexer {
	return Lexer{
		input_filename: input_filename,
		input:          input,
		head:           0,
		output:         []Token{},
	}
}

func (l Lexer) pos(line int, column int) Position {
	return Position{
		File:   l.input_filename,
		Line:   line,
		Column: column,
	}
}

func isWhitespace(x byte) bool {
	return x == ' ' || x == '\n' || x == '\t' || x == '\r'
}

func isDigit(x byte) bool {
	return x >= '0' && x <= '9'
}

func isLetter(x byte) bool {
	return (x >= 'a' && x <= 'z') || (x >= 'A' && x <= 'Z') || x == '_'
}

type UnexpectedTokenError struct {
	Token string
}

func (e UnexpectedTokenError) Error() string {
	return fmt.Sprintf("Unexpected token: \"%s\" []byte%v", e.Token, []byte(e.Token))
}

func NewUnexpectedTokenError(token string) error {
	return &UnexpectedTokenError{Token: token}
}

type UnterminatedStringLiteralError struct {
	literal string
}

func (e *UnterminatedStringLiteralError) Error() string {
	return fmt.Sprintf("Unterminated string literal: \"%s\"", e.literal)
}

func NewUnterminatedStringLiteral(literal string) error {
	return &UnterminatedStringLiteralError{literal: literal}
}

func longestMatch(token string, valid map[string]bool) (string, int) {
	match := ""
	length := 0
	for i := 1; i <= len(token); i++ {
		candidate := token[:i]
		if valid[candidate] {
			match = candidate
			length = i
		}
	}

	return match, length
}

func (l *Lexer) Lex() ([]Token, error) {
	line := 0
	column := 0
	for l.head < len(l.input) {
		column++
		c := l.input[l.head]

		// whitespaces
		if isWhitespace(c) {
			l.head++
			if c == '\n' {
				line++
				column = 0
			}
			continue
		}

		// integer literals
		if isDigit(c) {
			start := l.head
			start_column := column
			for l.head < len(l.input) && isDigit(l.input[l.head]) {
				l.head++
				column++
			}
			value := l.input[start:l.head]
			l.output = append(l.output, Token{
				Kind:     Literal,
				Value:    value,
				Position: l.pos(line, start_column),
			})
			continue
		}

		// string literals
		if c == '"' {
			start := l.head
			start_column := column
			l.head++ // prevent infinite loop
			for l.head < len(l.input) && l.input[l.head] != '"' {
				l.head++
				column++
			}
			l.head++ //include end quote
			if l.head >= len(l.input) {
				return nil, NewUnterminatedStringLiteral(l.input[:start])
			}
			value := l.input[start:l.head]
			l.output = append(l.output, Token{
				Kind:     Literal,
				Value:    value,
				Position: l.pos(line, start_column),
			})
			continue
		}

		// keywords and identifiers (and two literals for booleans)
		if isLetter(c) {
			start := l.head
			start_column := column
			for l.head < len(l.input) && (isLetter(l.input[l.head]) || isDigit(l.input[l.head])) {
				l.head++
				column++
			}
			value := l.input[start:l.head]
			kind := Identifier
			if KeywordTokens[value] {
				kind = Keyword
			}

			// check for booleans
			if value == "true" || value == "false" {
				kind = Literal
			}

			l.output = append(l.output, Token{
				Kind:     kind,
				Value:    value,
				Position: l.pos(line, start_column),
			})
			continue
		}

		// punctuators and operators
		{
			start := l.head
			start_column := column
			for l.head < len(l.input) && !isDigit(l.input[l.head]) && !isLetter(l.input[l.head]) && !isWhitespace(l.input[l.head]) {
				l.head++
				column++
			}
			value := l.input[start:l.head]
			var kind TokenKind

			operator_match, operator_length := longestMatch(value, OperatorTokens)
			punctuator_match, punctuator_length := longestMatch(value, PunctuatorTokens)

			if operator_length == 0 && punctuator_length == 0 {
				return nil, NewUnexpectedTokenError(value)
			}

			if operator_length > punctuator_length {
				l.head -= len(value) - len(operator_match)
				column -= len(value) - len(operator_match)
				value = operator_match
				kind = Operator
			} else {
				l.head -= len(value) - len(punctuator_match)
				column -= len(value) - len(punctuator_match)
				value = punctuator_match
				kind = Punctuator
			}

			l.output = append(l.output, Token{
				Kind:     kind,
				Value:    value,
				Position: l.pos(line, start_column),
			})
			continue
		}
	}

	return l.output, nil
}
