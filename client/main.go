package main

import (
	"crypto/rand"
	"io"
	"log/slog"
	"net"

	"github.com/playsthisgame/binq/types"
)

// an example of a client
func main() {

	file := make([]byte, 100)
	_, err := io.ReadFull(rand.Reader, file)
	if err != nil {
		slog.Error("Error creating file:", "error", err)
	}

	cmd := &types.TCPCommand{
		Command: 1,
		Data:    file,
	}

	data, err := cmd.MarshalBinary()
	if err != nil {
		slog.Error("Error marshalling data:", "error", err)
	}

	// dial server
	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		slog.Error("Error dialing server:", "error", err)
	}

	defer conn.Close()

	_, err = conn.Write(data)
	if err != nil {
		slog.Error("Error dialing server:", "error", err)
	}

	slog.Info("written %d bytes over the network", "bytes", len(data))
}
