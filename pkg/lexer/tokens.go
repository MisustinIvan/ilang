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
}

var OperatorTokens = map[string]bool{
	"=":  true,
	"+":  true,
	"-":  true,
	"*":  true,
	"/":  true,
	"<<": true,
	">>": true,
	"&&": true,
	"||": true,
}
