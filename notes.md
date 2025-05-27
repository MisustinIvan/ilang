# multiple pass compiler
## first pass
- [x] register function declarations and process their metadata
## second pass
- [x] push function scope
- [x] find local variables in scopes and register them
- [x] before popping the scope, go back and modify the ast for identifiers to refer to scope local variables
## third pass
- [x] calculate stack offset
- [ ] store strings in the data section and calculate their offset
## third pass
- [ ] TODO: do the rest...
