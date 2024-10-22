package main

import (
	"log/slog"
	"os"
	"plays-tcp/handler"
	"plays-tcp/tcp"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// init CommandHandler
	cmdHandler := handler.NewCommandHandler()

	// set up tcp server
	server, err := tcp.NewTCPServer(3000, cmdHandler) // create an option to start the app on a different port
	if err != nil {
		slog.Error("Error creating new TCP server:", "error", err)
	}
	defer server.Close()
	go server.Start()
	slog.Info("bin-queue started...")

	for {
		cmd := <-server.FromSockets
		cmdHandler.Handle(&cmd)
	}
}
