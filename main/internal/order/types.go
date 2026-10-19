package orders

import (
	"time"
	"uuid"
)

type orderItemRequest struct {
	ProductID string `json:"productId" binding:"required,uuid"`
	Quantity  int32  `json:"quantity"   binding:"required,gt=0"`
}

type createOrderRequest struct {
	ClientID int64              `json:"clientId" binding:"required,gt=0"`
	Items    []orderItemRequest `json:"items"     binding:"required,min=1,dive"`
}

type listOrdersQuery struct {
	ClientID int64 `json:"clientId" binding:"required,gt=0"`
	Size     int32 `form:"size,default=10" binding:"min=1,max=100"`
	Page     int32 `form:"page,default=1" binding:"min=1"`
}

type OrderList struct {
	Items []OrderWithItems `json:"items"`
	Total int64            `json:"total"`
	Page  int32            `json:"page"`
}

type OrderItem struct {
	ProductID uuid.UUID `json:"productId"`
	Quantity  int32     `json:"quantity"`
}

type OrderWithItems struct {
	ID        uuid.UUID   `json:"id"`
	ClientID  int64       `json:"clientId"`
	CreatedAt time.Time   `json:"createdAt"`
	Status    string      `json:"status"`
	Items     []OrderItem `json:"items"`
}
