package main

import (
	"drophunts/pkg/subscriptions"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=dev_password dbname=postgres port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(&subscriptions.User{}, &subscriptions.Subscription{}, &subscriptions.SubscriptionType{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	r := gin.Default()

	err = godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	stripeSecret := os.Getenv("STRIPE_SECRET_KEY")
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")

	subscriptionService := subscriptions.NewSubscriptionService(db, "", stripeSecret, webhookSecret)

	subscriptionHandler := subscriptions.NewSubscriptionHandler(subscriptionService)

	r.GET("/subscription", subscriptionHandler.GetUserSubscription)
	r.POST("/subscription", subscriptionHandler.AddSubscription)
	r.PUT("/subscription", subscriptionHandler.UpdateSubscription)
	r.DELETE("/subscription", subscriptionHandler.CancelSubscription)
	r.POST("/webhook", subscriptionHandler.Webhook)

	r.Run(":5555")
}

// type Subscription struct {
// 	SubscriptionID     string `gorm:"column:subscription_id;primaryKey"`
// 	UserID             uint   `gorm:"not null"`
// 	TypeID             uint   `gorm:"not null"`
// 	CustomerID         string
// 	StartDate          string `gorm:"type:date"`
// 	EndDate            string `gorm:"type:date"`
// 	SubscriptionStatus string
// 	PaymentStatus      string
// 	PaymentUrl         string
// 	DeletedAt          gorm.DeletedAt `gorm:"index"`
// }

// type SubscriptionType struct {
// 	TypeID    uint   `gorm:"primaryKey"`
// 	Name      string `gorm:"not null"`
// 	PriceID   string `gorm:"not null"`
// 	CreatedAt time.Time
// 	UpdatedAt time.Time
// 	DeletedAt gorm.DeletedAt `gorm:"index"`
// }

// type mainHandler struct {
// 	db            *gorm.DB
// 	sessionID     string
// 	stripeSecret  string
// 	webhookSecret string
// }

// func (h *mainHandler) GetUserSubscription(c *gin.Context) {
// 	var req struct {
// 		Email string `json:"email" binding:"required"`
// 	}

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	fmt.Println("Email: ", req.Email)

// 	var user User
// 	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
// 		return
// 	}

// 	var subscription Subscription
// 	if err := h.db.Where("user_id = ?", user.ID).First(&subscription).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "No active subscription"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"subscription": subscription})
// }

// func (h *mainHandler) AddSubscription(c *gin.Context) {
// 	var req struct {
// 		Email   string `json:"email" binding:"required"`
// 		PriceID string `json:"price_id" binding:"required"`
// 	}
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	stripe.Key = h.stripeSecret

// 	var user User
// 	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
// 		return
// 	}

// 	var subscription Subscription
// 	if err := h.db.Where("user_id = ?", user.ID).First(&subscription).Error; err == nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "User already has an active subscription"})
// 		return
// 	}

// 	params := &stripe.CheckoutSessionParams{
// 		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
// 		Mode:               stripe.String("subscription"),
// 		CustomerEmail:      stripe.String(req.Email),
// 		LineItems: []*stripe.CheckoutSessionLineItemParams{
// 			{
// 				Price:    stripe.String(req.PriceID),
// 				Quantity: stripe.Int64(1),
// 			},
// 		},
// 		SuccessURL: stripe.String("https://app.drophunt.xyz/account-panel?tab=packages"),
// 		CancelURL:  stripe.String("https://app.drophunt.xyz/account-panel?tab=packages"),
// 	}

// 	sess, err := session.New(params)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	var typeID uint

// 	if err := h.db.Raw("SELECT id FROM subscription_types WHERE price_id = ?", req.PriceID).Scan(&typeID).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription type"})
// 		return
// 	}

// 	newSubscription := Subscription{
// 		SubscriptionID: sess.ID,
// 		UserID:         user.ID,
// 		StartDate:      time.Now().Format("2006-01-02"),
// 		EndDate:        time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
// 		PaymentStatus:  "paymentPending",
// 		PaymentUrl:     sess.URL,
// 		TypeID:         typeID,
// 	}

// 	tx := h.db.Begin()

// 	if err := tx.Create(&newSubscription).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save subscription"})
// 		tx.Rollback()
// 		return
// 	}

// 	tx.Commit()

// 	c.JSON(http.StatusOK, gin.H{"checkout_url": sess.URL})
// }

// func (h *mainHandler) CancelSubscription(c *gin.Context) {
// 	var req struct {
// 		SubscriptionID string `json:"subscription_id" binding:"required"`
// 	}

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	stripe.Key = h.stripeSecret

// 	tx := h.db.Begin()

// 	if err := tx.Where("subscription_id = ?", req.SubscriptionID).Delete(&Subscription{}).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete subscription"})
// 		tx.Rollback()
// 		return
// 	}

// 	params := &stripe.SubscriptionCancelParams{}
// 	result, err := subscription.Cancel(req.SubscriptionID, params)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		tx.Rollback()
// 		return
// 	}

// 	tx.Commit()

// 	c.JSON(http.StatusOK, gin.H{"status": "subscription canceled", "subscription": result})
// }

// func (h *mainHandler) UpdateSubscription(c *gin.Context) {
// 	var req struct {
// 		SubscriptionID string `json:"subscription_id" binding:"required"`
// 		PriceID        string `json:"price_id" binding:"required"`
// 	}

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	tx := h.db.Begin()

// 	var type_id uint

// 	if err := tx.Raw("SELECT id FROM subscription_types WHERE price_id = ?", req.PriceID).Scan(&type_id).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get subscription type"})
// 		return
// 	}

// 	fmt.Println("Type ID:", type_id)

// 	if err := tx.Model(&Subscription{}).
// 		Where("subscription_id = ?", req.SubscriptionID).
// 		Update("type_id", type_id).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
// 		tx.Rollback()
// 		return
// 	}

// 	stripe.Key = h.stripeSecret

// 	sub, err := subscription.Get(req.SubscriptionID, nil)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve subscription details"})
// 		return
// 	}

// 	if len(sub.Items.Data) == 0 {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "No subscription items found"})
// 		return
// 	}

// 	itemID := sub.Items.Data[0].ID

// 	params := &stripe.SubscriptionParams{
// 		Items: []*stripe.SubscriptionItemsParams{
// 			{
// 				ID:    stripe.String(itemID),
// 				Price: stripe.String(req.PriceID),
// 			},
// 		},
// 	}

// 	if stripe.Key == "" {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Stripe API key is missing"})
// 		return
// 	}

// 	result, err := subscription.Update(req.SubscriptionID, params)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		tx.Rollback()
// 		return
// 	}

// 	tx.Commit()

// 	c.JSON(http.StatusOK, gin.H{"status": "subscription updated", "subscription": result})
// }

// func (h *mainHandler) Webhook(c *gin.Context) {
// 	stripe.Key = h.stripeSecret

// 	const MaxBodyBytes = int64(65536)
// 	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)

// 	payload, err := ioutil.ReadAll(c.Request.Body)
// 	if err != nil {
// 		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Error reading request body"})
// 		return
// 	}

// 	endpointSecret := h.webhookSecret
// 	sigHeader := c.GetHeader("Stripe-Signature")

// 	event, err := webhook.ConstructEvent(payload, sigHeader, endpointSecret)
// 	if err != nil {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
// 		return
// 	}

// 	var subscriptions []Subscription
// 	tx := h.db.Begin()

// 	switch event.Type {
// 	case "payment_intent.succeeded":
// 		var paymentIntent stripe.PaymentIntent
// 		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse payment_intent.succeeded"})
// 			return
// 		}
// 		fmt.Println("✅ PaymentIntent was successful! Payment Intent ID:", paymentIntent.ID)

// 	case "checkout.session.completed":
// 		var session stripe.CheckoutSession
// 		err := json.Unmarshal(event.Data.Raw, &session)
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse checkout.session.completed"})
// 			return
// 		}
// 		fmt.Println("✅ Checkout session completed! Session ID:", session.ID)

// 		if err := h.db.Where("subscription_id = ?", session.ID).Find(&subscriptions).Error; err != nil {
// 			c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
// 			return
// 		}

// 		for i := range subscriptions {
// 			subscriptions[i].PaymentStatus = "paymentSuccess"
// 		}

// 		h.sessionID = session.ID

// 		fmt.Println("Session ID:", h.sessionID)

// 		if err := tx.Save(&subscriptions).Error; err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
// 			tx.Rollback()
// 			return
// 		}

// 	case "customer.subscription.updated":
// 		var sub stripe.Subscription
// 		err := json.Unmarshal(event.Data.Raw, &sub)
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse customer.subscription.updated"})
// 			return
// 		}
// 		fmt.Println("✅ Subscription was updated! Subscription ID:", sub.ID)

// 		fmt.Println("Session ID:", h.sessionID)

// 		var subscriptions []Subscription
// 		if err := h.db.Where("subscription_id = ?", h.sessionID).Find(&subscriptions).Error; err != nil {
// 			fmt.Println("Error during query:", err)
// 			c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
// 			return
// 		}

// 		if len(subscriptions) == 0 {
// 			fmt.Println("No subscriptions found for ID:", sub.ID)
// 			c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
// 			return
// 		}

// 		for i := range subscriptions {
// 			subscriptions[i].SubscriptionStatus = "active"
// 			subscriptions[i].CustomerID = sub.Customer.ID
// 			subscriptions[i].SubscriptionID = sub.ID
// 		}

// 		if err := tx.Save(&subscriptions).Error; err != nil {
// 			fmt.Println("Error saving updated subscriptions:", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscription"})
// 			tx.Rollback()
// 			return
// 		}

// 		fmt.Println("Successfully updated subscription with ID:", sub.ID)

// 	default:
// 		fmt.Println("Unhandled event type:", event.Type)
// 	}

// 	tx.Commit()

// 	c.JSON(http.StatusOK, gin.H{"message": "Received"})
// }
