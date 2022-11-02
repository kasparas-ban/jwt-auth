package auth

import (
	"net/http"
	"time"

	env "jwt-auth/config"

	"github.com/dgrijalva/jwt-go"
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
