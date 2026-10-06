package promotions

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	srv Service
}

func (h *Handler) ListRegistrations(c *gin.Context) {
	var query listRegistredEmailQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	list, err := h.srv.ListRegistredEmails(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	count, err := h.srv.Count(c.Request.Context(), query.ClientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	res := registrationListResponse{
		Items: list,
		Total: count,
		Page:  query.Size,
	}

	c.JSON(http.StatusOK, res)
}

func (h *Handler) RegisterSubscription(c *gin.Context) {
	var req registerSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.srv.RegisterSubscription(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) CancelSubscription(c *gin.Context) {
	var req cancelSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.srv.CancelSubscription(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
