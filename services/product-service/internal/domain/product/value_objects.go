package product

import (
	"errors"
)

type Money struct {
	amount   int64  // amount in cents to avoid floating point precision issues
	currency string // ISO 4217
}

func NewMoney(amount int64, currency string) (Money, error) {
	if amount < 0 {
		return Money{}, errors.New("amount cannot be negative")
	}
	if currency == "" {
		currency = "USD" // default in USD
	}
	if !isValidCurrency(currency) {
		return Money{}, errors.New("invalid currency code")
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

func (m Money) IsZero() bool {
	return m.amount == 0
}

func (m Money) GreaterThan(other Money) bool {
	if m.currency != other.currency {
		panic("cannot compare different currencies")
	}
	return m.amount > other.amount
}

func (m Money) Equals(other Money) bool {
	return m.amount == other.amount && m.currency == other.currency
}

func isValidCurrency(code string) bool {
	validCurrencies := map[string]bool{
		"USD": true,
		"TWD": true,
	}
	return validCurrencies[code]
}

// Pricing represents the pricing strategy of a product
type Pricing struct {
	regularPrice   Money
	flashSalePrice *Money // optional flash sale price
}

// NewPricing creates a new pricing with regular price
func NewPricing(regularPrice Money) (Pricing, error) {
	if regularPrice.IsZero() {
		return Pricing{}, errors.New("regular price must be greater than zero")
	}
	return Pricing{
		regularPrice: regularPrice,
	}, nil
}

// NewPricingWithFlashSale creates a new pricing with both regular and flash sale price
func NewPricingWithFlashSale(regularPrice Money, flashSalePrice Money) (Pricing, error) {
	if regularPrice.IsZero() {
		return Pricing{}, errors.New("regular price must be greater than zero")
	}
	if flashSalePrice.IsZero() {
		return Pricing{}, errors.New("flash sale price must be greater than zero")
	}
	if !regularPrice.GreaterThan(flashSalePrice) {
		return Pricing{}, errors.New("regular price must be greater than flash sale price")
	}
	return Pricing{
		regularPrice:   regularPrice,
		flashSalePrice: &flashSalePrice,
	}, nil
}

func (p Pricing) RegularPrice() Money {
	return p.regularPrice
}

func (p Pricing) FlashSalePrice() *Money {
	return p.flashSalePrice
}

func (p Pricing) HasFlashSale() bool {
	return p.flashSalePrice != nil
}

// CurrentPrice returns the effective current price
func (p Pricing) CurrentPrice() Money {
	if p.flashSalePrice != nil {
		return *p.flashSalePrice
	}
	return p.regularPrice
}
