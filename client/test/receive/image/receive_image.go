package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log/slog"
	"os"

	"github.com/playsthisgame/binq/client"
	"github.com/playsthisgame/binq/types"
)

func main() {
	b_client, err := client.NewBinqClient(
		&client.Config{Host: "localhost", Port: 3000, PublicKey: ".cert/local/cert.pem"},
	)
	if err != nil {
		slog.Error("Error creating file:", "error", err)
	}
	queueName := "test_q"

	consumer, err := client.NewBinqConsumerClient(
		b_client,
		&types.ConsumerRequest{QueueName: queueName, BatchSize: 100},
	)

	for {
		msgs, err := consumer.Receive()
		if err != nil {
			slog.Error("error receiving messages", "queueName", queueName)
		}

		ids := make([]uint, len(msgs.Messages))
		for i, msg := range msgs.Messages {
			ids[i] = msg.ID

			img, format, err := byteSliceToImage(msg.Data)
			if err != nil {
				slog.Error("Failed to decode image", "Error", err)
			}

			slog.Info("Successfully decoded image, dimensions:", "format",
				format, "x", img.Bounds().Dx(), "y", img.Bounds().Dy())

			// Save the image to a file
			outputPath := "output." + format
			err = saveImage(img, format, outputPath)
			if err != nil {
				slog.Error("Failed to save image", "error", err)
			}

			slog.Info("Image saved", "path", outputPath)

		}

		if len(ids) > 0 {
			consumer.Acknowledge(&types.AckMessages{
				MessageIds: ids,
			})
		}
	}
}

func byteSliceToImage(data []byte) (image.Image, string, error) {
	// Create a reader from the byte slice
	reader := bytes.NewReader(data)

	// Decode the image
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, "", err
	}

	return img, format, nil
}

func saveImage(img image.Image, format, filePath string) error {
	// Create a new file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Encode and save the image based on format
	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
	case "png":
		err = png.Encode(file, img)
	default:
		return fmt.Errorf("unsupported image format: %s", format)
	}

	return err
}
