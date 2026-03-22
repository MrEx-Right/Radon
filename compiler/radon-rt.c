#include <stdint.h>
#include <stdlib.h>
#include <stdio.h>
#include <sys/shm.h>
#include <unistd.h>

// ============================================================================
// RADON RUNTIME INITIALIZER (radon-rt)
// Only responsible for attaching the 64KB coverage map at startup.
// The actual high-speed tracing is handled by radon-trace.S
// ============================================================================

#define MAP_SIZE 65536
#define SHM_ENV_VAR "EXCALIBUR_SHM_ID"

// Global variables accessed by our Assembly tracer
uint8_t *trace_bits = NULL;
uint32_t prev_loc = 0;

__attribute__((constructor)) void __radon_init() {
    char *shm_env = getenv(SHM_ENV_VAR);
    
    if (!shm_env) {
        trace_bits = (uint8_t *)calloc(MAP_SIZE, 1);
        return;
    }
    
    int shm_id = atoi(shm_env);
    trace_bits = (uint8_t *)shmat(shm_id, NULL, 0);
    
    if (trace_bits == (void *)-1) {
        perror("[-] radon-rt: Failed to attach to shared memory");
        exit(1);
    }
}