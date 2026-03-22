# ============================================================================
# RADON CORE TRACER (PURE ASSEMBLY)
# AFL-style XOR coverage map updater. Optimized for extreme execution speed.
# ============================================================================
.intel_syntax noprefix
.global __radon_trace

.text
__radon_trace:
    # 1. SAVE CPU STATE
    # Save volatile registers and CPU flags (EFLAGS) to avoid corrupting 
    # the target's original execution flow and logic conditions.
    push rax
    push rbx
    push rcx
    push rdx
    pushfq

    # 2. LOAD COVERAGE MAP POINTER
    # Fetch the trace_bits global pointer from the Radon Runtime.
    mov rax, qword ptr [trace_bits[rip]]
    test rax, rax
    jz .L_done      # Exit cleanly if shared memory is not attached

    # 3. LOAD PREVIOUS LOCATION
    # Fetch the prev_loc global variable.
    mov ebx, dword ptr [prev_loc[rip]]

    # 4. COMPUTE EDGE ID (AFL XOR MAGIC)
    # rcx holds the current block ID (injected by radon-cc's trampoline)
    mov rdx, rcx
    xor rdx, rbx

    # 5. UPDATE COVERAGE MAP
    # Increment the hit count for this specific branch edge: map[cur ^ prev]++
    inc byte ptr [rax + rdx]

    # 6. UPDATE PREVIOUS LOCATION
    # Shift current location right by 1 to distinguish A->B from B->A paths
    shr rcx, 1
    mov dword ptr [prev_loc[rip]], ecx

.L_done:
    # 7. RESTORE CPU STATE
    # Bring the registers and flags back to their exact original state.
    popfq
    pop rdx
    pop rcx
    pop rbx
    pop rax
    ret