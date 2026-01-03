package repository

import (
	"context"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/domain/auth/model"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/kernel"
)

type RefreshTokenRepository interface {
	Save(ctx context.Context, token model.RefreshToken) error
	Find(ctx context.Context, id kernel.TokenID) (*model.RefreshToken, error)
	Revoke(ctx context.Context, id kernel.TokenID) error
}
