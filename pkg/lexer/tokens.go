package lexer

var PunctuatorTokens = map[string]bool{
	"(": true,
	")": true,
	"[": true,
	"]": true,
	"{": true,
	"}": true,
	";": true,
	":": true,
	",": true,
	".": true,
}

var KeywordTokens = map[string]bool{
	"return": true,
	"let":    true,
	"if":     true,
	"else":   true,
	"for":    true,
	"break":  true,
	"extern": true,
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
