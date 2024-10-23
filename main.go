package main

import (
	"log/slog"
	"os"

	"github.com/playsthisgame/binq/handler"
	"github.com/playsthisgame/binq/tcp"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// init CommandHandler
	cmdHandler := handler.NewCommandHandler()

	// set up tcp server
	server, err := tcp.NewTCPServer(3000) // create an option to start the app on a different port
	if err != nil {
		slog.Error("Error creating new TCP server:", "error", err)
	}
	defer server.Close()
	go server.Start()
	slog.Info("binq started...")

	for {
		cmd := <-server.FromSockets
		cmdHandler.Handle(&cmd)
	}
}
