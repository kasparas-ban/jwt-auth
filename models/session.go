package models

import (
	"time"
)

type Session struct {
	SessionId string `json:"sessionId" gorm:"primarykey"`
	Username  string `json:"username"`
	UpdatedAt time.Time
}

func (user *Session) CheckSession(sessionId string) error {

	// var session Session
	// database.SessionDB.Instance.Where("sessionId = ?", sessionId).First(&session)
	return nil
}
