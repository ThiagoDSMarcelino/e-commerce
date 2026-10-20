package main

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type PaymentProcessor struct {
	broker  *Broker
	charges *ChargeClient
	delay   time.Duration
	mu      sync.Mutex
	pending map[string]*Order
}

func NewPaymentProcessor(broker *Broker, charges *ChargeClient, delay time.Duration) *PaymentProcessor {
	return &PaymentProcessor{
		broker:  broker,
		charges: charges,
		delay:   delay,
		pending: make(map[string]*Order),
	}
}

func (pp *PaymentProcessor) ProcessDelivery(ctx context.Context, data []byte) MessageResponse {
	order, err := ParseOrder(data)
	if err != nil {
		return Rejected
	}

	if err := waitDelay(ctx, pp.delay); err != nil {
		return Requeued
	}

	slog.Info("Iniciando pagamento", "order", order)

	checkoutUrl, err := pp.charges.CreateCharge(ctx, order.Id)
	if err != nil {
		slog.Error("Falha ao criar cobranca", "order", order.Id, "error", err)
		return Rejected
	}

	slog.Info("Cobranca criada", "order", order.Id, "checkoutUrl", checkoutUrl)

	pp.putPending(order)

	err = pp.publish(ctx, routingKeyPending, &Order{Id: order.Id, Products: order.Products, CheckoutUrl: checkoutUrl})
	if err != nil {
		slog.Error("Falha ao publicar "+routingKeyPending, "order", order.Id, "error", err)
		return Rejected
	}

	return Accepted
}

func (pp *PaymentProcessor) publish(ctx context.Context, routingKey string, order *Order) error {
	payload, err := order.Serialize()
	if err != nil {
		return err
	}

	return pp.broker.Publish(ctx, routingKey, payload)
}

func (pp *PaymentProcessor) putPending(order *Order) {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	pp.pending[order.Id] = order
}

func (pp *PaymentProcessor) takePending(orderId string) (*Order, bool) {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	order, ok := pp.pending[orderId]
	delete(pp.pending, orderId)

	return order, ok
}
