package products

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

func (h *Handler) ListProducts(c *gin.Context) {
	list, err := h.srv.ListAvailable(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, list)
}
