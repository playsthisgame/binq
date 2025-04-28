package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/playsthisgame/binq/client"
	"github.com/playsthisgame/binq/types"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	binq_client, err := client.NewBinqClient(&client.Config{Host: "localhost", Port: 3000})
	if err != nil {
		logger.Error("Error creating file:", "error", err)
	}

	// queueName := "coldbrew_q"
	queueName := "test_q"

	path := "/Users/chris/workspace/dev/go/src/github.com/playsthisgame/binq/client/test/publish/image/test.jpg"
	file, err := os.Open(path)
	if err != nil {
		slog.Info("unable to open file", "path", path)
		return
	}
	defer file.Close()
	fileName := filepath.Base(path)
	fileExtension := filepath.Ext(path)

	data, err := imageToBytes(path)
	if err != nil {
		logger.Info("Error getting []byte from image", "path", path)
		os.Exit(1)
	}

	binq_client.Publish(types.Message{
		QueueName:     queueName,
		FileExtension: fileExtension,
		FileName:      fileName,
		Data:          data,
	})
}

func imageToBytes(filePath string) ([]byte, error) {
	// Read the entire file into memory
	return os.ReadFile(filePath)
}
