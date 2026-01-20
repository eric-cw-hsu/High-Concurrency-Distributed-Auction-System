package order

import (
	"github.com/samborkent/uuidv7"
)

// OrderID represents a unique order identifier
type OrderID struct {
	value string
}

func NewOrderID() OrderID {
	return OrderID{value: uuidv7.New().String()}
}

func ParseOrderID(id string) (OrderID, error) {
	if id == "" {
		return OrderID{}, ErrEmptyOrderID
	}
	if uuidv7.IsValidString(id) == false {
		return OrderID{}, ErrInvalidOrderIDFormat
	}
	return OrderID{value: id}, nil
}

func (id OrderID) String() string {
	return id.value
}

func (id OrderID) Equals(other OrderID) bool {
	return id.value == other.value
}

// ReservationID represents a reservation identifier
type ReservationID struct {
	value string
}

func ParseReservationID(id string) (ReservationID, error) {
	if id == "" {
		return ReservationID{}, ErrEmptyReservationID
	}
	return ReservationID{value: id}, nil
}

func (id ReservationID) String() string {
	return id.value
}

// UserID represents a user identifier
type UserID struct {
	value string
}

func ParseUserID(id string) (UserID, error) {
	if id == "" {
		return UserID{}, ErrEmptyUserID
	}
	return UserID{value: id}, nil
}

func (id UserID) String() string {
	return id.value
}

// ProductID represents a product identifier
type ProductID struct {
	value string
}

func ParseProductID(id string) (ProductID, error) {
	if id == "" {
		return ProductID{}, ErrEmptyProductID
	}
	return ProductID{value: id}, nil
}

func (id ProductID) String() string {
	return id.value
}

// Money represents a monetary value
type Money struct {
	amount   int64 // in cents
	currency string
}

func NewMoney(amount int64, currency string) (Money, error) {
	if amount < 0 {
		return Money{}, ErrNegativeAmount
	}
	if currency == "" {
		return Money{}, ErrEmptyCurrency
	}
	return Money{
		amount:   amount,
		currency: currency,
	}, nil
}

func (m Money) Amount() int64 {
	return m.amount
}

func (m Money) Currency() string {
	return m.currency
}

func (m Money) Equals(other Money) bool {
	return m.amount == other.amount && m.currency == other.currency
}

// Pricing represents order pricing information
type Pricing struct {
	unitPrice  Money
	totalPrice Money
}

func NewPricing(unitPrice Money, quantity int) (Pricing, error) {
	if quantity <= 0 {
		return Pricing{}, ErrInvalidQuantity
	}

	totalAmount := unitPrice.Amount() * int64(quantity)
	totalPrice, err := NewMoney(totalAmount, unitPrice.Currency())
	if err != nil {
		return Pricing{}, err
	}

	return Pricing{
		unitPrice:  unitPrice,
		totalPrice: totalPrice,
	}, nil
}

func (p Pricing) UnitPrice() Money {
	return p.unitPrice
}

func (p Pricing) TotalPrice() Money {
	return p.totalPrice
}

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPendingPayment OrderStatus = "PENDING_PAYMENT"
	OrderStatusPaid           OrderStatus = "PAID"
	OrderStatusCancelled      OrderStatus = "CANCELLED"
	OrderStatusExpired        OrderStatus = "EXPIRED"
)

func (s OrderStatus) String() string {
	return string(s)
}

func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusPendingPayment, OrderStatusPaid, OrderStatusCancelled, OrderStatusExpired:
		return true
	}
	return false
}

// PaymentID represents a payment identifier
type PaymentID struct {
	value string
}

func NewPaymentID() PaymentID {
	return PaymentID{value: uuidv7.New().String()}
}

func ParsePaymentID(id string) (PaymentID, error) {
	if id == "" {
		return PaymentID{}, ErrEmptyPaymentID
	}
	return PaymentID{value: id}, nil
}

func (id PaymentID) String() string {
	return id.value
}

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusCompleted PaymentStatus = "COMPLETED"
	PaymentStatusFailed    PaymentStatus = "FAILED"
)

func (s PaymentStatus) String() string {
	return string(s)
}

// PaymentMethod represents the payment method
type PaymentMethod string

const (
	PaymentMethodMock       PaymentMethod = "MOCK"
	PaymentMethodCreditCard PaymentMethod = "CREDIT_CARD"
)

func (m PaymentMethod) String() string {
	return string(m)
}
