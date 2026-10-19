package user

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisUserCache struct {
	client *redis.Client
	ttl    time.Duration
	prefix string
}

func NewRedisUserCache(client *redis.Client, ttl time.Duration, prefix string) UserCache {
	return &RedisUserCache{
		client: client,
		ttl:    ttl,
		prefix: prefix,
	}
}

func (c *RedisUserCache) key(userID int64) string {
	return c.prefix + ":user:v1:" + strconv.FormatInt(userID, 10)
}

func (c *RedisUserCache) Get(ctx context.Context, userID int64) (*User, error) {
	key := c.key(userID)
	res, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheMiss
		}
		return nil, err
	}

	user := &User{}
	if err := json.Unmarshal([]byte(res), user); err != nil {
		return nil, err
	}

	return user, nil
}

func (c *RedisUserCache) Set(ctx context.Context, user *User) error {
	key := c.key(user.ID)
	ttl := c.ttl

	val, err := json.Marshal(user)
	if err != nil {
		return err
	}

	if err := c.client.Set(ctx, key, val, ttl).Err(); err != nil {
		return err
	}
	return nil
}

func (c *RedisUserCache) Delete(ctx context.Context, userID int64) error {
	key := c.key(userID)
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return err
	}
	return nil
}
