package main

import (
	"log/slog"

	"github.com/playsthisgame/binq/client"
	"github.com/playsthisgame/binq/types"
)

func main() {
	b_client, err := client.NewBinqClient(&client.Config{Host: "localhost", Port: 3000})
	if err != nil {
		slog.Error("Error creating file:", "error", err)
	}

	queueName := "test_q"

	// create a consumer and receive messages

	consumerClient, err := client.NewBinqConsumerClient(
		b_client,
		&types.ConsumerRequest{QueueName: queueName, BatchSize: 100},
	)

	for {
		msgs, err := consumerClient.Receive()
		if err != nil {
			slog.Error("error receiving messages", "queueName", queueName)
		}

		// ack messages
		ids := make([]uint, len(msgs.Messages))
		for i, msg := range msgs.Messages {
			ids[i] = msg.ID
			slog.Info("messages received", "message", string(msg.Data), "id", msg.ID)
		}

		if len(ids) > 0 {
			consumerClient.Acknowledge(&types.AckMessages{
				MessageIds: ids,
			})
		}
	}
}
