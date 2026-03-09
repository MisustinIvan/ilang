# Formal grammar definition in EBNF

```ebnf
program              ::= { declaration | external_declaration }

declaration          ::= basic_type identifier "(" [ function_argument { "," function_argument } ] ")" block
external_declaration ::= "extrn" basic_type identifier "(" [ function_argument { "," function_argument } ["," "..."] ] | "..." ")"
function_argument    ::= basic_type identifier

block                ::= "{" { expression ";" } [ expression ] "}"

expression           ::= return
                       | bind
                       | assignment
                       | value

value                ::= primary
                       | binary
                       | unary

return               ::= "return" value
bind                 ::= "let" identifier ":" type "=" value
assignment           ::= identifier | index "=" value

binary               ::= primary binary_operator value
unary                ::= unary_operator primary

index                ::= identifier "[" primary "]"

primary              ::= literal
                       | identifier
                       | call
                       | separated
                       | block
                       | condition
                       | index

call                 ::= identifier "(" [ value { "," value } ] ")"
separated            ::= "(" value ")"
condition            ::= "if" value value
                         [ "else" value ]


literal              ::= "*."
int_literal          ::= "*."
identifier           ::= "*."
type                 ::= basic_type | array_type
array_type           ::= "[" int_literal "]" basic_type
basic_type           ::= "int" | "bool" | "float" | "string" | "unit"
binary_operator      ::= "+" | "-" | "*" | "/" | "==" | "<" | ">" | "<=" | ">=" | "<<" | ">>" | "&&" | "||"
unary_operator       ::= "-" | "!"
```
