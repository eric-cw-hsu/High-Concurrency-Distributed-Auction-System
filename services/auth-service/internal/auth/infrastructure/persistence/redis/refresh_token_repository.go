package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/domain/auth/model"
	repository "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/domain/auth/repositroy"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/kernel"
	"github.com/redis/go-redis/v9"
)

type RedisRefreshRepo struct {
	rdb *redis.Client
}

var _ repository.RefreshTokenRepository = (*RedisRefreshRepo)(nil)

func NewRedisRefreshRepo(rdb *redis.Client) *RedisRefreshRepo {
	return &RedisRefreshRepo{rdb: rdb}
}

func (r *RedisRefreshRepo) Save(ctx context.Context, token model.RefreshToken) error {
	key := "refresh:" + string(token.ID)

	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	ttl := time.Until(token.ExpireAt)
	if ttl <= 0 {
		return errors.New("token already expired")
	}

	return r.rdb.Set(ctx, key, data, ttl).Err()
}

func (r *RedisRefreshRepo) Find(ctx context.Context, id kernel.TokenID) (*model.RefreshToken, error) {
	key := "refresh:" + string(id)
	data, err := r.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var token model.RefreshToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *RedisRefreshRepo) Revoke(ctx context.Context, id kernel.TokenID) error {
	key := "refresh:" + string(id)
	return r.rdb.Del(ctx, key).Err()
}
