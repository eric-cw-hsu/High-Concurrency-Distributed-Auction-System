package model

import (
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/kernel"
)

type RefreshToken struct {
	ID       kernel.TokenID
	UserID   kernel.UserID
	ExpireAt time.Time
	Revoked  bool
}
