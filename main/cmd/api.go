package main

import (
	"context"
	"log/slog"
	repo "main/internal/adapters/postgresql/sqlc"
	"main/internal/config"
	orders "main/internal/order"
	"net/http"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Application struct {
	config *config.Config
	db     *pgx.Conn
}

func NewApplication(config *config.Config, conn *pgx.Conn) *Application {
	return &Application{config: config, db: conn}

}

func (app *Application) Mount() http.Handler {
	r := gin.New()

	r.SetTrustedProxies(nil)

	r.Use(requestid.New())
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(timeout(60 * time.Second))

	ordersService := orders.NewService(repo.New(app.db), app.db)

	ordersHandler := orders.NewHandler(ordersService)

	r.GET("/orders", ordersHandler.GetOrders)
	r.POST("/orders", ordersHandler.CreateOrder)

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
