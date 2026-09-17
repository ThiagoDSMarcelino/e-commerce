package main

import (
	"context"
	"errors"
	"log/slog"
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

func handleStatusEvent(ctx context.Context, broker *Broker, event Event) MessageResponse {
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

	if !cancelsOrder[event.RoutingKey] {
		return Accepted
	}

	return cancelOrder(ctx, broker, order.Id, status)
}

func cancelOrder(ctx context.Context, broker *Broker, id string, reason OrderStatus) MessageResponse {
	order, first := orders.MarkCancelled(id)
	if !first {
		slog.Debug("Order was already cancelled", "id", id)
		return Accepted
	}

	payload, err := order.Serialize()
	if err != nil {
		slog.Error("Failed to serialize cancelled order", "id", id, "error", err)
		return Accepted
	}

	if err := broker.Publish(ctx, routingKeyDeleted, payload); err != nil {
		if !errors.Is(err, ErrUnroutable) {
			slog.Error("Failed to publish cancellation", "id", id, "error", err)
			return Accepted
		}

		slog.Warn("Cancellation event was not routed to any queue", "id", id)
	}

	slog.Info("Pedido excluído", "id", id, "motivo", reason)

	return Accepted
}
