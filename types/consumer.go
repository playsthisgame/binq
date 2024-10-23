package types

import (
	"errors"
	"fmt"
)

type ConsumerSocket struct {
	Instance   int
	Partitions []int
	conn       Connection
}

func NewConsumerSocket(instance int, totalInstances int, maxPartitions int, conn Connection) (*ConsumerSocket, error) {
	if instance > maxPartitions {
		return nil, errors.New(fmt.Sprintf("number of instances %d cannot exceed maxPartitions %d \n", instance, maxPartitions))
	}
	return &ConsumerSocket{
		Instance:   instance,
		conn:       conn,
		Partitions: SetPartitions(instance, totalInstances, maxPartitions),
	}, nil

}

func SetPartitions(instance int, totalInstances int, maxPartitions int) []int {
	partitions := []int{}
	for i := instance; i <= maxPartitions; i += totalInstances {
		partitions = append(partitions, i)
	}
	return partitions
}
