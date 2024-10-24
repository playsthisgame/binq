package types

import "gorm.io/gorm"

type Message struct {
	gorm.Model
	QueueName string `gorm:"index"`
	Partition int    `gorm:"index"`
	Data      []byte
}
