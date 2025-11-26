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
                       | value_expression

value_expression     ::= primary
                       | binary
                       | unary

return               ::= "return" value_expression
bind                 ::= "let" identifier ":" type "=" value_expression
assignment           ::= identifier "=" value_expression

binary               ::= primary binary_operator value_expression
unary                ::= unary_operator primary

primary              ::= literal
                       | identifier
                       | call
                       | separated
                       | block
                       | condition

call                 ::= identifier "(" [ value_expression { "," value_expression } ] ")"
separated            ::= "(" value_expression ")"
condition            ::= "if" value_expression value_expression
                         [ "else" value_expression ]


literal              ::= "*."
identifier           ::= "*."
type                 ::= "int" | "bool" | "float" | "string" | "unit"
binary_operator      ::= "+" | "-" | "*" | "/" | "==" | "<" | ">" | "<=" | ">=" | "<<" | ">>" | "&&" | "||"
unary_operator       ::= "-" | "!"
```
