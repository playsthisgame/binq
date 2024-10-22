package handler

import (
	"log/slog"
	"os"
	"plays-tcp/types"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type CommandHandler struct {
	db *gorm.DB
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
		db: db,
	}
}

// consider moving the db connection to this class

func (h *CommandHandler) Handle(cmdWrapper *types.TCPCommandWrapper) {

	cmds := make(map[int]string)
	cmds[0] = "WRITE"
	cmds[1] = "READ"

	cmd := cmdWrapper.Command.Command
	// data := cmdWrapper.Command.Data

	op, ok := cmds[int(cmd)]
	if ok {
		slog.Info("new operation", "op", op)
	} else {
		slog.Info("Unknown Operation")
	}

}
