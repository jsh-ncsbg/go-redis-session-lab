package session

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisSessionCache struct {
	client *redis.Client
	prefix string
}

func NewRedisSessionCache(client *redis.Client, prefix string) SessionCache {
	return &RedisSessionCache{
		client: client,
		prefix: prefix,
	}
}

func (c *RedisSessionCache) key(tokenHash string) string {
	return c.prefix + ":session:v1:" + tokenHash
}

func (c *RedisSessionCache) Get(ctx context.Context, tokenHash string) (*Session, error) {
	key := c.key(tokenHash)
	res, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheMiss
		}
		return nil, err
	}

	session := &Session{}
	if err := json.Unmarshal([]byte(res), session); err != nil {
		return nil, err
	}

	if session.ExpiresAt.Before(time.Now()) {
		return nil, ErrCacheMiss
	}

	return session, nil
}

func (c *RedisSessionCache) Set(ctx context.Context, session *Session) error {
	key := c.key(session.TokenHash)
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return c.Delete(ctx, session.TokenHash)
	}

	val, err := json.Marshal(session)
	if err != nil {
		return err
	}

	if err := c.client.Set(ctx, key, val, ttl).Err(); err != nil {
		return err
	}
	return nil
}

func (c *RedisSessionCache) Delete(ctx context.Context, tokenHash string) error {
	key := c.key(tokenHash)
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return err
	}
	return nil
}
