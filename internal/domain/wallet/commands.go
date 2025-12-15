package wallet

type AddFundCommand struct {
	UserID      string  `json:"user_id"`
	Amount      float64 `json:"amount" binding:"required,gt=0"` // Amount must be greater than 0
	Description string  `json:"description"`
}

type SubtractFundCommand struct {
	UserID      string  `json:"user_id"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
}

type ProcessPaymentCommand struct {
	UserID  string  `json:"user_id"`
	OrderId string  `json:"order_id"`
	Amount  float64 `json:"amount" binding:"required,gt=0"`
}

type ProcessRefundCommand struct {
	UserID  string  `json:"user_id"`
	OrderId string  `json:"order_id"`
	Amount  float64 `json:"amount" binding:"required,gt=0"`
}

type SuspendWalletCommand struct {
	UserID string `json:"user_id"`
	Reason string `json:"reason"`
}

type ActivateWalletCommand struct {
	UserID string `json:"user_id"`
}

type CreateWalletCommand struct {
	UserID string `json:"user_id"`
}
