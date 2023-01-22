package models

import (
	"fmt"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID         uint64 `gorm:"primaryKey"`
	Username   string `json:"username" gorm:"unique"`
	Email      string `json:"email" gorm:"unique"`
	Password   string `json:"password"`
	FullName   string `json:"full_name"`
	ProfilePic string `json:"profile_pic"`
	Location   string `json:"location"`
	Gender     string `json:"gender"`
	About      string `json:"about"`
	Birthday   string `json:"birthday"`
	CreatedAt  time.Time
	DeletedAt  time.Time
}

func (user *User) CheckPassword(providedPassword string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(providedPassword),
	)
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) // Change defaultCost ?
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func ValidateUsername(username string) error {
	match, err := regexp.MatchString("^[a-zA-Z0-9.]*$", username)
	if !match || err != nil {
		return fmt.Errorf("invalid username")
	}
	return nil
}

func ValidatePassword(password string) error {
	var lowercase, uppercase, digit, symbol bool
	r, _ := regexp.Compile("[@$!%*#?&^_-]")

	for _, c := range password {
		if c >= 'a' && c <= 'z' {
			lowercase = true
		}
		if c >= 'A' && c <= 'Z' {
			uppercase = true
		}
		if c >= '0' && c <= '9' {
			digit = true
		}
		if match := r.MatchString(string(c)); match {
			symbol = true
		}
	}

	if !(lowercase && uppercase && digit && symbol) {
		return fmt.Errorf("invalid signup form")
	}

	return nil
}
