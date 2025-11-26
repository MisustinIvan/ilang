# Formal grammar definition in EBNF

```ebnf
program              ::= { declaration | external_declaration }

declaration          ::= type identififer "(" [ function_parameter { "," function_parameter } ] ")" block
external_declaration ::= "extrn" type identifier "(" [ function_parameter { "," function_parameter } ] ")"
function_parameter   ::= type identifier

block                ::= "{" { expression ";" } [ expression ] "}"

expression           ::= return
                       | bind
                       | primary

return               ::= "return" primary
bind                 ::= "let" identifier ":" type "=" primary
binary               ::= primary binary_operator primary

primary              ::= literal
                       | binary
                       | identifier
                       | call
                       | separated
                       | unary
                       | block
                       | condition
                       | assignment


assignment           ::= identifier "=" primary
unary                ::= unary_operator primary
condition            ::= "if" primary expression
                         { "else" "if" primary expression }
                         [ "else" expression ]
call                 ::= identifier "(" [ primary { "," primary } ] ")"
separated            ::= "(" primary ")"

type                 ::= "int" | "bool" | "float" | "string" | "unit"

identifier           ::= "*."
literal              ::= "*."
binary_operator      ::= "+" | "-" | "*" | "/" | "==" | "<" | ">" | "<=" | ">=" | "<<" | ">>" | "&&" | "||"
unary_operator       ::= "-" | "!"
```
