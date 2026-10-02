package order

type OrderStatus string

const (
	StatusCreated       OrderStatus = "Criado"
	StatusStockOk       OrderStatus = "Estoque confirmado"
	StatusOutOfStock    OrderStatus = "Estoque indisponível"
	StatusPaymentOk     OrderStatus = "Pagamento aprovado"
	StatusPaymentFailed OrderStatus = "Pagamento recusado"
	StatusShipped       OrderStatus = "Enviado"
	StatusCancelled     OrderStatus = "Excluído"
)

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
