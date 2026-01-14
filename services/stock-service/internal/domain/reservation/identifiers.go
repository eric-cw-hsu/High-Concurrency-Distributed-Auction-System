package reservation

import (
	"github.com/samborkent/uuidv7"
)

// ReservationID represents unique identifier of a reservation
type ReservationID string

func NewReservationID() ReservationID {
	return ReservationID(uuidv7.New().String())
}

func ParseReservationID(id string) (ReservationID, error) {
	if id == "" {
		return "", ErrInvalidReservationID
	}
	if !uuidv7.IsValidString(id) {
		return "", ErrInvalidReservationID
	}
	return ReservationID(id), nil
}

func (id ReservationID) String() string {
	return string(id)
}

func (id ReservationID) IsEmpty() bool {
	return id == ""
}

// ProductID reference
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

// UserID reference
type UserID string

func ParseUserID(id string) (UserID, error) {
	if id == "" {
		return "", ErrInvalidUserID
	}
	return UserID(id), nil
}

func (id UserID) String() string {
	return string(id)
}

func (id UserID) IsEmpty() bool {
	return id == ""
}
