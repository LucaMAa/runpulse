package models

import "time"

type Result struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID   uint      `gorm:"index;not null"           json:"session_id"`
	UserID      uint      `gorm:"index;not null"           json:"user_id"`
	Role        string    `gorm:"not null;size:10"         json:"role"`     // "start" | "end"
	TimeSeconds float64   `gorm:"not null"                 json:"time_seconds"`
	CreatedAt   time.Time `                                json:"created_at"`

	Session Session `gorm:"foreignKey:SessionID" json:"session,omitempty"`
	User    User    `gorm:"foreignKey:UserID"    json:"user,omitempty"`
}
