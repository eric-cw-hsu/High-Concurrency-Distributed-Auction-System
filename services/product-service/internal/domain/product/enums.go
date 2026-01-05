package product

type ProductStatus string

const (
	ProductStatusDraft    ProductStatus = "DRAFT"
	ProductStatusActive   ProductStatus = "ACTIVE"
	ProductStatusInactive ProductStatus = "INACTIVE"
	ProductStatusSoldOut  ProductStatus = "SOLD_OUT"
)

func (s ProductStatus) IsValid() bool {
	switch s {
	case ProductStatusDraft, ProductStatusActive, ProductStatusInactive, ProductStatusSoldOut:
		return true
	default:
		return false
	}
}

func (s ProductStatus) CanPublish() bool {
	return s == ProductStatusDraft || s == ProductStatusInactive
}

func (s ProductStatus) CanDeactivate() bool {
	return s == ProductStatusActive
}

func (s ProductStatus) CanMarkAsSoldOut() bool {
	return s == ProductStatusActive
}

func (s ProductStatus) CanUpdate() bool {
	return s == ProductStatusDraft || s == ProductStatusInactive
}

// StockStatus represents the stock availability status (synced from Stock Service)
type StockStatus string

const (
	StockStatusUnknown    StockStatus = "UNKNOWN"      // not yet synced
	StockStatusInStock    StockStatus = "IN_STOCK"     // stock available
	StockStatusLowStock   StockStatus = "LOW_STOCK"    // stock below threshold
	StockStatusOutOfStock StockStatus = "OUT_OF_STOCK" // no stock
)

func (s StockStatus) IsValid() bool {
	switch s {
	case StockStatusUnknown, StockStatusInStock, StockStatusLowStock, StockStatusOutOfStock:
		return true
	default:
		return false
	}
}

// PriceType represents pricing strategy type (reserved for future auction support)
type PriceType string

const (
	PriceTypeFixed   PriceType = "FIXED"   // fixed price (flash sale)
	PriceTypeAuction PriceType = "AUCTION" // auction (future)
)

func (pt PriceType) IsValid() bool {
	switch pt {
	case PriceTypeFixed, PriceTypeAuction:
		return true
	default:
		return false
	}
}
