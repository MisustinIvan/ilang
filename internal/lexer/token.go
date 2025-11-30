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

const KeywordLet = "let"
const KeywordIf = "if"
const KeywordElse = "else"
const KeywordReturn = "return"
const KeywordExtrn = "extrn"

var KeywordTokens = map[string]bool{
	KeywordLet:    true,
	KeywordIf:     true,
	KeywordElse:   true,
	KeywordReturn: true,
	KeywordExtrn:  true,
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
