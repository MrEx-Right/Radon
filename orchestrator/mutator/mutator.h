#ifndef MUTATOR_H
#define MUTATOR_H

#include <stdint.h>
#include <stddef.h>

extern void fast_bit_flip(uint8_t* buffer, size_t length, size_t byte_offset, uint8_t bit_offset);

#endif