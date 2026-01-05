package product

import (
	"errors"

	"github.com/samborkent/uuidv7"
)

type ProductID string

func NewProductID() ProductID {
	return ProductID(uuidv7.New().String())
}

func ParseProductID(id string) (ProductID, error) {
	if id == "" {
		return "", ErrInvalidProductID
	}
	if uuidv7.IsValidString(id) == false {
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

func (id ProductID) Equals(other ProductID) bool {
	return id == other
}

type SellerID string

func ParseSellerID(id string) (SellerID, error) {
	if id == "" {
		return "", errors.New("seller id cannot be empty")
	}
	return SellerID(id), nil
}

func (id SellerID) String() string {
	return string(id)
}

func (id SellerID) IsEmpty() bool {
	return id == ""
}

func (id SellerID) Equals(other SellerID) bool {
	return id == other
}
