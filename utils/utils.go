package utils

import (
	"crypto/md5"
	"fmt"
	"math/rand/v2"
	"net"
)

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

func GetConnectionHash(conn net.Conn) string {
	localAddr := conn.LocalAddr().String()
	remoteAddr := conn.RemoteAddr().String()
	pointerAddr := fmt.Sprintf("%p", conn)

	data := localAddr + "|" + remoteAddr + "|" + pointerAddr
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}
