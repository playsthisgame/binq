package tcp

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"

	"sync"
	"syscall"

	"github.com/playsthisgame/binq/types"
)

type TCP struct {
	sockets     []types.Connection
	listener    net.Listener
	mutex       sync.RWMutex
	FromSockets chan types.TCPCommandWrapper
	NewSocket   chan *types.Connection
}

func (t *TCP) ConnectionCount() int {
	return len(t.sockets)
}

// send a command to all of the sockets, remove socket on err
func (t *TCP) Send(command *types.TCPCommand) {
	t.mutex.RLock()
	removals := make([]int, 0)
	slog.Debug("sending message", "msg", command)
	for i, conn := range t.sockets {
		err := conn.Writer.Write(command)
		if err != nil {
			if errors.Is(err, syscall.EPIPE) {
				slog.Debug("connection closed by client", "index", i)
			} else {
				slog.Error("removing due to error", "index", i, "error", err)
			}
			removals = append(removals, i)
		}
	}
	t.mutex.RUnlock()

	if len(removals) > 0 {
		t.mutex.Lock()
		for i := len(removals) - 1; i >= 0; i-- {
			idx := removals[i]
			t.sockets = append(t.sockets[:idx], t.sockets[idx+1:]...)
		}
		t.mutex.Unlock()
	}
}

func (t *TCP) Close() {
	t.listener.Close()
}

func NewTCPServer(port uint16) (*TCP, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	// TODO: Done channel
	return &TCP{
		sockets:     make([]types.Connection, 0, 100),
		listener:    listener,
		FromSockets: make(chan types.TCPCommandWrapper, 100),
		mutex:       sync.RWMutex{},
	}, nil
}

func readConnection(tcp *TCP, conn *types.Connection) {
	for {
		cmd, err := conn.Next() // once you have the command you can do whatever youd like with the data

		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Debug("socket received EOF", "id", conn.Id, "error", err)
			} else {
				slog.Error("received error while reading from socket", "id", conn.Id, "error", err)
			}
			break
		}

		tcp.FromSockets <- types.TCPCommandWrapper{Command: cmd, Conn: conn}
	}
}

func (tcp *TCP) Start() {
	id := 0
	for {
		conn, err := tcp.listener.Accept()
		id++

		if err != nil {
			slog.Error("server error:", "error", err)
		}

		newConn := types.NewConnection(conn, id)
		slog.Debug("new connection", "id", newConn.Id)

		tcp.mutex.Lock()
		tcp.sockets = append(tcp.sockets, newConn)
		tcp.mutex.Unlock()

		go readConnection(tcp, &newConn)
	}
}
