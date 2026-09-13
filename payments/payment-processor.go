package main

import (
	"context"
	"log/slog"
	"math/rand/v2"
)

type PaymentProcessor struct {
	broker *Broker
}

func NewPaymentProcessor(broker *Broker) *PaymentProcessor {
	return &PaymentProcessor{
		broker: broker,
	}
}

func (pp *PaymentProcessor) ProcessDelivery(ctx context.Context, data []byte) MessageResponse {
	order, err := ParseOrder(data)
	if err != nil {
		return Rejected
	}

	shouldAprove, err := pp.processPayment(order)
	if err != nil {
		return Rejected
	}

	var routingKey string
	if shouldAprove {
		slog.Info("Pagamento aprovado", "order", order)
		routingKey = routingKeyApproved
	} else {
		slog.Info("Pagamento recusado", "order", order)
		routingKey = routingKeyDeclined
	}

	payload, err := order.Serialize()
	if err != nil {
		return Requeued
	}

	err = pp.broker.Publish(ctx, routingKey, payload)
	if err != nil {
		return Rejected
	}

	return Accepted
}

func (pp *PaymentProcessor) processPayment(order *Order) (bool, error) {
	slog.Info("Iniciando pagamento", "order", order)

	randomInt := rand.IntN(100)
	shouldAprove := randomInt < 80 // 80% chance of approval

	return shouldAprove, nil
}
