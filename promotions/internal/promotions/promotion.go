package promotions

import (
	"math/rand/v2"
	"promotions/internal/catalog"
	"time"

	"github.com/google/uuid"
)

const (
	minDiscount  = 5
	maxDiscount  = 30
	validityDays = 30
)

type Promotion struct {
	ID            string
	ProductID     string
	ProductName   string
	Category      string
	CategoryLabel string
	Discount      int
	ValidUntil    time.Time
}

func Generate(products []catalog.Product) Promotion {
	product := products[rand.IntN(len(products))]

	return Promotion{
		ID:            uuid.NewString(),
		ProductID:     product.ID,
		ProductName:   product.Name,
		Category:      product.Category.ID,
		CategoryLabel: product.Category.Label,
		Discount:      rand.IntN(maxDiscount-minDiscount+1) + minDiscount,
		ValidUntil:    time.Now().AddDate(0, 0, validityDays),
	}
}
