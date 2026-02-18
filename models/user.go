package models

import (
	"time"
)

type User struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null"     json:"email"`
	PasswordHash string    `gorm:"not null"                 json:"-"`
	CreatedAt    time.Time `                                json:"created_at"`
	UpdatedAt    time.Time `                                json:"-"`

	// Relazioni
	Results []Result `gorm:"foreignKey:UserID" json:"results,omitempty"`
}
