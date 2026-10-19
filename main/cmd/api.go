package main

import (
	"context"
	"log/slog"
	repo "main/internal/adapters/postgresql/sqlc"
	"main/internal/adapters/rabbitmq"
	"main/internal/config"
	orders "main/internal/order"
	"main/internal/products"
	"main/internal/promotions"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Application struct {
	config *config.Config
	db     *pgxpool.Pool
	rmq    *rabbitmq.Broker
}

func NewApplication(config *config.Config, conn *pgxpool.Pool, broker *rabbitmq.Broker) *Application {
	return &Application{config: config, db: conn, rmq: broker}
}

func (app *Application) Mount(ctx context.Context) http.Handler {
	r := gin.New()

	r.SetTrustedProxies(nil)

	r.Use(requestid.New())
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(timeout(60 * time.Second))

	ordersService := orders.NewService(repo.New(app.db), app.db, app.rmq)

	go func() {
		err := app.rmq.Consume(ctx, func(ctx context.Context, event rabbitmq.Event) rabbitmq.MessageResponse {
			return ordersService.HandleStatusEvent(ctx, event)
		})

		if err != nil {
			slog.Error("consumer error", "error", err)
			os.Exit(1)
		}
	}()

	ordersHandler := orders.NewHandler(ordersService)

	r.GET("/orders", ordersHandler.GetOrders)
	r.POST("/orders", ordersHandler.CreateOrder)
	r.POST("/orders/:id/cancel", ordersHandler.CancelOrder)

	promotionsHandler := promotions.NewHandler(promotions.NewService(repo.New(app.db), app.rmq))

	r.GET("/promotions", promotionsHandler.ListRegistrations)
	r.POST("/promotions", promotionsHandler.RegisterSubscription)
	r.DELETE("/promotions", promotionsHandler.CancelSubscription)

	productsHandler := products.NewHandler(products.NewService(app.config.StorageURL))

	r.GET("/products", productsHandler.ListProducts)

	return r
}

func (app *Application) Run(h http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.Addr,
		Handler:      h,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  time.Minute,
	}

	slog.Debug("server has started", "addr", app.config.Addr)

	return srv.ListenAndServe()
}

func timeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), d)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		if ctx.Err() == context.DeadlineExceeded && !c.Writer.Written() {
			c.AbortWithStatus(http.StatusGatewayTimeout)
		}
	}
}
