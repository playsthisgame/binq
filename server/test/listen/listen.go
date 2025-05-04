package main

import (
	"log/slog"

	"github.com/google/uuid"

	"github.com/playsthisgame/binq/server"
)

func main() {
	// passkey, _ := uuid.Parse("8edad377-0f0d-421e-b7bb-22421a873918")
	passkey := uuid.Nil
	binqServer, err := server.NewBinqServer(
		&server.Config{Port: 3000, MaxPartitions: 100, Passkey: passkey},
	)
	if err != nil {
		slog.Error("Error starting Binq Server", "Error", err)
	}

	binqServer.Listen()
}
