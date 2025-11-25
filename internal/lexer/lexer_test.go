package lexer

import (
	"testing"
)

// compareTokens compares two slices of tokens, ignoring the Position field.
func compareTokens(a, b []Token) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Kind != b[i].Kind || a[i].Value != b[i].Value {
			return false
		}
	}
	return true
}

func TestLex(t *testing.T) {
	tests := []struct {
		name          string
		source        SourceFile
		expected      []Token
		expectedError bool
	}{
		{
			name: "Empty",
			source: SourceFile{
				filename: "test.ilang",
				content:  "",
			},
			expected:      []Token{},
			expectedError: false,
		},
		{
			name: "Whitespace",
			source: SourceFile{
				filename: "test.ilang",
				content:  " \t\n\r ",
			},
			expected:      []Token{},
			expectedError: false,
		},
		{
			name: "Integer Literal",
			source: SourceFile{
				filename: "test.ilang",
				content:  "123",
			},
			expected: []Token{
				{Kind: Literal, Value: "123"},
			},
			expectedError: false,
		},
		{
			name: "String Literal",
			source: SourceFile{
				filename: "test.ilang",
				content:  `"hello world"`,
			},
			expected: []Token{
				{Kind: Literal, Value: `"hello world"`},
			},
			expectedError: false,
		},
		{
			name: "Unterminated String",
			source: SourceFile{
				filename: "test.ilang",
				content:  `"hello`,
			},
			expected:      nil,
			expectedError: true,
		},
		{
			name: "Keywords",
			source: SourceFile{
				filename: "test.ilang",
				content:  "let if else return extrn",
			},
			expected: []Token{
				{Kind: Keyword, Value: "let"},
				{Kind: Keyword, Value: "if"},
				{Kind: Keyword, Value: "else"},
				{Kind: Keyword, Value: "return"},
				{Kind: Keyword, Value: "extrn"},
			},
			expectedError: false,
		},
		{
			name: "Identifiers",
			source: SourceFile{
				filename: "test.ilang",
				content:  "foo bar_baz",
			},
			expected: []Token{
				{Kind: Identifier, Value: "foo"},
				{Kind: Identifier, Value: "bar_baz"},
			},
			expectedError: false,
		},
		{
			name: "Boolean Literals",
			source: SourceFile{
				filename: "test.ilang",
				content:  "true false",
			},
			expected: []Token{
				{Kind: Literal, Value: "true"},
				{Kind: Literal, Value: "false"},
			},
			expectedError: false,
		},
		{
			name: "Punctuators",
			source: SourceFile{
				filename: "test.ilang",
				content:  "(){};:,",
			},
			expected: []Token{
				{Kind: Punctuator, Value: "("},
				{Kind: Punctuator, Value: ")"},
				{Kind: Punctuator, Value: "{"},
				{Kind: Punctuator, Value: "}"},
				{Kind: Punctuator, Value: ";"},
				{Kind: Punctuator, Value: ":"},
				{Kind: Punctuator, Value: ","},
			},
			expectedError: false,
		},
		{
			name: "Operators",
			source: SourceFile{
				filename: "test.ilang",
				content:  "= + - ! * / == != < > <= >= << >> && ||",
			},
			expected: []Token{
				{Kind: Operator, Value: "="},
				{Kind: Operator, Value: "+"},
				{Kind: Operator, Value: "-"},
				{Kind: Operator, Value: "!"},
				{Kind: Operator, Value: "*"},
				{Kind: Operator, Value: "/"},
				{Kind: Operator, Value: "=="},
				{Kind: Operator, Value: "!="},
				{Kind: Operator, Value: "<"},
				{Kind: Operator, Value: ">"},
				{Kind: Operator, Value: "<="},
				{Kind: Operator, Value: ">="},
				{Kind: Operator, Value: "<<"},
				{Kind: Operator, Value: ">>"},
				{Kind: Operator, Value: "&&"},
				{Kind: Operator, Value: "||"},
			},
			expectedError: false,
		},
		{
			name: "Simple Program",
			source: SourceFile{
				filename: "test.ilang",
				content: `
int main(){
	let x = 5;
	return x;
};
`,
			},
			expected: []Token{
				{Kind: Identifier, Value: "int"},
				{Kind: Identifier, Value: "main"},
				{Kind: Punctuator, Value: "("},
				{Kind: Punctuator, Value: ")"},
				{Kind: Punctuator, Value: "{"},
				{Kind: Keyword, Value: "let"},
				{Kind: Identifier, Value: "x"},
				{Kind: Operator, Value: "="},
				{Kind: Literal, Value: "5"},
				{Kind: Punctuator, Value: ";"},
				{Kind: Keyword, Value: "return"},
				{Kind: Identifier, Value: "x"},
				{Kind: Punctuator, Value: ";"},
				{Kind: Punctuator, Value: "}"},
				{Kind: Punctuator, Value: ";"},
			},
			expectedError: false,
		},
		{
			name: "Unexpected Token",
			source: SourceFile{
				filename: "test.ilang",
				content:  "@",
			},
			expected:      nil,
			expectedError: true,
		},
		{
			name: "Simple Addition",
			source: SourceFile{
				filename: "test.ilang",
				content:  "1 + 2",
			},
			expected: []Token{
				{Kind: Literal, Value: "1"},
				{Kind: Operator, Value: "+"},
				{Kind: Literal, Value: "2"},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source)
			tokens, err := l.Lex()

			if (err != nil) != tt.expectedError {
				t.Errorf("Lex() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			if !compareTokens(tokens, tt.expected) {
				t.Errorf("Lex() produced incorrect tokens for source: %q", tt.source.content)
				t.Logf("Got (%d tokens):", len(tokens))
				for _, token := range tokens {
					t.Logf("  Kind: %s, Value: %q", token.Kind, token.Value)
				}
				t.Logf("Want (%d tokens):", len(tt.expected))
				for _, token := range tt.expected {
					t.Logf("  Kind: %s, Value: %q", token.Kind, token.Value)
				}
				t.Fail()
			}
		})
	}
}

