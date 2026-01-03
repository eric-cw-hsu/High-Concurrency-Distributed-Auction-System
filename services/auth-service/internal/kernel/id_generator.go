package kernel

type IDGenerator interface {
	NewUserID() UserID
	NewTokenID() TokenID
}
