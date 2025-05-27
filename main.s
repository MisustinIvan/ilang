.globl main
.string1:
	.string	"balls"
	.text
main:
.LFB0:
# prologue
	push %rbp
	mov %rsp, %rbp
# stack allocation
	sub $16, %rsp
# load address of string1 to rax(position independent)
	lea .string1(%rip), %rax
# move rax to rdi (1st argument register)
# call printf
	mov %rax, %rdi
	call printf@PLT

# epilogue
	leave
	ret
