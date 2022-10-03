package models

import (
	b64 "encoding/base64"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `json:"username"`
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"password"`
}

func (user *User) CheckPassword(providedPassword string) error {
	// Decrypt hashed password
	hashedPass, err := b64.StdEncoding.DecodeString(user.Password)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword(
		hashedPass,
		[]byte(providedPassword),
	)
	return err
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) // Need salt ?
	if err != nil {
		return "", err
	}
	return b64.StdEncoding.EncodeToString(hashedPassword), nil
}
