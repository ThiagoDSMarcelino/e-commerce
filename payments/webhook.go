package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

const (
	statusApproved = "APROVADO"
	statusDeclined = "RECUSADO"
)

type webhookRequest struct {
	OrderId string `json:"orderId"`
	Status  string `json:"status"`
}

func (pp *PaymentProcessor) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var req webhookRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	var routingKey string
	switch req.Status {
	case statusApproved:
		routingKey = routingKeyApproved
	case statusDeclined:
		routingKey = routingKeyDeclined
	default:
		http.Error(w, "status must be APROVADO or RECUSADO", http.StatusBadRequest)
		return
	}

	order, ok := pp.takePending(req.OrderId)
	if !ok {
		slog.Warn("Webhook para pedido sem cobranca pendente", "order", req.OrderId, "status", req.Status)
		http.Error(w, "no pending charge for this order", http.StatusNotFound)
		return
	}

	if err := pp.publish(r.Context(), routingKey, order); err != nil {
		slog.Error("Falha ao publicar "+routingKey, "order", order.Id, "error", err)
		pp.putPending(order)
		http.Error(w, "failed to publish payment result", http.StatusServiceUnavailable)
		return
	}

	slog.Info("Pagamento "+req.Status, "order", order.Id)

	w.WriteHeader(http.StatusNoContent)
}
