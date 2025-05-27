.text
.globl main
main:
#   prologue
    push %rbp
    mov %rsp, %rbp
#   unknown expression
    mov %rax, %rdi
#   unknown expression
    mov %rax, %rsi
    call magic
#   epilogue
    pop %rbp
    ret
magic:
#   prologue
    push %rbp
    mov %rsp, %rbp
#   unknown expression
#   epilogue
    pop %rbp
    ret
