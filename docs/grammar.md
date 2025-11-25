# Formal grammar definition in EBNF

```ebnf
program              ::= { fun_decl | extrn_decl }

fun_decl             ::= type ident "(" param_types ")" block_expr
param_types          ::= [ type ident { "," type ident } ]
extrn_decl           ::= "extrn" type ident "(" param_types ")"

block_expr           ::= "{" { expr ";" } [ expr ] "}"

expr                 ::= simple_expr
                       | bind_expr
                       | return_expr

simple_expr          ::= assg_expr
                       | bin_expr
                       | unary_expr

bind_expr            ::= "let" ident ":" type "=" simple_expr
return_expr          ::= "return" simple_expr
assg_expr            ::= ident "=" simple_expr
bin_expr             ::= pexpr binop simple_expr
unary_expr           ::= unop primary
con_expr             ::= "if" simple_expr expr
                         { "else" "if" simple_expr expr }
                         [ "else" expr ]


pexpr                ::= literal
                       | ident
                       | call_expr
                       | block_expr
                       | sep_expr
                       | con_expr


call_expr            ::= ident "(" [ simple_expr { "," simple_expr } ] ")"
sep_expr             ::= "(" simple_expr ")"

ident                ::= "*."
type                 ::= "*."
literal              ::= "*."
binop                ::= "+" | "-" | "*" | "/" | "==" | "<" | ">" | "<=" | ">=" | "<<" | ">>" | "&&" | "||"
unop                 ::= "-" | "!"
```
