// ============================================================================
// EXCALIBUR HIGH-SPEED MUTATION ENGINE (ASM CORE)
// ============================================================================
.intel_syntax noprefix
.global fast_bit_flip
.type fast_bit_flip, @function


fast_bit_flip:
    cmp rdx, rsi
    jge .L_done      

    add rdi, rdx

    mov r8, 1
    shl r8, cl       .


    xor byte ptr [rdi], r8b

.L_done:
    ret