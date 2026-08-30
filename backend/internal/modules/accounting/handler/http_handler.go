package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/natekester/inventory-payment-integration/backend/internal/modules/accounting/domain"
)

type HTTPHandler struct {
	service *domain.Service
}

func NewHTTPHandler(service *domain.Service) *HTTPHandler {
	return &HTTPHandler{service: service}
}

func (h *HTTPHandler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/accounting")
	{
		group.POST("/entries", h.PostJournalEntry)
		group.GET("/entries", h.ListEntries)
		group.GET("/entries/:id", h.GetEntry)
	}
}

type postEntryRequest struct {
	ReferenceID string `json:"reference_id" binding:"required"`
	Memo        string `json:"memo" binding:"required"`
	AmountCents int64  `json:"amount_cents" binding:"required"`
	Currency    string `json:"currency" binding:"required"`
	Provider    string `json:"provider"` // Optional; defaults to 'rillet'
}

func (h *HTTPHandler) PostJournalEntry(c *gin.Context) {
	var req postEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	provider := req.Provider
	if provider == "" {
		provider = "rillet"
	}

	record := domain.SyncRecord{
		ReferenceID: req.ReferenceID,
		Memo:        req.Memo,
		AmountCents: req.AmountCents,
		Currency:    req.Currency,
	}

	entry, err := h.service.PostJournalEntry(c.Request.Context(), provider, record)
	if err != nil {
		if err == domain.ErrUnsupportedStrategy {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entry)
}

func (h *HTTPHandler) ListEntries(c *gin.Context) {
	entries, err := h.service.ListEntries(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"entries": entries})
}

func (h *HTTPHandler) GetEntry(c *gin.Context) {
	id := c.Param("id")
	entry, err := h.service.GetEntry(c.Request.Context(), id)
	if err != nil {
		if err == domain.ErrEntryNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entry)
}
