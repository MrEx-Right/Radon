#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/wait.h>
#include <time.h>

// ============================================================================
// RADON COMPILER WRAPPER (radon-cc)
// Intercepts GCC compilation, generates assembly, injects coverage 
// trampolines (AFL-style XOR mapping), and builds the final binary.
// ============================================================================

#define MAP_SIZE 65536 // 64KB coverage map, standard for modern fuzzers

/**
 * @brief Executes the system GCC compiler as a child process and waits for its completion.
 * * @param args Null-terminated array of arguments to pass to GCC.
 */
void execute_gcc(char **args) {
    pid_t pid = fork();
    if (pid == 0) {
        execvp("gcc", args);
        perror("[-] FATAL: execvp failed to execute gcc");
        exit(1);
    } else if (pid > 0) {
        wait(NULL);
    } else {
        perror("[-] FATAL: fork failed during compiler execution");
        exit(1);
    }
}

/**
 * @brief Parses the raw assembly output and injects coverage-tracking payloads 
 * at the start of each basic block.
 * * @param asm_file Path to the target assembly (.s) file.
 */
void instrument_assembly(const char *asm_file) {
    char temp_file[256];
    snprintf(temp_file, sizeof(temp_file), "%s.radon", asm_file);

    FILE *in = fopen(asm_file, "r");
    FILE *out = fopen(temp_file, "w");
    if (!in || !out) {
        perror("[-] FATAL: Failed to open assembly files for instrumentation");
        exit(1);
    }

    char line[1024];
    int injected_count = 0;
    
    // Seed the PRNG with a combination of time and PID to ensure 
    // unique block IDs across parallel compilation processes.
    srand((unsigned int)(time(NULL) ^ getpid())); 

    while (fgets(line, sizeof(line), in)) {
        fputs(line, out); // Write the original instruction
        
        // Target the 'main' function and all GCC-generated branch labels (.L).
        // Exclude debug and internal labels such as .LFB (Function Begin) and .LVL (Locals).
        if (strncmp(line, "main:", 5) == 0 || 
           (strncmp(line, ".L", 2) == 0 && strstr(line, "FB") == NULL && strstr(line, "VL") == NULL)) {
            
            // Assign a pseudo-random identifier (0-65535) to this basic block
            int block_id = rand() % MAP_SIZE;

            // ========================================================================
            // X86_64 Assembly Injection (The Trampoline)
            // ========================================================================
            fputs("\t# --- RADON TRAMPOLINE START ---\n", out);
            
            // 1. RED ZONE PROTECTION: Shift the stack pointer by 128 bytes to prevent 
            //    the tracer from corrupting the target's local variables defined in 
            //    the System V AMD64 ABI red zone.
            fputs("\tleaq -128(%rsp), %rsp\n", out);
            
            // 2. Backup the RCX register (Using single '%' for fputs).
            fputs("\tpushq %rcx\n", out);
            
            // 3. Load the generated Block ID into RCX and invoke the Radon runtime tracer.
            //    (Using double '%%' here because fprintf parses format specifiers).
            fprintf(out, "\tmovq $%d, %%rcx\n", block_id);
            fputs("\tcall __radon_trace\n", out);
            
            // 4. Restore the RCX register and the stack pointer to their original states.
            fputs("\tpopq %rcx\n", out);
            fputs("\tleaq 128(%rsp), %rsp\n", out);
            
            fputs("\t# --- RADON TRAMPOLINE END ---\n", out);
            
            injected_count++;
        }
    }

    fclose(in);
    fclose(out);
    
    // Replace the original assembly with the instrumented version
    rename(temp_file, asm_file);
    printf("[*] radon-cc: Successfully injected %d coverage trampolines.\n", injected_count);
}

int main(int argc, char **argv) {
    if (argc < 2) {
        printf("[-] FATAL: No arguments provided to radon-cc\n");
        return 1;
    }

    char *input_file = NULL;
    char *output_file = "a.out"; // Default GCC output name

    // Parse arguments to find the input source and target output file
    for (int i = 1; i < argc; i++) {
        if (strstr(argv[i], ".c") != NULL) {
            input_file = argv[i];
        } else if (strcmp(argv[i], "-o") == 0 && i + 1 < argc) {
            output_file = argv[i+1];
        }
    }

    // Pass through to GCC directly if no C file is specified (e.g., during linking phase)
    if (!input_file) {
        char **gcc_args = malloc((argc + 1) * sizeof(char*));
        gcc_args[0] = "gcc";
        for (int i = 1; i < argc; i++) gcc_args[i] = argv[i];
        gcc_args[argc] = NULL;
        execute_gcc(gcc_args);
        free(gcc_args);
        return 0;
    }

    printf("[*] radon-cc: Intercepting compilation for '%s'\n", input_file);

    // Phase 1: Compile the C source code down to raw Assembly (.s)
    char asm_file[256];
    snprintf(asm_file, sizeof(asm_file), "%s.s", input_file);
    
    char *asm_args[] = {"gcc", "-S", input_file, "-o", asm_file, NULL};
    execute_gcc(asm_args);
    

    // Phase 2: Inject the XOR coverage map logic into the Assembly
    instrument_assembly(asm_file);

    // Phase 3: Assemble the poisoned code into the final executable
    // NOTE: In the future, we must link the radon-rt.o runtime here!
    char *final_args[] = {"gcc", asm_file, "compiler/radon-rt.o", "compiler/radon-trace.o", "-o", output_file, NULL};
    execute_gcc(final_args);

    // Cleanup temporary files
    remove(asm_file);
    
    printf("[+] radon-cc: Compilation finished for '%s'\n", output_file);
    return 0;
}