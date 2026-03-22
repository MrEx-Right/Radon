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

// execute_gcc acts as a passthrough to the system's GCC compiler
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

// instrument_assembly parses the raw Assembly and injects tracking payloads
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
    
    // Seed the random number generator for unique basic block IDs
    srand((unsigned int)time(NULL)); 

    while (fgets(line, sizeof(line), in)) {
        fputs(line, out); // Write the original instruction
        
        // Target the 'main' function and all GCC-generated branch labels (.L)
        // Exclude debug labels like .LFB (Function Begin) and .LVL (Locals)
        if (strncmp(line, "main:", 5) == 0 || 
           (strncmp(line, ".L", 2) == 0 && strstr(line, "FB") == NULL && strstr(line, "VL") == NULL)) {
            
            // Assign a random identifier (0-65535) to this basic block
            int block_id = rand() % MAP_SIZE;

            // X86_64 Assembly Injection: 
            // Store the block ID in %rcx and call the Radon runtime tracer
            fputs("\t# --- RADON TRAMPOLINE START ---\n", out);
            fprintf(out, "\tmovq $%d, %%rcx\n", block_id);
            fputs("\tcall __radon_trace\n", out);
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

    // Pass through to GCC directly if no C file is specified (e.g., linking phase)
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