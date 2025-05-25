package main

import (
	"log/slog"

	"github.com/playsthisgame/binq/server"
)

func main() {
	binqServer, err := server.NewBinqServer(
		&server.Config{
			Port:          3000,
			MaxPartitions: 100,
			CertPath:      ".cert/",
		},
	)
	if err != nil {
		slog.Error("Error starting Binq Server", "Error", err)
	}

	binqServer.Listen()
}
