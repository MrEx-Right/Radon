package mutator

import (
	"math/rand"
	"time"
)

// Initialize the random seed for the mutation engine.
func init() {
	rand.Seed(time.Now().UnixNano())
}

// Mutate applies random bitwise or bytewise corruptions to the input data.
func Mutate(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	// Create a deep copy to preserve the original seed in the corpus
	mutated := make([]byte, len(data))
	copy(mutated, data)

	// Randomly select a mutation strategy (Bit Flip or Byte Flip)
	strategy := rand.Intn(2)

	switch strategy {
	case 0:
		// Strategy 0: Byte Flip (Overwrite a random byte with a random value)
		targetIdx := rand.Intn(len(mutated))
		mutated[targetIdx] = byte(rand.Intn(256))
	case 1:
		// Strategy 1: Bit Flip (Invert a random bit within a random byte)
		targetIdx := rand.Intn(len(mutated))
		bitIdx := rand.Intn(8)
		mutated[targetIdx] ^= (1 << bitIdx)
	}

	return mutated
}