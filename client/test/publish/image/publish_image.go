package main

import (
	"log/slog"
	"os"

	"github.com/playsthisgame/binq/client"
	"github.com/playsthisgame/binq/types"
)

// TODO: theres some exception when publishing large []byte like this example
func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	binq_client, err := client.NewBinqClient(&client.Config{Host: "localhost", Port: 3000})
	if err != nil {
		logger.Error("Error creating file:", "error", err)
	}

	queueName := "test_q"

	path := "/Users/chris/workspace/dev/go/src/github.com/playsthisgame/binq/client/test/publish/image/test.jpg"
	data, err := imageToBytes(path)
	if err != nil {
		logger.Info("Error getting []byte from image", "path", path)
		os.Exit(1)
	}

	binq_client.Publish(types.Message{
		QueueName: queueName,
		Data:      data,
	})
}

func imageToBytes(filePath string) ([]byte, error) {
	// Read the entire file into memory
	return os.ReadFile(filePath)
}
