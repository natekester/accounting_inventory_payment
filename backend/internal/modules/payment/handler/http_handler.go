package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/natekester/inventory-payment-integration/backend/internal/modules/payment/domain"
)

type HTTPHandler struct {
	service *domain.Service
}

func NewHTTPHandler(service *domain.Service) *HTTPHandler {
	return &HTTPHandler{service: service}
}

func (h *HTTPHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/payment")
	{
		group.POST("/charge", h.ProcessPayment)
		group.GET("/transactions", h.ListTransactions)
		group.GET("/transactions/:id", h.GetTransaction)
	}
}

type processPaymentRequest struct {
	CustomerID  string `json:"customer_id" binding:"required"`
	AmountCents int64  `json:"amount_cents" binding:"required"`
	Currency    string `json:"currency" binding:"required"`
	Provider    string `json:"provider"` // Optional; defaults to 'stripe'
}

func (h *HTTPHandler) ProcessPayment(c *gin.Context) {
	var req processPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	provider := req.Provider
	if provider == "" {
		provider = "stripe"
	}

	payReq := domain.PaymentRequest{
		CustomerID:  req.CustomerID,
		AmountCents: req.AmountCents,
		Currency:    req.Currency,
	}

	tx, err := h.service.ProcessPayment(c.Request.Context(), provider, payReq)
	if err != nil {
		if err == domain.ErrUnsupportedStrategy {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tx)
}

func (h *HTTPHandler) ListTransactions(c *gin.Context) {
	txs, err := h.service.ListTransactions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transactions": txs})
}

func (h *HTTPHandler) GetTransaction(c *gin.Context) {
	id := c.Param("id")
	tx, err := h.service.GetTransaction(c.Request.Context(), id)
	if err != nil {
		if err == domain.ErrTransactionNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tx)
}
