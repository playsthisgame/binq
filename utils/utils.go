package utils

import "math/rand/v2"

// Function to split a slice into batches
func ChunkSlice(slice []uint, chunkSize int) [][]uint {
	var chunks [][]uint
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

// Return a random integer in a range
func RandRange(min, max int) int {
	return rand.IntN(max+1-min) + min
}
