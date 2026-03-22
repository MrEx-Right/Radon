#ifndef SHM_H
#define SHM_H

#include <stdint.h>

// AFL-style 64KB coverage bitmap. Standard for high-performance fuzzing.
#define MAP_SIZE (64 * 1024)
#define SHM_ENV_VAR "EXCALIBUR_SHM_ID"

extern uint8_t* trace_bits;

#endif