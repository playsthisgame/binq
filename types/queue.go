package types

import (
	"encoding/json"

	"gorm.io/gorm"
)

type Queue struct {
	gorm.Model
	Name          string `gorm:"index" json:"name"`
	MaxPartitions int    `json:"maxPartitions"`
}

func (m *Queue) MarshalBinary() (bytes []byte, err error) {
	return json.Marshal(m)
}

func (m *Queue) UnmarshalBinary(bytes []byte) error {
	err := json.Unmarshal(bytes, m)
	if err != nil {
		return err
	}
	return nil
}
