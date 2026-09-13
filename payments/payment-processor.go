package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"

	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

type PaymentProcessor struct {
	broker *Broker
}

func NewPaymentProcessor(broker *Broker) *PaymentProcessor {
	return &PaymentProcessor{
		broker: broker,
	}
}

func (pp *PaymentProcessor) ProcessDelivery(ctx context.Context, delivery rmq.IDeliveryContext) error {
	message := delivery.Message()

	if len(message.Data) == 0 {
		return fmt.Errorf("Received message with no data")
	}

	if len(message.Data) > 1 {
		return fmt.Errorf("Received message with multiple data parts")
	}

	if len(message.Data[0]) == 0 {
		return fmt.Errorf("Received message with empty data")
	}

	order, err := ParseOrder(message.Data[0])
	if err != nil {
		return fmt.Errorf("Failed to parse order: %v", err)
	}

	shouldAprove, err := pp.processPayment(order)
	if err != nil {
		return fmt.Errorf("Failed to process payment: %v", err)
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
		return fmt.Errorf("Failed to serialize order: %v", err)
	}

	err = pp.broker.Publish(ctx, routingKey, payload)
	if err != nil {
		return fmt.Errorf("Failed to publish message: %v", err)
	}

	err = delivery.Accept(ctx)
	if err != nil {
		return fmt.Errorf("Failed to accept message: %v", err)
	}

	return nil
}

func (pp *PaymentProcessor) processPayment(order *Order) (bool, error) {
	slog.Info("Iniciando pagamento", "order", order)

	randomInt := rand.IntN(100)
	shouldAprove := randomInt < 80 // 80% chance of approval

	return shouldAprove, nil
}
