package redis

import (
	"fmt"

	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/domain/reservation"
	"github.com/eric-cw-hsu/high-concurrency-distributed-auction-system/stock-service/internal/domain/stock"
)

// stockKey generates Redis key for stock
func stockKey(productID stock.ProductID) string {
	return fmt.Sprintf("stock:product:%s", productID.String())
}

func reservationKey(id reservation.ReservationID) string {
	return fmt.Sprintf("reservation:%s", id.String())
}

// stockMetadataKey generates Redis key for stock metadata
func stockMetadataKey(productID stock.ProductID) string {
	return fmt.Sprintf("stock:product:%s:meta", productID.String())
}
