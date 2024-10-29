package types

import (
	"encoding/json"

	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	QueueName string `gorm:"index"`
	Partition int    `gorm:"index"`
	Data      []byte
}

func (m *Message) MarshalBinary() (bytes []byte, err error) {
	return json.Marshal(m)
}

func (m *Message) UnmarshalBinary(bytes []byte) error {
	err := json.Unmarshal(bytes, m)
	if err != nil {
		return err
	}
	return nil
}

type MessageBatch struct {
	Messages []Message
}

func (m *MessageBatch) MarshalBinary() (bytes []byte, err error) {
	return json.Marshal(m)
}

func (m *MessageBatch) UnmarshalBinary(bytes []byte) error {
	err := json.Unmarshal(bytes, m)
	if err != nil {
		return err
	}
	return nil
}

type AckMessages struct {
	MessageIds []uint
}

func (a *AckMessages) MarshalBinary() (bytes []byte, err error) {
	return json.Marshal(a)
}

func (a *AckMessages) UnmarshalBinary(bytes []byte) error {
	err := json.Unmarshal(bytes, a)
	if err != nil {
		return err
	}
	return nil
}
