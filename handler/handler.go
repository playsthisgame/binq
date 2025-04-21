package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/playsthisgame/binq/types"
	"github.com/playsthisgame/binq/utils"
)

const batchSize = 10

type CommandHandler struct {
	db              *gorm.DB
	consumerSockets []types.ConsumerSocket
	maxPartitions   int
	mutex           sync.RWMutex
}

func NewCommandHandler(maxPartitions *int) *CommandHandler {
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

	// permanently delete soft deleted records older than 1 day, TODO: find a better way to do this
	db.Unscoped().Where("deleted_at < ?", time.Now().AddDate(0, 0, -1)).Delete(&types.Message{})
	return &CommandHandler{
		db:            db,
		maxPartitions: *maxPartitions,
		consumerSockets: make(
			[]types.ConsumerSocket,
			0,
			*maxPartitions,
		), // probably can use maxPartitions here
	}
}

func (h *CommandHandler) Handle(cmdWrapper *types.TCPCommandWrapper) error {
	// TODO: figure out how to use iota,also move this to the struct?
	const (
		create  = "CREATE"
		publish = "PUBLISH"
		receive = "RECEIVE"
		ack     = "ACK"
		oust    = "OUST"
	)

	cmds := make(map[int]string)
	cmds[1] = "CREATE"
	cmds[2] = "PUBLISH"
	cmds[3] = "RECEIVE"
	cmds[4] = "ACK"
	cmds[5] = "OUST"

	cmd := cmdWrapper.Command.Command

	op, ok := cmds[int(cmd)]
	if ok {
		switch op {
		case create:
			err := createQueue(cmdWrapper.Command.Data, *h.db)
			if err != nil {
				slog.Error("Error while create queue")
				return err
			}
		case publish:
			err := createMessage(cmdWrapper.Command.Data, h.maxPartitions, *h.db)
			if err != nil {
				slog.Error("Error while publishing")
				return err
			}
		case receive:
			var request types.ConsumerRequest
			err := request.UnmarshalBinary(cmdWrapper.Command.Data)
			if err != nil {
				slog.Error("error unmarshalling message")
				cmdWrapper.Conn.Close()
				return err
			}

			consumerCount := len(h.consumerSockets) + 1

			// make a new consumer socket
			consumerSocket, err := types.NewConsumerSocket(
				consumerCount,
				consumerCount,
				h.maxPartitions,
				request.QueueName,
				*cmdWrapper.Conn,
			)
			if err != nil {
				slog.Error("Error create client socket", "id", cmdWrapper.Conn.Id, "error", err)
				cmdWrapper.Conn.Close()
				return err
			}

			h.mutex.Lock()
			h.consumerSockets = append(h.consumerSockets, *consumerSocket)
			h.mutex.Unlock()

			rebalanceConsumers(h)

			slog.Info(
				"consumer added",
				"instance",
				consumerSocket.Instance,
				"partition count",
				len(consumerSocket.Partitions),
				"partitions",
				consumerSocket.Partitions,
			)

			go sendMessages(consumerSocket, &request, h.db)

		case ack:
			err := ackMessages(cmdWrapper.Command.Data, *h.db)
			if err != nil {
				slog.Error("Error while create queue")
			}
		case oust:
			// oust consumer socket
			for i := len(h.consumerSockets) - 1; i >= 0; i-- {
				if h.consumerSockets[i].Conn.Id == cmdWrapper.Conn.Id {
					h.mutex.Lock()
					h.consumerSockets = slices.Delete(h.consumerSockets, i, i+1)
					h.mutex.Unlock()

					slog.Info("ousting consumer", "id", cmdWrapper.Conn.Id)
					break
				}
			}
			// reset the consumer socket instances
			for i := range h.consumerSockets {
				h.consumerSockets[i].Instance = i + 1
			}
			// rebalance after ousting
			if len(h.consumerSockets) > 0 {
				rebalanceConsumers(h)
			}
		}
	}
	return nil
}

func rebalanceConsumers(h *CommandHandler) {
	for i := range h.consumerSockets {
		// h.mutex.Lock()
		(h.consumerSockets)[i].Partitions = types.SetPartitions(
			(h.consumerSockets)[i].Instance,
			len(h.consumerSockets),
			h.maxPartitions,
		)
		slog.Info(
			"consumer rebalanced",
			"instance",
			(h.consumerSockets)[i].Instance,
			"partition count",
			len((h.consumerSockets)[i].Partitions),
			"partitions",
			(h.consumerSockets)[i].Partitions,
		)
		// h.mutex.Unlock()
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
func createMessage(data []byte, maxPartitions int, db gorm.DB) error {
	var msg types.Message
	err := msg.UnmarshalBinary(data)
	if err != nil {
		slog.Error("error unmarshalling binary", "error", err)
		return errors.New("error unmarshalling message")
	}
	// assign the partition
	msg.Partition = utils.RandRange(1, maxPartitions)

	res := db.Create(&msg)
	if res.Error != nil {
		return errors.New(fmt.Sprintf("Error creating message for %s", msg.QueueName))
	}
	slog.Info("Message Created for Queue", "queue", msg.QueueName, "size", len(msg.Data))
	return nil
}

func sendMessages(consumer *types.ConsumerSocket, req *types.ConsumerRequest, db *gorm.DB) error {
	for {
		var msgs []types.Message
		db.Limit(req.BatchSize).
			Where("queue_name = ? AND partition IN ? AND (lock_date_time IS NULL OR lock_date_time <= ?)", consumer.QueueName, consumer.Partitions, time.Now()).
			Find(&msgs)

		msgBatch := &types.MessageBatch{
			Messages: msgs,
		}

		// lock messages
		lockMessages(&msgs, db)

		// marshal data
		data, err := msgBatch.MarshalBinary()
		if err != nil {
			return err
		}

		// write to client
		err = consumer.Conn.Writer.Write(&types.TCPCommand{
			Data: data,
		})
		if err != nil {
			// TODO: if theres an error sending to the consumer, then remove it from consumer sockets
			return err
		}
	}
}

func lockMessages(msgs *[]types.Message, db *gorm.DB) {
	ids := make([]uint, len(*msgs))
	for i, msg := range *msgs {
		ids[i] = msg.ID
	}

	if len(ids) > 0 {
		// Split ids into smaller batches
		batches := utils.ChunkSlice(ids, batchSize)

		// Process each batch separately
		for _, batch := range batches {
			db.Table("messages").
				Where("id IN ?", batch).
				Updates(types.Message{LockDateTime: time.Now().Add(time.Minute * 10)})
		}
	}
}

func ackMessages(data []byte, db gorm.DB) error {
	var ackMessage types.AckMessages
	err := ackMessage.UnmarshalBinary(data)
	if err != nil {
		return err
	}

	if len(ackMessage.MessageIds) > 0 {
		// split messageIds into chunks
		batches := utils.ChunkSlice(ackMessage.MessageIds, batchSize)

		// process each batch separately
		for _, batch := range batches {
			var messages []types.Message
			db.Delete(&messages, batch)
		}
	}

	return nil
}
