# Formal grammar definition in EBNF

```ebnf
program              ::= { declaration | external_declaration | comment }

comment              ::= "#" { "*" } "\n"
declaration          ::= basic_type identifier "(" [ function_argument { "," function_argument } ] ")" block
external_declaration ::= "extrn" basic_type identifier "(" [ function_argument { "," function_argument } ["," "..."] ] | "..." ")"
function_argument    ::= type identifier

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
assignment           ::= identifier | index | deref "=" value
deref                ::= "@" identifier

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

literal              ::= basic_literal
                       | array_literal

basic_literal        ::= int_literal
                       | float_literal
                       | string_literal
                       | bool_literal

array_literal        ::= "[" [ value { "," value } ] "]"

call                 ::= identifier "(" [ value { "," value } ] ")"
separated            ::= "(" value ")"
condition            ::= "if" value value
                         [ "else" value ]

identifier           ::= letter { letter | digit | "_" }
int_literal          ::= digit { digit }
float_literal        ::= digit { digit } "." { digit }
string_literal       ::= "\"" { * } "\""
bool_literal         ::= "true" | "false"

type                 ::= basic_type | array_type | slice_type | pointer_type
array_type           ::= "[" int_literal "]" basic_type
slice_type           ::= "[" [identifier] "]" basic_type
basic_type           ::= "int" | "bool" | "float" | "string" | "unit"
pointer_type         ::= "^" basic_type

binary_operator      ::= "+" | "-" | "*" | "/" | "==" | "!=" | "<" | ">" | "<=" | ">=" | "<<" | ">>" | "&&" | "||"
unary_operator       ::= "-" | "!" | "^" | "@"

letter               ::= "a" | ... | "z" | "A" | ... | "Z" | "_"
digit                ::= "0" | ... | "9"
```
