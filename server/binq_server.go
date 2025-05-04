package server

import (
	"log/slog"
	"os"

	"github.com/google/uuid"

	"github.com/playsthisgame/binq/handler"
	"github.com/playsthisgame/binq/store"
	"github.com/playsthisgame/binq/tcp"
)

type Config struct {
	Port          uint16
	MaxPartitions int
	Passkey       uuid.UUID
}

type BinqServer struct {
	cmdHandler *handler.CommandHandler
	server     *tcp.TCP
	port       uint16
}

func NewBinqServer(conf *Config) (*BinqServer, error) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// init CommandHandler
	var maxPartitions int = 100
	if conf.MaxPartitions != 0 {
		maxPartitions = conf.MaxPartitions
	}
	db, err := store.Setup()
	if err != nil {
		panic(err)
	}
	cmdHandler := handler.NewCommandHandler(db, &maxPartitions)

	// set up tcp server
	var port uint16 = 3000
	if conf.Port != 0 {
		port = conf.Port
	}

	var passkey uuid.UUID = uuid.Nil
	if conf.Passkey != uuid.Nil {
		passkey = conf.Passkey
	}

	server, err := tcp.NewTCPServer(port, passkey)
	if err != nil {
		slog.Error("Error starting Binq on", "port", port, "Error", err)
		return nil, err
	}

	return &BinqServer{
		cmdHandler: cmdHandler,
		server:     server,
		port:       port,
	}, nil
}

func (b *BinqServer) Listen() {
	defer b.server.Close()
	go b.server.Start()
	slog.Info("Binq started on", "port", b.port)

	go store.ScheduleCleanup()

	for {
		cmd := <-b.server.FromSockets
		b.cmdHandler.Handle(&cmd)
	}
}

func (b *BinqServer) Close() {
	b.server.Close()
}
