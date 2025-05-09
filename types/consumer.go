package types

import (
	"encoding/json"
	"errors"
	"fmt"
)

type ConsumerSocket struct {
	Instance   int
	Partitions []int
	Conn       Connection
	QueueName  string
}

func NewConsumerSocket(
	instance int,
	totalInstances int,
	maxPartitions int,
	queueName string,
	conn Connection,
) (*ConsumerSocket, error) {
	if totalInstances > maxPartitions {
		return nil, errors.New(
			fmt.Sprintf(
				"number of instances %d cannot exceed maxPartitions %d \n",
				instance,
				maxPartitions,
			),
		)
	}
	return &ConsumerSocket{
		Instance:   instance,
		Conn:       conn,
		Partitions: SetPartitions(instance, totalInstances, maxPartitions),
		QueueName:  queueName,
	}, nil
}

// TODO: this method seems to be wrong, when ousting
func SetPartitions(instance int, totalInstances int, maxPartitions int) []int {
	partitions := []int{}
	for i := instance; i <= maxPartitions; i += totalInstances {
		partitions = append(partitions, i)
	}
	return partitions
}

type ConsumerRequest struct {
	QueueName string
	BatchSize int
}

func (r *ConsumerRequest) MarshalBinary() (data []byte, err error) {
	return json.Marshal(r)
}

func (r *ConsumerRequest) UnmarshalBinary(bytes []byte) error {
	err := json.Unmarshal(bytes, r)
	if err != nil {
		return err
	}
	return nil
}

type ConsumerAck struct {
	QueueName  string
	MessageIds []string
}

func (a *ConsumerAck) MarshalBinary() (data []byte, err error) {
	return json.Marshal(a)
}

func (a *ConsumerAck) UnmarshalBinary(bytes []byte) error {
	err := json.Unmarshal(bytes, a)
	if err != nil {
		return err
	}
	return nil
}
