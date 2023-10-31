package cache

import (
	"context"

	redis "github.com/redis/go-redis/v9"

	"go.uber.org/zap"
)

const (
	IP_ZONE_HASH_KEY = "ip_zone"
)

type IPZoneCache struct {
	logger  *zap.SugaredLogger
	cache   *redis.Client
	context context.Context
}

func NewIPZoneCache(cache *redis.Client, logger *zap.SugaredLogger) *IPZoneCache {
	return &IPZoneCache{
		logger:  logger,
		cache:   cache,
		context: context.Background(),
	}
}

func (c *IPZoneCache) SetIPZone(ip string, zone string) error {
	return c.cache.HSet(c.context, IP_ZONE_HASH_KEY, ip, zone).Err()
}

func (c *IPZoneCache) GetIPZone(ip string) (string, error) {
	zone, errInGet := c.cache.HGet(c.context, IP_ZONE_HASH_KEY, ip).Result()
	if errInGet == redis.Nil {
		return "", nil
	} else if errInGet != nil {
		return "", errInGet
	}
	return zone, nil
}
