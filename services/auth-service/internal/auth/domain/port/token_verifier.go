package port

import (
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/auth/domain/auth/model"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/kernel"
)

type TokenVerifier interface {
	VerifyAccess(token string) (*model.Access, error)
	VerifyRefresh(token string) (kernel.TokenID, error)
}
