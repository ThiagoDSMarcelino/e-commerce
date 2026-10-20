package orders

type OrderStatus string

const (
	StatusCreated       OrderStatus = "Criado"
	StatusStockOk       OrderStatus = "Estoque confirmado"
	StatusAwaitingPay   OrderStatus = "Aguardando pagamento"
	StatusOutOfStock    OrderStatus = "Estoque indisponível"
	StatusPaymentOk     OrderStatus = "Pagamento aprovado"
	StatusPaymentFailed OrderStatus = "Pagamento recusado"
	StatusShipped       OrderStatus = "Enviado"
	StatusCancelled     OrderStatus = "Cancelado"
)

var nonCancellableStatuses = []string{
	string(StatusShipped),
	string(StatusCancelled),
	string(StatusOutOfStock),
	string(StatusPaymentFailed),
}

var statusByRoutingKey = map[string]OrderStatus{
	"pagamento.aprovado":   StatusPaymentOk,
	"pagamento.recusado":   StatusPaymentFailed,
	"pedido.enviado":       StatusShipped,
	"pedido.estoque_ok":    StatusStockOk,
	"estoque.indisponivel": StatusOutOfStock,
}

var cancelsOrder = map[string]bool{
	"pagamento.recusado":   true,
	"estoque.indisponivel": true,
}
