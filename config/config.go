package config

import (
	"os"
	"strconv"
)

var PORT,
	JWT_KEY,
	JWT_RESET_KEY,
	HOST_SERVER,
	EMAIL_DOMAIN,
	EMAIL_USER,
	EMAIL_PASS,
	CACHE_PASS string

var EMAIL_PORT int

func GetEnvAsInt(s string, defaultVal int) int {
	variable, err := strconv.Atoi(os.Getenv(s))
	if err != nil {
		return defaultVal
	}
	return variable
}
