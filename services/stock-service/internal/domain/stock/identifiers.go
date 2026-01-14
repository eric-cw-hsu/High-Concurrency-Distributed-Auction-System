package stock

import (
	"github.com/samborkent/uuidv7"
)

// ProductID represents a reference to a product
type ProductID string

func ParseProductID(id string) (ProductID, error) {
	if id == "" {
		return "", ErrInvalidProductID
	}
	if !uuidv7.IsValidString(id) {
		return "", ErrInvalidProductID
	}
	return ProductID(id), nil
}

func (id ProductID) String() string {
	return string(id)
}

func (id ProductID) IsEmpty() bool {
	return id == ""
}
