# headers
.globl main
main:
#   prologue
    push %rbp
    mov %rsp, %rbp
#   stack allocation
    sub $8, %rsp
#   move function parameters into local stack space
#   function body
#   block expression
#   implicit return expression
#   conditional expression
#   block expression
#   implicit return expression
#   literal expression
    mov $0, %rax
#   bind expression
    mov %rax, -8(%rbp)
    cmp $0, %rax
    je .conditional_label_0
#   block expression
#   implicit return expression
#   literal expression
    mov $69, %rax
    jmp .conditional_label_1
.conditional_label_0:
#   block expression
#   implicit return expression
#   literal expression
    mov $420, %rax
.conditional_label_1:
#   epilogue
    leave
    ret
