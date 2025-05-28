# headers
.globl main
main:
#   prologue
    push %rbp
    mov %rsp, %rbp
#   stack allocation
    sub $-16, %rsp
#   move function parameters into local stack space
    mov %rdi, -8(%rbp)
#   function body
#   block expression
#   block expression
#   literal expression
    mov $420, %rax
#   bind expression
    mov %rax, -16(%rbp)
#   assignment expression
    mov %rax, -8(%rbp)
#   epilogue
    leave
    ret
