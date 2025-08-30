package wallet

type AddFundCommand struct {
	UserId      string  `json:"user_id"`
	Amount      float64 `json:"amount" binding:"required,gt=0"` // Amount must be greater than 0
	Description string  `json:"description"`
}

type SubtractFundCommand struct {
	UserId      string  `json:"user_id"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
}

type ProcessPaymentCommand struct {
	UserId  string  `json:"user_id"`
	OrderId string  `json:"order_id"`
	Amount  float64 `json:"amount" binding:"required,gt=0"`
}

type ProcessRefundCommand struct {
	UserId  string  `json:"user_id"`
	OrderId string  `json:"order_id"`
	Amount  float64 `json:"amount" binding:"required,gt=0"`
}

type SuspendWalletCommand struct {
	UserId string `json:"user_id"`
	Reason string `json:"reason"`
}

type ActivateWalletCommand struct {
	UserId string `json:"user_id"`
}

type CreateWalletCommand struct {
	UserId string `json:"user_id"`
}
