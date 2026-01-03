package port

import (
	"time"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/auth-service/internal/kernel"
)

type TokenIssuer interface {
	// Return AccessToken, Error
	IssueAccess(userID kernel.UserID, role kernel.Role) (string, error)
	// Return RefreshToken, Error
	IssueRefresh(tokenID kernel.TokenID, userID kernel.UserID) (string, time.Time, error)
}
