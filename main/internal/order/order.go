package order

type OrderItem struct {
	ProductId   string `json:"id"`
	ProductName string `json:"name"`
	Amount      int    `json:"amount"`
}

type Order struct {
	Id     string      `json:"id"`
	Items  []OrderItem `json:"items"`
	Status OrderStatus
}
