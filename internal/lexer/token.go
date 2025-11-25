package lexer

import (
	"fmt"
)

//go:generate stringer -type=TokenKind
type TokenKind int

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
	return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Column)
}

type Token struct {
	Kind     TokenKind
	Value    string
	Position Position
}

func (t *Token) String() string {
	return fmt.Sprintf("%s %s @ %s", t.Kind, t.Value, t.Position.String())
}

var KeywordTokens = map[string]bool{
	"let":    true,
	"if":     true,
	"else":   true,
	"return": true,
	"extrn":  true,
}

var PunctuatorTokens = map[string]bool{
	"(": true,
	")": true,
	"{": true,
	"}": true,
	";": true,
	":": true,
	",": true,
}

var OperatorTokens = map[string]bool{
	"=":  true,
	"+":  true,
	"-":  true,
	"!":  true,
	"*":  true,
	"/":  true,
	"==": true,
	"!=": true,
	"<":  true,
	">":  true,
	"<=": true,
	">=": true,
	"<<": true,
	">>": true,
	"&&": true,
	"||": true,
}
