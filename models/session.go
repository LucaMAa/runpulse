package models

import (
	"math/rand"
	"strings"
	"time"
)

type SessionStatus string

const (
	StatusWaiting  SessionStatus = "waiting"  // solo START connesso
	StatusActive   SessionStatus = "active"   // START + END connessi
	StatusFinished SessionStatus = "finished" // sessione terminata
)

type Session struct {
	ID          uint          `gorm:"primaryKey;autoIncrement" json:"id"`
	Code        string        `gorm:"uniqueIndex;not null;size:8" json:"code"`
	Status      SessionStatus `gorm:"default:'waiting'"           json:"status"`
	StartUserID *uint         `gorm:"index"                        json:"start_user_id"`
	EndUserID   *uint         `gorm:"index"                        json:"end_user_id"`
	CreatedAt   time.Time     `                                    json:"created_at"`
	UpdatedAt   time.Time     `                                    json:"-"`

	// Relazioni
	StartUser *User    `gorm:"foreignKey:StartUserID" json:"start_user,omitempty"`
	EndUser   *User    `gorm:"foreignKey:EndUserID"   json:"end_user,omitempty"`
	Results   []Result `gorm:"foreignKey:SessionID"   json:"results,omitempty"`
}

// GenerateCode genera un codice alfanumerico casuale di 6 caratteri
func GenerateCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no 0/O/1/I per evitare confusione
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var sb strings.Builder
	for i := 0; i < 6; i++ {
		sb.WriteByte(chars[r.Intn(len(chars))])
	}
	return sb.String()
}
