[x] variadics
[x] arrays
    - [x] refactor types
        - [x] ast structure
        - [x] grammar
        - [x] lexing
        - [x] parsing
        - [x] resolving
        - [x] checking
        - [x] complex types only in block expressions
        - [x] visualising
    - [x] zero initialization
    - [x] allocation
    - [x] indexing
        - [x] ast structure
        - [x] parsing
        - [x] resolving
        - [x] checking
        - [x] visualising
        - [x] assignment
        - [x] indexing - value
    - [x] literals
    - [x] array to array assign semantics - copy
    - [x] slices
        - declared as
        `
            let a: [3]int = 0;
            let b: []int = a;
            let c: [c_len]int = a;
        `
    - [x] array literals for initialization(bind)
        - declared as
        - iterate the values, generate their code and assign them to correct indices in the array
        `
            let array: [4]int = [3, 3+3, add2(2,3), a[3]]
        `

# assignment, binding, literals and function calls
- assignment is - `value = target`
    - scalar assignment is just putting a value into an address
    - array-to-array assignment is `rep stosq`
    - slices are stored in 2 different ways
        - in memory as (pointer, length)
                       ^ base here
        - as a value pointer in `%rax` and length in `%rbx`

- binding is the same, only if binding a slice with a lenght identifier, we have bind the length also

- literals should just be generated simply
    - scalars - constants registered, value in `%rax`
    - arrays - generates itself into its designated stack space, leaves pointer in `%rax` and length in `%rbx`

- function calls
    - scalars are in registers
    - arrays are coerced into slices - passed as pointer and length
    - slices are passed in the same way

- [x] more than 6 function arguments -> push to stack before calling, access in callee through `8(%rbp), 16(%rbp), ...`

- [x] adress-off operator
    - [x] for basic types
- [x] iteration without recursion
- [x] allocating slices on the heap with malloc(size)
    - something like
    ```
        extrn ^int malloc(int size)

        let size: int = 10;
        let allocated_slice: [size]int = malloc(size)
    ```
- [x] floats
- [ ] precedence climbing
- [ ] swap keyword through xor
