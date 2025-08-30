package product

type Product struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   int64  `json:"created_at"` // Unix timestamp in seconds
	UpdatedAt   int64  `json:"updated_at"` // Unix timestamp in seconds
}
