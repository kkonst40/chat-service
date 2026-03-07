package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisUserLoginCache struct {
	client    *redis.Client
	ttl       time.Duration
	keyPrefix string
}

func NewRedisUserLoginCache(client *redis.Client, ttl time.Duration) *RedisUserLoginCache {
	return &RedisUserLoginCache{
		client:    client,
		ttl:       ttl,
		keyPrefix: "user_login:",
	}
}

func (c *RedisUserLoginCache) GetUserLogins(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	result := make(map[uuid.UUID]string, len(userIDs))

	if c == nil || c.client == nil || len(userIDs) == 0 {
		return result, nil
	}

	keys := make([]string, len(userIDs))
	for i, id := range userIDs {
		keys[i] = c.formatKey(id)
	}

	values, err := c.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	for i, v := range values {
		if v == nil {
			continue
		}

		login, ok := v.(string)
		if !ok || login == "" {
			continue
		}

		result[userIDs[i]] = login
	}

	return result, nil
}

func (c *RedisUserLoginCache) SetUserLogins(ctx context.Context, logins map[uuid.UUID]string) error {
	if c == nil || c.client == nil || len(logins) == 0 {
		return nil
	}

	pipe := c.client.Pipeline()
	for id, login := range logins {
		key := c.formatKey(id)
		pipe.Set(ctx, key, login, c.ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (c *RedisUserLoginCache) formatKey(id uuid.UUID) string {
	return fmt.Sprintf("%s%s", c.keyPrefix, id.String())
}

