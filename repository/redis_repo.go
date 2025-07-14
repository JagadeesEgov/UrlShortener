package repository

import (
	"encoding/json"
	"os"

	"urlShortner/models"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

type RedisRepository struct {
	rdb *redis.Client
	ctx context.Context
}

func NewRedisRepository() *RedisRepository {
	client := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})
	return &RedisRepository{rdb: client, ctx: context.Background()}
}

func (r *RedisRepository) SaveURL(key string, req models.ShortenRequest) error {
	data, _ := json.Marshal(req)
	return r.rdb.Set(r.ctx, key, data, 0).Err()
}

func (r *RedisRepository) GetURL(key string) (string, error) {
	val, err := r.rdb.Get(r.ctx, key).Result()
	if err != nil {
		return "", err
	}
	var req models.ShortenRequest
	json.Unmarshal([]byte(val), &req)
	return req.URL, nil
} 