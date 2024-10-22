package handler

import (
	"database/sql"
	"log/slog"
	"plays-tcp/types"
)

type CommandHandler struct {
    db *sql.DB
}

func NewCommandHandler() *CommandHandler {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		slog.Error("Error initializing sqlite","error",err)
	}
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
