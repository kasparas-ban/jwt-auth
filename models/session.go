package models

type Session struct {
	SessionId string `json:"sessionId" gorm:"primarykey"`
	Username  string `json:"username"`
}
