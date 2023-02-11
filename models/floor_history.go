package models

import "time"

type FloorHistory struct {
	/// base info
	ID        int       `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Reason    string    `json:"reason"`
	Content   string    `json:"content" gorm:"size:15000"`
	FloorID   int       `json:"floor_id"`
	// The one who modified the floor
	UserID int `json:"user_id"`
}

type FloorHistorySlice []*FloorHistory
