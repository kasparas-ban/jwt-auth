package database

import (
	"jwt-auth/models"
	"time"
)

type ProfileInfo struct {
	ID         uint64
	Username   string
	Email      string
	FullName   string
	ProfilePic string
	Location   string
	Gender     string
	About      string
	Birthday   string
}

type FriendInfo struct {
	ID         uint64
	Username   string
	FullName   string
	ProfilePic string
	Location   string
}

type Relation struct {
	RequesterId uint64 `json:"requester_id"`
	AddresseeId uint64 `json:"addressee_id"`
	CreateTime  time.Time
}

func GetProfileInfo(userID uint64) (ProfileInfo, error) {
	var user models.User
	var profile ProfileInfo

	record := MainDB.Instance.Where("id = ?", 1).First(&user)
	if record.Error != nil {
		return profile, record.Error
	}

	profile = ProfileInfo{
		ID:         user.ID,
		Username:   user.Username,
		Email:      user.Email,
		FullName:   user.FullName,
		ProfilePic: user.ProfilePic,
		Location:   user.Location,
		Gender:     user.Gender,
		About:      user.About,
		Birthday:   user.Birthday,
	}

	return profile, nil
}

func GetAllFriendships(userID uint64, limit int, offset int) (*[]FriendInfo, error) {
	var allFriends *[]FriendInfo

	subQuery := MainDB.Instance.Select("addressee_id").Where("requester_id = ?", userID).Table("friendships")
	err := MainDB.Instance.Select("id", "username", "full_name", "profile_pic", "location").Where("id IN (?)", subQuery).Table("users").Limit(limit).Offset(offset).Find(&allFriends).Error
	if err != nil {
		return nil, err
	}

	return allFriends, nil
}
