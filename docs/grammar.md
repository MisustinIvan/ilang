# Formal grammar definition in EBNF

```ebnf
program              ::= { declaration | external_declaration }

declaration          ::= type identifier "(" function_parameters ")" block
external_declaration ::= "extrn" type identifier "(" function_parameters ")"
function_parameters  ::= [ function_parameter { "," function_parameter } ]
function_parameter   ::= type identifier

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
assignment           ::= identifier "=" value

binary               ::= primary binary_operator value
unary                ::= unary_operator primary

primary              ::= literal
                       | identifier
                       | call
                       | separated
                       | block
                       | condition

call                 ::= identifier "(" [ value { "," value } ] ")"
separated            ::= "(" value ")"
condition            ::= "if" value value
                         [ "else" value ]


literal              ::= "*."
identifier           ::= "*."
type                 ::= "int" | "bool" | "float" | "string" | "unit"
binary_operator      ::= "+" | "-" | "*" | "/" | "==" | "<" | ">" | "<=" | ">=" | "<<" | ">>" | "&&" | "||"
unary_operator       ::= "-" | "!"
```
