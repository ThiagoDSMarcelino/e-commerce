package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

const (
	routingKeyCreated = "pedido.criado"
	routingKeyDeleted = "pedido.excluido"
)

var bindingKeys = []string{
	"pagamento.aprovado",
	"pagamento.recusado",
	"pedido.enviado",
	"pedido.estoque_ok",
	"estoque.indisponivel",
}

type readFunc func() (line string, quit bool, err error)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// Create a context that will be canceled on SIGINT (Ctrl+C) or SIGTERM.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	settings, err := LoadSettings()
	if err != nil {
		return fmt.Errorf("Failed to load settings: %v", err)
	}

	SetupLogger(settings.LogLevel)

	signer, err := NewSigner(settings)
	if err != nil {
		return fmt.Errorf("Failed to create signer: %v", err)
	}

	broker, err := NewBroker(ctx, settings, signer, bindingKeys)
	if err != nil {
		return fmt.Errorf("Failed to create broker: %v", err)
	}
	defer func() {
		_ = broker.Close()
	}()

	consumeErrs := make(chan error, 1)
	go func() {
		consumeErrs <- broker.Consume(ctx, handleStatusEvent)
	}()

	lines := make(chan string)
	readErrs := make(chan error, 1)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				readErrs <- err
				return
			}
			lines <- line
		}
	}()

	var (
		quitting bool
		quitErr  error
	)

	readLine := func() (line string, quit bool, err error) {
		if quitting {
			return "", true, quitErr
		}

		quitting = true

		select {
		case <-ctx.Done():
			fmt.Println("\nEncerrando...")
			return "", true, nil

		case err := <-consumeErrs:
			if err != nil {
				quitErr = fmt.Errorf("consumer error: %w", err)
				return "", true, quitErr
			}
			return "", true, nil

		case err := <-readErrs:
			if errors.Is(err, io.EOF) {
				return "", true, nil
			}
			quitErr = fmt.Errorf("failed to read input: %w", err)
			return "", true, quitErr

		case line := <-lines:
			quitting = false
			return line, false, nil
		}
	}

	for {
		fmt.Print("\033[H\033[2J")
		fmt.Println("--------------------------------")
		fmt.Println("O que deseja fazer?")
		fmt.Println("1 - Visualizar produtos")
		fmt.Println("2 - Realizar pedido")
		fmt.Println("3 - Excluir pedido")
		fmt.Println("4 - Consultar pedidos realizados")
		fmt.Println("5 - Sair")
		fmt.Println("--------------------------------")

		input, quit, err := readLine()
		if quit {
			return err
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue // redraws the menu, handy right after a docker attach
		}

		option, convErr := strconv.Atoi(input)
		switch {
		case convErr != nil:
			fmt.Println("Input inválido")
		case option == 1:
			view_products()
		case option == 2:
			if quit, err := create_order(ctx, broker, readLine); quit {
				return err
			}
		case option == 3:
			if quit, err := delete_order(ctx, broker, readLine); quit {
				return err
			}
		case option == 4:
			view_orders()
		case option == 5:
			return nil
		default:
			fmt.Println("Opção inválida")
		}

		fmt.Println("\nPressione Enter para continuar...")
		if _, quit, err := readLine(); quit {
			return err
		}
	}
}

var products = []Product{}

func view_products() {
	if len(products) == 0 {
		load_products()
	}

	fmt.Print("\033[H\033[2J")
	fmt.Println("--------------------------------")
	fmt.Println("Produtos disponíveis:")
	for i, p := range products {
		fmt.Println(i+1, "-", p.Id, "| Nome:", p.Name, "| Preço:", p.Price)
	}
	fmt.Println("--------------------------------")
}

func load_products() {
	file, err := os.Open("products.json")
	if err != nil {
		slog.Error("No products.json found")
	}
	defer file.Close()

	data, err := io.ReadAll(file)

	var products_data Products
	err = json.Unmarshal(data, &products_data)
	if err != nil {
		slog.Error("Error parsing products.json")
	}

	products = products_data.Users
}

func create_order(ctx context.Context, broker *Broker, readLine readFunc) (bool, error) {
	if len(products) == 0 {
		load_products()
	}

	if len(products) == 0 {
		fmt.Println("Nenhum produto disponível")
		return false, nil
	}

	view_products()

	var items []ProductRequest

	for {
		fmt.Println("\nNúmero do produto (0 para finalizar o pedido):")

		input, quit, err := readLine()
		if quit {
			return true, err
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		index, convErr := strconv.Atoi(input)
		if convErr != nil {
			fmt.Println("Input inválido")
			continue
		}

		if index == 0 {
			break
		}

		if index < 1 || index > len(products) {
			fmt.Println("Produto inválido")
			continue
		}

		product := products[index-1]

		fmt.Printf("Quantidade de %s:\n", product.Name)

		input, quit, err = readLine()
		if quit {
			return true, err
		}

		amount, convErr := strconv.Atoi(strings.TrimSpace(input))
		if convErr != nil || amount <= 0 {
			fmt.Println("Quantidade inválida")
			continue
		}

		items = add_item(items, product, amount)

		fmt.Println("Adicionado:", amount, "x", product.Name)
	}

	if len(items) == 0 {
		fmt.Println("Pedido vazio, nenhum pedido foi criado")
		return false, nil
	}

	order := Order{
		Id:       orders.NextId(),
		Products: items,
	}

	orders.Add(order, StatusCreated)
	fmt.Println("Pedido", order.Id, "criado")

	payload, err := order.Serialize()
	if err != nil {
		slog.Error("Failed to serialize order", "id", order.Id, "error", err)
		fmt.Println("Não foi possível publicar o pedido")
		return false, nil
	}

	if err := broker.Publish(ctx, routingKeyCreated, payload); err != nil {
		if !errors.Is(err, ErrUnroutable) {
			slog.Error("Failed to publish order", "id", order.Id, "error", err)
			fmt.Println("Não foi possível publicar o pedido")
			return false, nil
		}

		slog.Warn("Order event was not routed to any queue", "id", order.Id)
	}

	fmt.Println("Pedido", order.Id, "publicado")

	return false, nil
}

func add_item(items []ProductRequest, product Product, amount int) []ProductRequest {
	for i := range items {
		if items[i].Id == product.Id {
			items[i].Amount += amount
			return items
		}
	}

	return append(items, ProductRequest{
		Id:     product.Id,
		Name:   product.Name,
		Amount: amount,
	})
}

func delete_order(ctx context.Context, broker *Broker, readLine readFunc) (bool, error) {
	list := orders.List()
	if len(list) == 0 {
		fmt.Println("Nenhum pedido realizado")
		return false, nil
	}

	fmt.Print("\033[H\033[2J")
	fmt.Println("--------------------------------")
	print_orders(list)
	fmt.Println("--------------------------------")
	fmt.Println("Id do pedido que deseja excluir (vazio para cancelar):")

	input, quit, err := readLine()
	if quit {
		return true, err
	}

	id := strings.TrimSpace(input)
	if id == "" {
		fmt.Println("Operação cancelada")
		return false, nil
	}

	order, ok := orders.Remove(id)
	if !ok {
		fmt.Println("Pedido", id, "não encontrado")
		return false, nil
	}

	fmt.Println("Pedido", order.Id, "excluído")

	payload, err := order.Serialize()
	if err != nil {
		slog.Error("Failed to serialize order", "id", order.Id, "error", err)
		fmt.Println("Não foi possível publicar a exclusão do pedido")
		return false, nil
	}

	if err := broker.Publish(ctx, routingKeyDeleted, payload); err != nil {
		if !errors.Is(err, ErrUnroutable) {
			slog.Error("Failed to publish deletion", "id", order.Id, "error", err)
			fmt.Println("Não foi possível publicar a exclusão do pedido")
			return false, nil
		}

		slog.Warn("Deletion event was not routed to any queue", "id", order.Id)
	}

	fmt.Println("Exclusão do pedido", order.Id, "publicada")

	return false, nil
}

func view_orders() {
	list := orders.List()
	if len(list) == 0 {
		fmt.Println("Nenhum pedido realizado")
		return
	}

	print_orders(list)
}

func print_orders(list []StoredOrder) {
	for i, stored := range list {
		fmt.Println(i+1, "-", stored.Order.Id, "| Status:", stored.Status)
		for _, p := range stored.Order.Products {
			fmt.Println("     ", p.Amount, "x", p.Id, "|", p.Name)
		}
	}
}
