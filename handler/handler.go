package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"sync"

	"github.com/playsthisgame/binq/types"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type CommandHandler struct {
	db              *gorm.DB
	consumerSockets []types.ConsumerSocket
	maxPartitions   int
	mutex           sync.RWMutex
}

func NewCommandHandler() *CommandHandler {
	// create .store if not exists
	path := ".store"
	_ = os.MkdirAll(path, os.ModePerm)

	// init db
	db, err := gorm.Open(sqlite.Open(".store/binq.db"), &gorm.Config{})
	if err != nil {
		slog.Error("Error initializing sqlite", "error", err)
	}
	// automigrate db
	db.AutoMigrate(&types.Message{})
	db.AutoMigrate(&types.Queue{})
	return &CommandHandler{
		db:              db,
		maxPartitions:   100,                                  // TODO: bring this in from a config, defaulting to 100 for now
		consumerSockets: make([]types.ConsumerSocket, 0, 100), //probably can use maxPartitions here
	}
}

func (h *CommandHandler) Handle(cmdWrapper *types.TCPCommandWrapper) {

	// TODO: figure out how to use iota
	const (
		create  = "CREATE"
		produce = "PRODUCE"
		consume = "CONSUME"
	)

	cmds := make(map[int]string)
	cmds[1] = "CREATE"
	cmds[2] = "PRODUCE"
	cmds[3] = "CONSUME"

	cmd := cmdWrapper.Command.Command
	// data := cmdWrapper.Command.Data

	op, ok := cmds[int(cmd)]
	if ok {
		switch op {
		case create:
			err := createQueue(cmdWrapper.Command.Data, *h.db)
			if err != nil {
				slog.Error("Error while create queue")
			}
		case produce:
			err := createMessage(cmdWrapper.Command.Data, *h.db)
			if err != nil {
				slog.Error("Error while create queue")
			}
		case consume:
			// make a new consumer socket
      queueName := string(cmdWrapper.Command.Data)
			consumerSocket, err := types.NewConsumerSocket(cmdWrapper.Conn.Id, cmdWrapper.Conn.Id, h.maxPartitions, queueName, *cmdWrapper.Conn)
			if err != nil {
				slog.Error("Error create client socket", "id", cmdWrapper.Conn.Id)
			}
			// if a consumerSocket already exists then rebalance
			if len(h.consumerSockets) > 0 {
				rebalanceConsumers(&h.consumerSockets, cmdWrapper.Conn.Id, h.maxPartitions)
			}

			h.mutex.Lock()
			h.consumerSockets = append(h.consumerSockets, *consumerSocket)
			h.mutex.Unlock()

			slog.Info("consumer added", "instance", consumerSocket.Instance, "partition count", len(consumerSocket.Partitions), "partitions", consumerSocket.Partitions)

			// does this go into a go routine?
			for i := 0; i < len(h.consumerSockets); i++ {
				// consumer := h.consumerSockets[i]

				// create the query
				// write to the consumer
			}
		}
	}
}

func rebalanceConsumers(consumerSockets *[]types.ConsumerSocket, totalInstance int, maxPartitions int) {
	// TODO: there seems to be some kind of bug when getting to around 20 instances
	for i := 0; i < len(*consumerSockets); i++ {
		(*consumerSockets)[i].Partitions = types.SetPartitions((*consumerSockets)[i].Instance, totalInstance, maxPartitions)
		slog.Info("consumer rebalanced", "instance", (*consumerSockets)[i].Instance, "partition count", len((*consumerSockets)[i].Partitions), "partitions", (*consumerSockets)[i].Partitions)
	}
}

func createQueue(data []byte, db gorm.DB) error {
	var queue types.Queue
	err := json.Unmarshal(data, &queue)
	if err != nil {
		return errors.New("error unmarshalling queue")
	}

	res := db.Create(&queue)
	if res.Error != nil {
		return errors.New(fmt.Sprintf("Error creating queue %s", queue.Name))
	}
	slog.Info("Queue Created", "name", queue.Name)
	return nil
}

// assign the partition to the message here
func createMessage(data []byte, db gorm.DB) error {
	var msg types.Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return errors.New("error unmarshalling message")
	}
	// assign the partition
	// TODO: just using 100 here but we should get the maxPartitions from the queue, without having to query the queue everytime.
	msg.Partition = randRange(1, 100)

	res := db.Create(&msg)
	if res.Error != nil {
		return errors.New(fmt.Sprintf("Error creating message for %s", msg.QueueName))
	}
	slog.Info("Message Created for Queue", "queue", msg.QueueName)
	return nil
}

func randRange(min, max int) int {
	return rand.IntN(max+1-min) + min
}

func sendMessages(h *CommandHandler){
  var wg sync.WaitGroup
	for i := 0; i < len(h.consumerSockets); i++ {
    wg.Add(1)
		consumer := h.consumerSockets[i]

    var messages []types.Message
    
    h.db.Where("queue_name = ? AND partitions = ?", h., consumer.Partitions).Find(&messages)

		// create the query
		// write to the consumer
	}
  wg.Wait()
}
