package redis

import (
	"encoding/json"
	"time"

	"github.com/go-redis/redis"
)

type RedisHandler struct {
	client   *redis.Client
	duration time.Duration
}

// NewHandler ...
func NewHandler(client *redis.Client, duration time.Duration) *RedisHandler {
	return &RedisHandler{
		client:   client,
		duration: duration,
	}
}

// Set ...
func (r *RedisHandler) Set(key string, value interface{}) error {
	jbyt, err := json.Marshal(value)
	if err != nil {
		return err
	}
	status := r.client.Set(key, string(jbyt), r.duration)
	return status.Err()
}

// Get ...
func (r *RedisHandler) Get(key string) (res []byte, err error) {
	cmd := r.client.Get(key)
	return cmd.Bytes()
}
