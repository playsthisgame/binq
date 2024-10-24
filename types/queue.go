package types

import "gorm.io/gorm"

type Queue struct {
	gorm.Model
	Name          string `gorm:"index" json:"name"`
	maxPartitions string `json:"maxPartitions"`
}
