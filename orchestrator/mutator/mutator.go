package mutator

import (
	"encoding/binary"
	"math/rand"
	"time"
)

// Initialize the global pseudo-random number generator.
func init() {
	rand.Seed(time.Now().UnixNano())
}

// Mutate applies a series of Havoc-style mutations to the provided byte slice.
// This function utilizes stacking to apply multiple random strategies in a single pass.
func Mutate(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	mutated := make([]byte, len(data))
	copy(mutated, data)

	// Perform havoc stacking: apply between 1 and 4 mutations consecutively.
	havocCount := rand.Intn(4) + 1

	for i := 0; i < havocCount; i++ {
		// Select one of the 5 available mutation strategies.
		strategy := rand.Intn(5) 

		switch strategy {
		case 0:
			// Strategy 0: Byte Overwrite. 
			// Replace a random byte with a completely random value.
			targetIdx := rand.Intn(len(mutated))
			mutated[targetIdx] = byte(rand.Intn(256))

		case 1:
			// Strategy 1: Bit Flip. 
			// Invert a single random bit within a random byte.
			targetIdx := rand.Intn(len(mutated))
			bitIdx := rand.Intn(8)
			mutated[targetIdx] ^= (1 << bitIdx)

		case 2:
			// Strategy 2: Magic Number Injection. 
			// Insert known boundary values to trigger edge-case vulnerabilities (e.g., overflows/underflows).
			if len(mutated) >= 4 {
				magicNumbers := []uint32{
					0xFFFFFFFF, // -1
					0x00000000, // 0
					0x7FFFFFFF, // Max Int32
					0x80000000, // Min Int32
					0x0000FFFF, // Max Int16
				}
				targetIdx := rand.Intn(len(mutated) - 3)
				magic := magicNumbers[rand.Intn(len(magicNumbers))]
				
				// Inject the selected magic number in Little Endian format.
				binary.LittleEndian.PutUint32(mutated[targetIdx:], magic)
			}

		case 3:
			// Strategy 3: Block Overwrite. 
			// Overwrite a random-sized block with random junk data.
			if len(mutated) > 2 {
				// Ensure the overwrite size does not exceed half of the total payload.
				blockSize := rand.Intn(len(mutated)/2) + 1 
				targetIdx := rand.Intn(len(mutated) - blockSize)
				for j := 0; j < blockSize; j++ {
					mutated[targetIdx+j] = byte(rand.Intn(256))
				}
			}

		case 4:
			// Strategy 4: Block Copy/Swap. 
			// Clone a random block of data and overwrite another section within the payload.
			if len(mutated) > 4 {
				blockSize := rand.Intn(len(mutated)/4) + 1
				srcIdx := rand.Intn(len(mutated) - blockSize)
				dstIdx := rand.Intn(len(mutated) - blockSize)
				copy(mutated[dstIdx:dstIdx+blockSize], mutated[srcIdx:srcIdx+blockSize])
			}
		}
	}

	return mutated
}