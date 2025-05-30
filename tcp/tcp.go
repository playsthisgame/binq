package tcp

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"path/filepath"
	"slices"
	"sync"
	"syscall"

	"github.com/playsthisgame/binq/cert"
	"github.com/playsthisgame/binq/types"
	"github.com/playsthisgame/binq/utils"
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

func NewTCPServer(port uint16, certPath string) (*TCP, error) {
	err := cert.Setup(certPath)
	if err != nil {
		slog.Error("error creating cert", "error", err)
	}

	config, err := tlsServerConfig(certPath)
	if err != nil {
		slog.Error("error getting tls server  config", "error", err)
	}

	listener, err := tls.Listen("tcp", fmt.Sprintf(":%d", port), config)
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

func loadCertPair(certPath, keyPath string) (tls.Certificate, error) {
	return tls.LoadX509KeyPair(certPath, keyPath)
}

func tlsServerConfig(certPath string) (*tls.Config, error) {
	certMap := map[string]tls.Certificate{}
	// Load multiple certs
	localCert, err := loadCertPair(
		filepath.Join(certPath, "localhost.pem"),
		filepath.Join(certPath, "localhost-key.pem"),
	)
	if err != nil {
		return nil, err
	}
	certMap["localhost"] = localCert

	remoteCert, err := loadCertPair(
		filepath.Join(certPath, "remotehost.pem"),
		filepath.Join(certPath, "remotehost-key.pem"),
	)
	if err != nil {
		return nil, err
	}
	certMap["remote.host"] = remoteCert

	config := &tls.Config{
		GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			// Match based on SNI
			if cert, ok := certMap[hello.ServerName]; ok {
				return &cert, nil
			}
			// Fallback cert
			return &localCert, nil
		},
		MinVersion: tls.VersionTLS12,
	}
	return config, nil
}
