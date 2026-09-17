package main

import (
	"context"
	"log/slog"
)

var statusByRoutingKey = map[string]OrderStatus{
	"pagamento.aprovado":   StatusPaymentOk,
	"pagamento.recusado":   StatusPaymentFailed,
	"pedido.enviado":       StatusShipped,
	"pedido.estoque_ok":    StatusStockOk,
	"estoque.indisponivel": StatusOutOfStock,
}

func handleStatusEvent(_ context.Context, event Event) MessageResponse {
	order, err := ParseOrder(event.Data)
	if err != nil {
		slog.Error("Failed to parse order", "routingKey", event.RoutingKey, "error", err)
		return Rejected
	}

	status, ok := statusByRoutingKey[event.RoutingKey]
	if !ok {
		slog.Warn("Received event with unknown routing key",
			"routingKey", event.RoutingKey, "id", order.Id)
		return Rejected
	}

	if !orders.SetStatus(order.Id, status) {
		slog.Debug("Received event for an unknown order",
			"id", order.Id, "routingKey", event.RoutingKey, "status", status)
		return Accepted
	}

	slog.Debug("Order status updated", "id", order.Id, "status", status)

	return Accepted
}
