package database

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
)

type Session struct {
	SessionId string `json:"sessionId" gorm:"primarykey"`
	UserId    uint   `json:"userId"`
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

func GetCacheSession(ctx *gin.Context, sessionId string) (string, error) {
	val, err := SessionCache.Client.Get(ctx, sessionId).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}
