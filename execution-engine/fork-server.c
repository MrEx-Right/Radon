/**
 * @file fork-server.c
 * @brief High-performance Fork Server for the Coverage-Guided Fuzzer.
 */

#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/wait.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <signal.h>
#include <string.h>
#include <errno.h>
#include <stdint.h>
#include <sys/ipc.h>
#include <sys/shm.h>

// Import the shared memory definitions
#include "../ipc/shm.h"

#define FORKSRV_CTRL_FD   3  // Control command (Go -> C)
#define FORKSRV_STATUS_FD 4  // Status/Handshake (C -> Go)

// ============================================================================
// THE EYES: 64KB Coverage Map
// ============================================================================
uint8_t* trace_bits;

/**
 * @brief Attaches to the 64KB shared memory segment allocated by the Orchestrator.
 */
void setup_shared_memory() {
    // Retrieve the SHM ID key from the environment variable
    char* shm_env = getenv(SHM_ENV_VAR);
    if (!shm_env) {
        fprintf(stderr, "[-] FATAL: %s environment variable not set. Orchestrator sleeping?\n", SHM_ENV_VAR);
        exit(EXIT_FAILURE);
    }

    int shm_id = atoi(shm_env);
    
    // Attach to the memory segment
    trace_bits = (uint8_t*)shmat(shm_id, NULL, 0);
    
    if (trace_bits == (void*)-1) {
        perror("[-] FATAL: shmat() failed to attach coverage bitmap");
        exit(EXIT_FAILURE);
    }
    
    printf("[+] Shared Memory attached successfully.\n");
}

void setup_child_io_redirection() {
    
    int input_fd = open("/dev/shm/radon_cur_input", O_RDONLY);
    if (input_fd < 0) {
        perror("[-] FATAL: Failed to open payload from shm");
        exit(EXIT_FAILURE);
    }

    
    int dev_null_fd = open("/dev/null", O_RDWR);
    if (dev_null_fd < 0) {
        perror("[-] FATAL: Failed to open /dev/null");
        exit(EXIT_FAILURE);
    }
    dup2(dev_null_fd, STDOUT_FILENO);
    dup2(dev_null_fd, STDERR_FILENO);
    close(dev_null_fd);

    
    dup2(input_fd, STDIN_FILENO);
    close(input_fd);
}

void run_fork_server(char* target_path, char** target_argv) {
    int status;
    pid_t child_pid;
    uint32_t cmd_trigger;

    // Handshake Phase
    uint32_t ready_signal = 0xDEADBEEF;
    if (write(FORKSRV_STATUS_FD, &ready_signal, 4) != 4) {
        fprintf(stderr, "[-] FATAL: Cannot communicate with Orchestrator on FD %d.\n", FORKSRV_STATUS_FD);
        exit(EXIT_FAILURE);
    }
    
    printf("[+] Fork Server handshake complete. Entering high-speed fuzzing loop...\n");

    // Execution Loop
    while (1) {
        if (read(FORKSRV_CTRL_FD, &cmd_trigger, 4) != 4) {
            printf("[!] Orchestrator disconnected. Shutting down Fork Server gracefully.\n");
            break;
        }

        child_pid = fork();
        
        if (child_pid < 0) {
            perror("[-] FATAL: fork() failed");
            exit(EXIT_FAILURE);
        }

        // CHILD PROCESS
        if (child_pid == 0) {
            setup_child_io_redirection();

            // TODO: Inject PTRACE agent here for deep binary coverage tracking later
            
            execv(target_path, target_argv);
            
            // execv sadece hata verirse buraya döner
            perror("[-] FATAL: execv() failed in child process");
            exit(EXIT_FAILURE); 
        }

        // PARENT PROCESS
        if (write(FORKSRV_STATUS_FD, &child_pid, 4) != 4) {
            perror("[-] FATAL: Failed to send child PID to Orchestrator");
            exit(EXIT_FAILURE);
        }

        if (waitpid(child_pid, &status, 0) <= 0) {
            perror("[-] FATAL: waitpid() failed");
            exit(EXIT_FAILURE);
        }

        if (write(FORKSRV_STATUS_FD, &status, 4) != 4) {
            perror("[-] FATAL: Failed to send execution status to Orchestrator");
            exit(EXIT_FAILURE);
        }
    }
}

int main(int argc, char **argv) {
    if (argc < 2) {
        fprintf(stderr, "Usage: %s <target_binary_path> [args...]\n", argv[0]);
        return EXIT_FAILURE;
    }

    printf("[================================================]\n");
    printf("[*] Execution Engine (Fork Server) Initializing...\n");
    printf("[*] Target Binary: %s\n", argv[1]);
    printf("[================================================]\n");

    // Initialize Coverage Mapping
    setup_shared_memory();

    run_fork_server(argv[1], &argv[1]);

    return EXIT_SUCCESS;
}