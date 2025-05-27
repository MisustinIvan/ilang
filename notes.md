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
## third pass
- [ ] TODO: do the rest...
