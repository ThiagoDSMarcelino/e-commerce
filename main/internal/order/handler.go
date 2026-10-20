package orders

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"uuid"

	"github.com/gin-gonic/gin"
)

const (
	sseEventName = "order-status"
	sseKeepAlive = 15 * time.Second
)

type Handler struct {
	srv Service
}

func NewHandler(srv Service) *Handler {
	return &Handler{srv: srv}
}

func (h *Handler) GetOrders(c *gin.Context) {
	var query listOrdersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	list, err := h.srv.ListOrders(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	count, err := h.srv.Count(c.Request.Context(), query.ClientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	res := OrderList{
		Items: list,
		Total: count,
		Page:  query.Page,
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.srv.PlaceOrder(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *Handler) CancelOrder(c *gin.Context) {
	var uri cancelOrderUri
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orderId, err := uuid.Parse(uri.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req cancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.srv.CancelOrder(c.Request.Context(), orderId, req.ClientID)
	switch {
	case errors.Is(err, ErrOrderNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	case errors.Is(err, ErrOrderNotCancellable):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) StreamEvents(c *gin.Context) {
	var query orderEventsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rc := http.NewResponseController(c.Writer)
	if err := rc.SetWriteDeadline(time.Time{}); err != nil {
		slog.Warn("Failed to clear write deadline for SSE", "error", err)
	}

	events, unsubscribe := h.srv.Subscribe(query.ClientID)
	defer unsubscribe()

	header := c.Writer.Header()
	header.Set("Content-Type", "text/event-stream")
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	header.Set("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	if _, err := fmt.Fprint(c.Writer, ": ping\n\n"); err != nil {
		return
	}
	c.Writer.Flush()

	ticker := time.NewTicker(sseKeepAlive)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			if _, err := fmt.Fprint(c.Writer, ": ping\n\n"); err != nil {
				return
			}
		case event, ok := <-events:
			if !ok {
				return
			}
			data, err := json.Marshal(event)
			if err != nil {
				slog.Error("Failed to marshal order status event", "order_id", event.OrderID, "error", err)
				continue
			}
			if _, err := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", sseEventName, data); err != nil {
				return
			}
		}
		c.Writer.Flush()
	}
}
