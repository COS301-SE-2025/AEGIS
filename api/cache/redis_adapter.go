package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct{ rdb *redis.Client }

func NewRedis(addr, password string, db int) *Redis {
	return &Redis{rdb: redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: db})}
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
