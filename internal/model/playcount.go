package model

import "time"

type PlayCount struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"uniqueIndex:idx_user_path"`
	FilePath   string    `json:"file_path" gorm:"size:512;uniqueIndex:idx_user_path"`
	Count      int       `json:"count"`
	LastPlayed time.Time `json:"last_played"`
}
