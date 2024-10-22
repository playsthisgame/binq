package types

import "gorm.io/gorm"

type Message struct {
	gorm.Model
	Partition int16
	Data      []byte
}
