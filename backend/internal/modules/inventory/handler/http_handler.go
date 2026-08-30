package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/natekester/inventory-payment-integration/backend/internal/modules/inventory/domain"
)

type HTTPHandler struct {
	service *domain.Service
}

func NewHTTPHandler(service *domain.Service) *HTTPHandler {
	return &HTTPHandler{service: service}
}

func (h *HTTPHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/inventory")
	{
		group.POST("/items", h.CreateItem)
		group.GET("/items", h.ListItems)
		group.GET("/items/:id", h.GetItem)
		group.POST("/items/:id/adjust", h.AdjustStock)
	}
}

type createItemRequest struct {
	SKU         string `json:"sku" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	PriceCents  int64  `json:"price_cents" binding:"required"`
}

func (h *HTTPHandler) CreateItem(c *gin.Context) {
	var req createItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item := &domain.Item{
		SKU:         req.SKU,
		Name:        req.Name,
		Description: req.Description,
		Quantity:    req.Quantity,
		PriceCents:  req.PriceCents,
	}

	if err := h.service.CreateItem(c.Request.Context(), item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *HTTPHandler) ListItems(c *gin.Context) {
	items, err := h.service.ListItems(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *HTTPHandler) GetItem(c *gin.Context) {
	id := c.Param("id")
	item, err := h.service.GetItem(c.Request.Context(), id)
	if err != nil {
		if err == domain.ErrItemNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

type adjustStockRequest struct {
	Delta int `json:"delta" binding:"required"`
}

func (h *HTTPHandler) AdjustStock(c *gin.Context) {
	id := c.Param("id")
	var req adjustStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.service.AdjustStock(c.Request.Context(), id, req.Delta)
	if err != nil {
		if err == domain.ErrItemNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrInsufficientStock {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}
