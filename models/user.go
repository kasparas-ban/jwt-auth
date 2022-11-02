package models

import (
	"fmt"
	"regexp"

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
