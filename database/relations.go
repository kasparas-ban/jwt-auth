package database

import (
	"jwt-auth/models"
	"time"
)

type Relation struct {
	RequesterId uint64 `json:"requester_id"`
	AddresseeId uint64 `json:"addressee_id"`
	CreateTime  time.Time
}

func GetAllFriendships(userID uint64, limit int, offset int) (*[]models.User, error) {
	var allFriends *[]models.User

	subQuery := MainDB.Instance.Select("addressee_id").Where("requester_id = ?", userID).Table("friendships")
	err := MainDB.Instance.Select("username", "email").Where("id IN (?)", subQuery).Limit(limit).Offset(offset).Find(&allFriends).Error
	if err != nil {
		return nil, err
	}

	return allFriends, nil
}
