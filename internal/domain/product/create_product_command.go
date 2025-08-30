package product

type CreateProductCommand struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
