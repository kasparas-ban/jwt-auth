package models

type Session struct {
	SessionId string `json:"sessionId" gorm:"primarykey"`
	UserId    uint   `json:"userId"`
}
