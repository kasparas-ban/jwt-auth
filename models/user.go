package models

import (
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
	// hashedPass, err := b64.StdEncoding.DecodeString(user.Password)
	// if err != nil {
	// 	return err
	// }

	err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(providedPassword),
	)
	return err
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) // Need salt (seems like its implemented) ? Change defaultCost ?
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
