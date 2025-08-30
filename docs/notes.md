# multiple pass compiler
## first pass
- [x] register function declarations and process their metadata
## second pass
- [x] push function scope
- [x] find local variables in scopes and register them
- [x] before popping the scope, go back and modify the ast for identifiers to refer to scope local variables
## third pass
- [x] calculate stack offset
- [x] store string literals to be referenced later by a unique identifier
## fourth pass
- [x] generate prologue/epilogue
- [ ] move parameters to onto stack
- [ ] generate expressions
- [ ] generate .text and .data sections

# type checking
- when to check types?
- in binding
- in assignments
- in function calls
- in binary operations
- in unary operations

# TODO:
- figure out when to calculate all the types(needs local variable context because of identifiers)
- figure out when to run the type check step (probably after the above mentioned step)
