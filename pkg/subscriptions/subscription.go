package subscriptions

import (
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SubscriptionHandler struct {
	subscriptionService *SubscriptionService
}

func NewSubscriptionHandler(subscriptionService *SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
	}
}


func (h *SubscriptionHandler) GetUserSubscription(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subscription, err := h.subscriptionService.GetUserSubscription(req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No active subscription found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subscription": subscription})
}

func (h *SubscriptionHandler) AddSubscription(c *gin.Context) {
	var req struct {
		Email   string `json:"email" binding:"required"`
		PriceID string `json:"price_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.subscriptionService.CreateSubscription(req.Email, req.PriceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"checkout_url": session.URL})
}

func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
	var req struct {
		Email   string `json:"email" binding:"required"`
		PriceID string `json:"price_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.subscriptionService.UpdateUserSubscription(req.Email, req.PriceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *SubscriptionHandler) CancelSubscription(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.subscriptionService.CancelUserSubscription(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription cancelled"})
}

func (h *SubscriptionHandler) Webhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)

	payload, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Error reading request body"})
		return
	}

	sigHeader := c.GetHeader("Stripe-Signature")

	event, err := h.subscriptionService.Webhook(payload, sigHeader)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"event": event})
}
