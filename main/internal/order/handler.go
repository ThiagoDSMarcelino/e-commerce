package order

import (
	"context"
	"main/internal/platform/rabbitmq"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	repo *OrdersRepository
	rmq  *rabbitmq.Broker
}

func NewOrderHandler(repo *OrdersRepository, rmq *rabbitmq.Broker) *OrderHandler {
	return &OrderHandler{repo: repo, rmq: rmq}
}

func (s *OrderHandler) GetOrders(c *gin.Context) {
	list := s.repo.GetAll()
	c.IndentedJSON(http.StatusOK, list)
}

// func createOrder(ctx context.Context, broker *Broker) error {
// 	var items []ProductRequest
//
// 	order := Order{
// 		Id:       orders.NextId(),
// 		Products: items,
// 	}
//
// 	orders.Add(order, StatusCreated)
//
// 	payload, err := order.Serialize()
// 	if err != nil {
// 		slog.Error("Failed to serialize order", "id", order.Id, "error", err)
// 		return err
// 	}
//
// 	if err := broker.Publish(ctx, routingKeyCreated, payload); err != nil {
// 		if !errors.Is(err, ErrUnroutable) {
// 			slog.Error("Failed to publish order", "id", order.Id, "error", err)
// 			return err
// 		}
//
// 		slog.Warn("Order event was not routed to any queue", "id", order.Id)
// 	}
//
// 	return nil
// }
//
// func deleteOrder(ctx context.Context, broker *Broker, readLine readFunc) (bool, error) {
// 	list := orders.List()
// 	if len(list) == 0 {
// 		fmt.Println("Nenhum pedido realizado")
// 		return false, nil
// 	}
//
// 	fmt.Print("\033[H\033[2J")
// 	fmt.Println("--------------------------------")
// 	print_orders(list)
// 	fmt.Println("--------------------------------")
// 	fmt.Println("Qual pedido que deseja excluir (vazio para cancelar):")
//
// 	input, quit, err := readLine()
// 	if quit {
// 		return true, err
// 	}
//
// 	input = strings.TrimSpace(input)
// 	if input == "" {
// 		return false, nil
// 	}
//
// 	index, convErr := strconv.Atoi(input)
// 	if convErr != nil {
// 		fmt.Println("Input inválido")
// 		return false, nil
// 	}
//
// 	if index < 1 || index > len(list) {
// 		fmt.Println("Ordem inválida")
// 		return false, nil
// 	}
//
// 	id := list[index-1].Order.Id
//
// 	order, ok := orders.MarkCancelled(id)
// 	if !ok {
// 		fmt.Println("Pedido", id, "já estava excluído")
// 		return false, nil
// 	}
//
// 	orders.SetStatus(id, StatusCancelled)
//
// 	fmt.Println("Pedido", order.Id, "excluído")
//
// 	payload, err := order.Serialize()
// 	if err != nil {
// 		slog.Error("Failed to serialize order", "id", order.Id, "error", err)
// 		fmt.Println("Não foi possível publicar a exclusão do pedido")
// 		return false, nil
// 	}
//
// 	if err := broker.Publish(ctx, routingKeyDeleted, payload); err != nil {
// 		if !errors.Is(err, ErrUnroutable) {
// 			slog.Error("Failed to publish deletion", "id", order.Id, "error", err)
// 			fmt.Println("Não foi possível publicar a exclusão do pedido")
// 			return false, nil
// 		}
//
// 		slog.Warn("Deletion event was not routed to any queue", "id", order.Id)
// 	}
//
// 	fmt.Println("Exclusão do pedido", order.Id, "publicada")
//
// 	return false, nil
// }

func (h *OrderHandler) HandleStatusEvent(ctx context.Context, event rabbitmq.Event) rabbitmq.MessageResponse {
	// 	order, err := ParseOrder(event.Data)
	// 	if err != nil {
	// 		slog.Error("Failed to parse order", "routingKey", event.RoutingKey, "error", err)
	// 		return Rejected
	// 	}
	//
	// 	status, ok := statusByRoutingKey[event.RoutingKey]
	// 	if !ok {
	// 		slog.Warn("Received event with unknown routing key",
	// 			"routingKey", event.RoutingKey, "id", order.Id)
	// 		return Rejected
	// 	}
	//
	// 	if !orders.SetStatus(order.Id, status) {
	// 		slog.Debug("Received event for an unknown order",
	// 			"id", order.Id, "routingKey", event.RoutingKey, "status", status)
	// 		return Accepted
	// 	}
	//
	// 	slog.Debug("Order status updated", "id", order.Id, "status", status)
	//
	// 	if !cancelsOrder[event.RoutingKey] {
	// 		return Accepted
	// 	}

	return rabbitmq.Accepted

	// return cancelOrder(ctx, broker, order.Id, status)
}

//
// func cancelOrder(ctx context.Context, broker *Broker, id string, reason OrderStatus) MessageResponse {
// 	order, first := orders.MarkCancelled(id)
// 	if !first {
// 		slog.Debug("Order was already cancelled", "id", id)
// 		return Accepted
// 	}
//
// 	payload, err := order.Serialize()
// 	if err != nil {
// 		slog.Error("Failed to serialize cancelled order", "id", id, "error", err)
// 		return Accepted
// 	}
//
// 	if err := broker.Publish(ctx, routingKeyDeleted, payload); err != nil {
// 		if !errors.Is(err, ErrUnroutable) {
// 			slog.Error("Failed to publish cancellation", "id", id, "error", err)
// 			return Accepted
// 		}
//
// 		slog.Warn("Cancellation event was not routed to any queue", "id", id)
// 	}
//
// 	slog.Info("Pedido excluído", "id", id, "motivo", reason)
//
// 	return Accepted
// }
