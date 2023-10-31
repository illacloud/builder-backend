package cache

import (
	redis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Cache struct {
	IPZoneCache *IPZoneCache
}

func NewCache(redisDriver *redis.Client, logger *zap.SugaredLogger) *Cache {
	ipZoneCache := NewIPZoneCache(redisDriver, logger)
	return &Cache{
		IPZoneCache: ipZoneCache,
	}
}
