package product

import "errors"

// Domain layer error definitions
var (
	ErrProductNotFound                     = errors.New("product not found")
	ErrInvalidProductID                    = errors.New("invalid product id")
	ErrCannotPublishProduct                = errors.New("only draft products can be published")
	ErrCannotDeactivateProduct             = errors.New("only active products can be deactivated")
	ErrCannotMarkAsSoldOut                 = errors.New("only active products can be marked as sold out")
	ErrCannotUpdateActiveProduct           = errors.New("cannot update active product info")
	ErrCannotUpdatePricingForActiveProduct = errors.New("cannot update pricing for active product")
	ErrUnauthorizedDelete                  = errors.New("unauthorized to delete this product")
)
