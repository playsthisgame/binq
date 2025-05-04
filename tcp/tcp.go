package tcp

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"slices"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"

	"github.com/playsthisgame/binq/types"
	"github.com/playsthisgame/binq/utils"
)

type TCP struct {
	passkey     uuid.UUID
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

func NewTCPServer(port uint16, passkey uuid.UUID) (*TCP, error) {
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
		passkey:     passkey,
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
			// remove from sockets
			for i := len(tcp.sockets) - 1; i >= 0; i-- {
				if tcp.sockets[i].Id == conn.Id {
					tcp.sockets = slices.Delete(tcp.sockets, i, i+1)
					break
				}
			}
			// TODO: send a drop command to remove the socket from the consumers if its a consumer socket
			tcp.FromSockets <- types.TCPCommandWrapper{Command: &types.TCPCommand{Command: 5}, Conn: conn}
			break
		}

		tcp.FromSockets <- types.TCPCommandWrapper{Command: cmd, Conn: conn}
	}
}

func (tcp *TCP) Start() {
	id := 0
	for {
		conn, err := tcp.listener.Accept()
		if err != nil {
			// if the listener is closed then this will log as an error
			slog.Error("server error:", "error", err)
			break
		}

		// check clients passkey
		if tcp.passkey != uuid.Nil && !tcp.IsClientConnected(conn) {
			if !hasValidPasskey(conn, tcp.passkey) {
				slog.Info("Invalid passkey")
				conn.Write([]byte("ERROR: Invalid passkey"))
				conn.Close()
				continue
			}
			slog.Info("passkey valid")
			conn.Write([]byte("passkey authenticated"))
		}
		id++

		newConn := types.NewConnection(conn, id)
		slog.Info("new connection", "id", newConn.Id, "connHash", newConn.ConnHash)

		tcp.mutex.Lock()
		tcp.sockets = append(tcp.sockets, newConn)
		tcp.mutex.Unlock()

		go readConnection(tcp, &newConn)
	}
}

func (tcp *TCP) IsClientConnected(conn net.Conn) bool {
	newHash := utils.GetConnectionHash(conn)

	for _, socket := range tcp.sockets {
		if socket.ConnHash == newHash {
			return true
		}
	}
	return false
}

func hasValidPasskey(conn net.Conn, passkey uuid.UUID) bool {
	if passkey == uuid.Nil {
		return true
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetReadDeadline(time.Time{}) // always clear

	buf := make([]byte, 16) // UUID is 16 bytes
	n, err := io.ReadFull(conn, buf)
	if err != nil {
		slog.Error("Error reading passkey",
			"remoteAddr", conn.RemoteAddr(),
			"bytesRead", n,
			"error", err,
		)
		return false
	}

	clientPasskey, err := uuid.FromBytes(buf)
	if err != nil {
		slog.Error("Error converting passkey",
			"remoteAddr", conn.RemoteAddr(),
			"bytesRead", n,
			"error", err,
		)
		return false
	}

	return clientPasskey == passkey
}
