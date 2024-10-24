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

func (m *Message) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}

func (m *Message) UnmarshalBinary(bytes []byte) error {
	return json.Unmarshal(bytes, &Message{})
}

type MessageBatch struct {
	Messages []Message
}

func (m *MessageBatch) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}

func (m *MessageBatch) UnmarshalBinary(bytes []byte) error {
	return json.Unmarshal(bytes, &MessageBatch{})
}
