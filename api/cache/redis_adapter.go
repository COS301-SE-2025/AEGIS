package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient interface for mocking
type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
}

type Redis struct{ rdb RedisClient }

func NewRedis(addr, password string, db int) *Redis {
	return &Redis{rdb: redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})}
}

// NewRedisWithClient allows dependency injection for testing
func NewRedisWithClient(client RedisClient) *Redis {
	return &Redis{rdb: client}
}

func (c *Redis) Get(ctx context.Context, key string) (string, bool, error) {
	res, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return res, true, nil
}

func (c *Redis) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *Redis) Del(ctx context.Context, keys ...string) (int, error) {
	res := c.rdb.Del(ctx, keys...)
	return int(res.Val()), res.Err()
}
