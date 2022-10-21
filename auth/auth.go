package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	env "jwt-auth/config"
	db "jwt-auth/database"
	"jwt-auth/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-sql-driver/mysql"
)

type AuthError struct {
	Status int
	Msg    string
	Err    error
}

type JWTClaim struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	HashPass string `json:"hashPass"`
	jwt.StandardClaims
}

type JWTResetClaim struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

// === Sign Up JWT ===

func GenerateJWT(name, email, pass string) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &JWTClaim{
		Email:    email,
		Username: name,
		HashPass: pass,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(env.JWT_KEY))
	return tokenString, err
}

func ValidateJWT(signedToken string) (claims *JWTClaim, authErr AuthError) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(env.JWT_KEY), nil
		},
	)

	if err != nil {
		return claims,
			AuthError{
				Status: http.StatusInternalServerError,
				Msg:    "Failed to parse the token",
				Err:    err,
			}
	}

	claims, ok := token.Claims.(*JWTClaim)
	if !ok {
		return claims,
			AuthError{
				Status: http.StatusInternalServerError,
				Msg:    "Failed to parse the claims",
			}
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		return claims,
			AuthError{
				Status: http.StatusUnauthorized,
				Msg:    "Token has expired",
			}
	}

	return
}

// === Reset Password JWT ===

func GenerateResetJWT(email string) (string, error) {
	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &JWTResetClaim{
		Email: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(env.JWT_RESET_KEY))
	return tokenString, err
}

func ValidateResetJWT(signedToken string) (claims *JWTResetClaim, authErr AuthError) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTResetClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(env.JWT_RESET_KEY), nil
		},
	)

	if err != nil {
		return claims,
			AuthError{
				Status: http.StatusInternalServerError,
				Msg:    "Failed to parse the token",
				Err:    err,
			}
	}

	claims, ok := token.Claims.(*JWTResetClaim)
	if !ok {
		return claims,
			AuthError{
				Status: http.StatusInternalServerError,
				Msg:    "Failed to parse the claims",
			}
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		return claims,
			AuthError{
				Status: http.StatusUnauthorized,
				Msg:    "Token has expired",
			}
	}

	return
}

// === Session management ===

func GenerateSession(userId uint) (models.Session, error) {
	session := models.Session{}
	b := make([]byte, 20)
	_, err := rand.Read(b)
	if err != nil {
		return session, fmt.Errorf("failed to generate a random number")
	}
	session.SessionId = base64.URLEncoding.EncodeToString(b)
	session.UserId = userId
	return session, nil
}

func ValidateSession(sessionId string) error {
	// Check cache for sessionId

	// Check sessionDB for sessionId
	var session *models.Session
	err := db.SessionDB.Instance.Where("session_id = ?", sessionId).First(&session).Error
	if err != nil {
		return fmt.Errorf("no session found")
	}

	// Save session to cache
	return nil
}

func SaveSession(s models.Session) error {
	result := db.SessionDB.Instance.Create(&s)

	// If duplicate, don't return an error
	var mysqlErr *mysql.MySQLError
	if result.Error == nil || (errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062) {
		return nil
	}

	return fmt.Errorf("failed to save session to the database")
}
