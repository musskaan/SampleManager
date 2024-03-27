package model

import "github.com/lib/pq"

type SampleMapping struct {
	ClmSegments  pq.StringArray `json:"clm_segments" gorm:"type:text[]"`
	ItemId       string         `json:"item_id" gorm:"primaryKey:composite"`
	SampleItemId string         `json:"sample_item_id" gorm:"primaryKey:composite"`
}
