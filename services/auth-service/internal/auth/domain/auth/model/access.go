package model

import (
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/kernel"
)

type Access struct {
	UserID   kernel.UserID
	Role     kernel.Role
	IssuedAt time.Time
	ExpireAt time.Time
}
