package models

import "time"

type Record struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint      `gorm:"index;not null"           json:"user_id"`
	DistanceM   int       `gorm:"not null"                 json:"distance_m"`
	TimeSeconds float64   `gorm:"not null"                 json:"time_seconds"`
	Notes       string    `gorm:"size:255"                 json:"notes"`
	RecordedAt  time.Time `gorm:"not null"                 json:"recorded_at"`
	CreatedAt   time.Time `                                json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
