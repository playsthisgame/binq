package handler

import (
	"log/slog"
	"plays-tcp/types"
)

func Handle(cmdWrapper *types.TCPCommandWrapper) {

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
