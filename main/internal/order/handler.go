package orders

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
