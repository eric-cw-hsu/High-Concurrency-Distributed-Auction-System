package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/domain/auth/model"
	repository "github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/domain/auth/repositroy"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/common/logger"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/kernel"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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

	logger.DebugContext(ctx, "saving refresh token to redis",
		zap.String("token_id", string(token.ID)),
		zap.String("key", key),
	)

	data, err := json.Marshal(token)
	if err != nil {
		logger.ErrorContext(ctx, "failed to marshal refresh token",
			zap.String("token_id", string(token.ID)),
			zap.Error(err),
		)
		return err
	}

	ttl := time.Until(token.ExpireAt)
	if ttl <= 0 {
		logger.WarnContext(ctx, "refresh token already expired",
			zap.String("token_id", string(token.ID)),
			zap.Time("expire_at", token.ExpireAt),
		)
		return errors.New("token already expired")
	}

	if err := r.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		logger.ErrorContext(ctx, "failed to save refresh token to redis",
			zap.String("token_id", string(token.ID)),
			zap.String("key", key),
			zap.Duration("ttl", ttl),
			zap.Error(err),
		)
		return err
	}

	logger.DebugContext(ctx, "refresh token saved successfully",
		zap.String("token_id", string(token.ID)),
		zap.Duration("ttl", ttl),
	)

	return nil
}

func (r *RedisRefreshRepo) Find(ctx context.Context, id kernel.TokenID) (*model.RefreshToken, error) {
	key := "refresh:" + string(id)

	logger.DebugContext(ctx, "finding refresh token in redis",
		zap.String("token_id", string(id)),
		zap.String("key", key),
	)

	data, err := r.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			logger.DebugContext(ctx, "refresh token not found in redis",
				zap.String("token_id", string(id)),
			)
			return nil, nil
		}

		logger.ErrorContext(ctx, "failed to get refresh token from redis",
			zap.String("token_id", string(id)),
			zap.String("key", key),
			zap.Error(err),
		)
		return nil, err
	}

	var token model.RefreshToken
	if err := json.Unmarshal(data, &token); err != nil {
		logger.ErrorContext(ctx, "failed to unmarshal refresh token",
			zap.String("token_id", string(id)),
			zap.Error(err),
		)
		return nil, err
	}

	logger.DebugContext(ctx, "refresh token found",
		zap.String("token_id", string(id)),
	)

	return &token, nil
}

func (r *RedisRefreshRepo) Revoke(ctx context.Context, id kernel.TokenID) error {
	key := "refresh:" + string(id)

	logger.InfoContext(ctx, "revoking refresh token",
		zap.String("token_id", string(id)),
	)

	if err := r.rdb.Del(ctx, key).Err(); err != nil {
		logger.ErrorContext(ctx, "failed to revoke refresh token",
			zap.String("token_id", string(id)),
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}

	logger.InfoContext(ctx, "refresh token revoked successfully",
		zap.String("token_id", string(id)),
	)

	return nil
}
