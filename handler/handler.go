package handler

import (
	"log/slog"
	"os"
	"plays-tcp/types"
	"sync"

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
	return &CommandHandler{
		db:              db,
		maxPartitions:   100,                                  // TODO: bring this in from a config, defaulting to 100 for now
		consumerSockets: make([]types.ConsumerSocket, 0, 100), //probably can use maxPartitions here
	}
}

func (h *CommandHandler) Handle(cmdWrapper *types.TCPCommandWrapper) {

	// TODO: figure out how to use iota
	const (
		produce = "PRODUCE"
		consume = "CONSUME"
	)

	cmds := make(map[int]string)
	cmds[1] = "PRODUCE"
	cmds[2] = "CONSUME"

	cmd := cmdWrapper.Command.Command
	// data := cmdWrapper.Command.Data

	op, ok := cmds[int(cmd)]
	if ok {
		switch op {
		case produce:
			// return handleRead(cmdWrapper)
			slog.Info("producer added", "op", op)
		case consume:
			// make a new consumer socket
			consumerSocket, err := types.NewConsumerSocket(cmdWrapper.Conn.Id, cmdWrapper.Conn.Id, h.maxPartitions, *cmdWrapper.Conn)
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
		}
	}
	// return fmt.Errorf("Unknown Operation %d", int(cmd))
}

func rebalanceConsumers(consumerSockets *[]types.ConsumerSocket, totalInstance int, maxPartitions int) {
	// TODO: there seems to be some kind of bug when getting to around 20 instances
	for i := 0; i < len(*consumerSockets); i++ {
		(*consumerSockets)[i].Partitions = types.SetPartitions((*consumerSockets)[i].Instance, totalInstance, maxPartitions)
		slog.Info("consumer rebalanced", "instance", (*consumerSockets)[i].Instance, "partition count", len((*consumerSockets)[i].Partitions), "partitions", (*consumerSockets)[i].Partitions)
	}
}
