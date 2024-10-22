package main

import (
	"log/slog"
	"os"
	"plays-tcp/tcp"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	tcp, err := tcp.NewTCPServer(3000)
	if err != nil {
		slog.Error("Error creating new TCP server:", "error", err)
	}
	tcp.Start()
}
