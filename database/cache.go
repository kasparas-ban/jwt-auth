package database

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"github.com/go-sql-driver/mysql"
)

type Session struct {
	SessionId string `json:"sessionId" gorm:"primarykey"`
	UserId    uint64 `json:"userId"`
}

type cache struct {
	Client *redis.Client
}

var SessionCache = &cache{}

func (c *cache) Connect(url string) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		panic("Failed to connect to Redis cache")
	}

	c.Client = redis.NewClient(opt)
}

func SaveCacheSession(ctx *gin.Context, s *Session) error {
	if res := SessionCache.Client.Set(ctx, s.SessionId, fmt.Sprint(s.UserId), 0); res.Err() != nil {
		return res.Err()
	}
	return nil
}

func ReadSessionCache(ctx *gin.Context, sessionId string) (Session, error) {
	var session Session
	val, err := SessionCache.Client.Get(ctx, sessionId).Result()
	if err != nil {
		return session, err
	}

	id, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return session, err
	}
	session = Session{sessionId, id}

	return session, nil
}

func ReadSessionDB(ctx *gin.Context, sessionId string) (*Session, error) {
	var session *Session
	result := SessionDB.Instance.Where("session_id = ?", sessionId).First(&session)
	if result.Error != nil {
		return session, fmt.Errorf("no session found")
	}
	return session, nil
}

func GenerateSession(userId uint64) (Session, error) {
	session := Session{}
	b := make([]byte, 20)
	_, err := rand.Read(b)
	if err != nil {
		return session, fmt.Errorf("failed to generate a random number")
	}
	session.SessionId = base64.URLEncoding.EncodeToString(b)
	session.UserId = userId
	return session, nil
}

func ValidateSession(ctx *gin.Context, sessionId string) error {
	// Check cache for sessionId
	if session, err := ReadSessionCache(ctx, sessionId); err == nil {
		ctx.Set("userID", session.UserId)
		return nil
	}

	// Check sessionDB for sessionId
	session, err := ReadSessionDB(ctx, sessionId)
	if err != nil {
		return fmt.Errorf("no session found")
	}
	ctx.Set("userID", session.UserId)

	// Session was found in DB, save it to cache
	if err := SaveCacheSession(ctx, session); err != nil {
		return err
	}

	return nil
}

func SaveSession(ctx *gin.Context, s *Session) error {
	result := SessionDB.Instance.Create(&s)

	// If duplicate, don't return an error
	var mysqlErr *mysql.MySQLError
	if !(result.Error == nil || (errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062)) {
		return fmt.Errorf("failed to save session to the database")
	}

	// Save to cache
	if err := SaveCacheSession(ctx, s); err != nil {
		fmt.Println("SAVING TO CACHE FAILED")
		return err
	}

	return nil
}

func RemoveUserSession(ctx *gin.Context, sessionId string) error {
	// Remove session from sessionDB
	result := SessionDB.Instance.Where("session_id = ?", sessionId).Delete(&Session{})
	if result.Error != nil {
		return result.Error
	}

	// Remove session from session cache
	_, err := SessionCache.Client.Del(ctx, sessionId).Result()
	if err != nil {
		return err
	}

	return nil
}

func ExtractSessionId(cookie *http.Cookie) string {
	sessionId := cookie.Value
	sessionId = strings.Replace(sessionId, "%3D", "=", -1)
	return sessionId
}
