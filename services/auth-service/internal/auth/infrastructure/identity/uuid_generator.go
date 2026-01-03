package identity

import (
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/kernel"
	"github.com/samborkent/uuidv7"
)

type UUIDGenerator struct {
}

func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

func (g *UUIDGenerator) NewUserID() kernel.UserID {
	return kernel.UserID(uuidv7.New().String())
}

func (g *UUIDGenerator) NewTokenID() kernel.TokenID {
	return kernel.TokenID(uuidv7.New().String())
}
