package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"os"

	"github.com/playsthisgame/binq/types"
)

type Config struct {
	Host      string
	Port      uint16
	PublicKey string
}

type BinqClient struct {
	conn *types.Connection
}

func NewBinqClient(conf *Config) (*BinqClient, error) {
	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)

	// Load the server's public certificate (trusted root cert)
	rootCAs := x509.NewCertPool()
	serverCert, err := os.ReadFile(conf.PublicKey)
	if err != nil {
		slog.Error("could not read server public cert", "error", err)
		return nil, err
	}

	ok := rootCAs.AppendCertsFromPEM(serverCert)
	if !ok {
		slog.Error("failed to append server cert to root pool")
		return nil, fmt.Errorf("could not append server cert to root CAs")
	}

	tlsConfig := &tls.Config{
		RootCAs: rootCAs,
		// ServerName: conf.Host, // Optional: required if the server cert has a specific CN/SAN
	}

	// Dial using TLS
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		slog.Error("TLS connection failed", "error", err)
		return nil, err
	}

	newConn := types.NewConnection(conn, 1)

	return &BinqClient{
		conn: &newConn,
	}, nil
}

func (c *BinqClient) Close() {
	c.conn.Close()
}

// create a queue
func (c *BinqClient) Create(queue types.Queue) error {
	data, err := queue.MarshalBinary()
	if err != nil {
		return err
	}

	cmd := &types.TCPCommand{
		Command: 1,
		Data:    []byte(data),
	}

	return sendCommand(c, cmd)
}

// publish message
func (c *BinqClient) Publish(message types.Message) error {
	data, err := message.MarshalBinary()
	if err != nil {
		return err
	}

	cmd := &types.TCPCommand{
		Command: 2,
		Data:    data,
	}

	return sendCommand(c, cmd)
}

func sendCommand(c *BinqClient, cmd *types.TCPCommand) error {
	data, err := cmd.MarshalBinary()
	if err != nil {
		slog.Error("Error marshalling data:", "error", err)
		return err
	}

	_, err = c.conn.Writer.Writer.Write(data)
	if err != nil {
		slog.Error("Error writing to server", "error", err)
		return err
	}
	return nil
}

type BinqConsumerClient struct {
	binqClient      *BinqClient
	consumerRequest *types.ConsumerRequest
}

func NewBinqConsumerClient(
	binqClient *BinqClient,
	consumerRequest *types.ConsumerRequest,
) (*BinqConsumerClient, error) {
	// establish connection as consumer client
	req, err := consumerRequest.MarshalBinary()
	if err != nil {
		return nil, err
	}

	cmd := &types.TCPCommand{
		Command: 3,
		Data:    req,
	}
	err = sendCommand(binqClient, cmd)
	if err != nil {
		return nil, err
	}

	return &BinqConsumerClient{
		binqClient:      binqClient,
		consumerRequest: consumerRequest,
	}, nil
}

// receive messages
func (c *BinqConsumerClient) Receive() (*types.MessageBatch, error) {
	cmd, err := c.binqClient.conn.Next()
	if err != nil {
		slog.Error("Error receiving messages", "error", err)
		return nil, err
	}

	var msgBatch types.MessageBatch
	err = msgBatch.UnmarshalBinary(cmd.Data)
	if err != nil {
		return nil, err
	}

	return &msgBatch, nil
}

func (c *BinqConsumerClient) Acknowledge(ackMessages *types.AckMessages) error {
	data, err := ackMessages.MarshalBinary()
	if err != nil {
		return err
	}

	cmd := &types.TCPCommand{
		Command: 4,
		Data:    data,
	}

	err = sendCommand(c.binqClient, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (c *BinqConsumerClient) Stop() {
	c.Stop()
}
