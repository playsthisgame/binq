package types

import (
	"errors"
	"fmt"
)

type ConsumerSocket struct {
	Instance   int
	Partitions []int
	Conn       Connection
	QueueName  string
}

func NewConsumerSocket(instance int, totalInstances int, maxPartitions int, queueName string, conn Connection) (*ConsumerSocket, error) {
	if instance > maxPartitions {
		return nil, errors.New(fmt.Sprintf("number of instances %d cannot exceed maxPartitions %d \n", instance, maxPartitions))
	}
	return &ConsumerSocket{
		Instance:   instance,
		Conn:       conn,
		Partitions: SetPartitions(instance, totalInstances, maxPartitions),
		QueueName:  queueName,
	}, nil

}

func SetPartitions(instance int, totalInstances int, maxPartitions int) []int {
	partitions := []int{}
	for i := instance; i <= maxPartitions; i += totalInstances {
		partitions = append(partitions, i)
	}
	return partitions
}

type ReceiveRequest struct {
	queueName string
	batchSize int
}

func NewReceiveRequest(queueName string, batchSize int) *ReceiveRequest {
	return &ReceiveRequest{
		queueName: queueName,
		batchSize: batchSize,
	}
}
