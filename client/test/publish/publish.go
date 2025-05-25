package main

import (
	"fmt"
	"log/slog"

	"github.com/playsthisgame/binq/client"
	"github.com/playsthisgame/binq/types"
)

func main() {
	client, err := client.NewBinqClient(
		&client.Config{Host: "localhost", Port: 3000, PublicKey: ".cert/localhost.pem"},
	)
	if err != nil {
		slog.Error("Error creating file:", "error", err)
	}

	queueName := "test_q"

	// create a new queue

	// client.Create(types.Queue{
	// 	Name:          queueName,
	// 	MaxPartitions: 100,
	// })

	// send 100 message to a queue

	for i := range make([]struct{}, 10000) {
		client.Publish(types.Message{
			QueueName: queueName,
			Data:      []byte(fmt.Sprintf("this is a message thats sent to binq %v", i)),
		})
	}
}
