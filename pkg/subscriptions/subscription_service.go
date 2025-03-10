package subscriptions

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/subscription"
	"github.com/stripe/stripe-go/v81/webhook"

	"gorm.io/gorm"
)


type SubscriptionService struct {
	db            *gorm.DB
	sessionID     string
	stripeSecret  string
	webhookSecret string
}

func NewSubscriptionService(db *gorm.DB, sesionID, stripeSecret, webhookSecret string) *SubscriptionService {
	return &SubscriptionService{
		db:            db,
		sessionID:     sesionID,
		stripeSecret:  stripeSecret,
		webhookSecret: webhookSecret,
	}

}

func (s *SubscriptionService) GetUserSubscription(userEmail string) (*Subscription, error) {
	var user User
	if err := s.db.Where("email = ?", userEmail).First(&user).Error; err != nil {
		return nil, err
	}

	var subscription Subscription
	if err := s.db.Where("user_id = ?", user.ID).First(&subscription).Error; err != nil {
		return nil, err
	}

	return &subscription, nil
}

func (s *SubscriptionService) CreateSubscription(userEmail, priceID string) (*stripe.CheckoutSession, error) {
	stripe.Key = s.stripeSecret

	var user User
	if err := s.db.Where("email = ?", userEmail).First(&user).Error; err != nil {
		return nil, err
	}

	var subscription Subscription
	if err := s.db.Where("user_id = ?", user.ID).First(&subscription).Error; err == nil {
		return nil, fmt.Errorf("User already has an active subscription")
	}

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		Mode:               stripe.String("subscription"),
		CustomerEmail:      stripe.String(userEmail),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String("https://app.drophunt.xyz/account-panel?tab=packages"),
		CancelURL:  stripe.String("https://app.drophunt.xyz/account-panel?tab=packages"),
	}

	sess, err := session.New(params)
	if err != nil {
		return nil, err
	}

	var typeID uint
	if err := s.db.Raw("SELECT id FROM subscription_types WHERE price_id = ?", priceID).Scan(&typeID).Error; err != nil {
		return nil, err
	}

	newSubscription := Subscription{
		SubscriptionID: sess.ID,
		UserID:         user.ID,
		StartDate:      time.Now().Format("2006-01-02"),
		EndDate:        time.Now().AddDate(0, 1, 0).Format("2006-01-02"),
		PaymentStatus:  "paymentPending",
		PaymentUrl:     sess.URL,
		TypeID:         typeID,
	}

	tx := s.db.Begin()

	if err := tx.Create(&newSubscription).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return sess, nil
}

func (s *SubscriptionService) UpdateUserSubscription(userEmail, priceID string) (map[string]string, error) {
	stripe.Key = s.stripeSecret

	var user User
	if err := s.db.Where("email = ?", userEmail).First(&user).Error; err != nil {
		return nil, err
	}

	var subs Subscription
	if err := s.db.Where("user_id = ?", user.ID).First(&subs).Error; err != nil {
		return nil, err
	}

	var currentTypeID uint
	if err := s.db.Raw("SELECT price FROM subscription_types WHERE type_id = ?", subs.TypeID).Scan(&currentTypeID).Error; err != nil {
		return nil, err
	}

	var newTypeID uint
	if err := s.db.Raw("SELECT price FROM subscription_types WHERE price_id = ?", priceID).Scan(&newTypeID).Error; err != nil {
		return nil, err
	}

	stripe.Key = s.stripeSecret

	if newTypeID > currentTypeID {
		sub, err := subscription.Get(subs.SubscriptionID, nil)
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve subscription details")
		}

		if len(sub.Items.Data) == 0 {
			return nil, fmt.Errorf("No subscription items found")
		}

		params := &stripe.CheckoutSessionParams{
			PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
			Mode:               stripe.String("subscription"),
			CustomerEmail:      stripe.String(userEmail),
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				{
					Price:    stripe.String(priceID),
					Quantity: stripe.Int64(1),
				},
			},
			SuccessURL: stripe.String("https://app.drophunt.xyz/account-panel?tab=packages"),
			CancelURL:  stripe.String("https://app.drophunt.xyz/account-panel?tab=packages"),
		}

		sess, err := session.New(params)
		if err != nil {
			return nil, err
		}

		tx := s.db.Begin()

		subs.TypeID = newTypeID
		subs.SubscriptionStatus = "upgradePending"

		if err := tx.Save(&subs).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		tx.Commit()

		return map[string]string{
			"status":       "upgrade pending",
			"checkout_url": sess.URL,
		}, nil

	} else if newTypeID < currentTypeID {
		tx := s.db.Begin()

		subs.TypeID = newTypeID
		subs.SubscriptionStatus = "downgraded"

		if err := tx.Save(&subs).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		sub, err := subscription.Get(subs.SubscriptionID, nil)
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve subscription details")
		}

		if len(sub.Items.Data) == 0 {
			return nil, fmt.Errorf("No subscription items found")
		}

		itemID := sub.Items.Data[0].ID

		params := &stripe.SubscriptionParams{
			Items: []*stripe.SubscriptionItemsParams{
				{
					ID:    stripe.String(itemID),
					Price: stripe.String(priceID),
				},
			},
		}

		if stripe.Key == "" {
			return nil, fmt.Errorf("No stripe key found")
		}

		result, err := subscription.Update(subs.SubscriptionID, params)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		fmt.Println("Subscription updated:", result.ID)

		tx.Commit()

		return map[string]string{
			"status": "downgrade success",
		}, nil
	} else {
		return map[string]string{
			"status": "no change",
		}, nil
	}
}

func (s *SubscriptionService) CancelUserSubscription(userEmail string) error {
	var user User
	if err := s.db.Where("email = ?", userEmail).First(&user).Error; err != nil {
		return err
	}

	var subs Subscription
	if err := s.db.Where("user_id = ?", user.ID).First(&subs).Error; err != nil {
		return err
	}

	tx := s.db.Begin()

	stripe.Key = s.stripeSecret

	if err := tx.Delete(&subs).Error; err != nil {
		tx.Rollback()
		return err
	}

	params := &stripe.SubscriptionCancelParams{}
	result, err := subscription.Cancel(subs.SubscriptionID, params)
	if err != nil {
		tx.Rollback()
		return err
	}

	fmt.Println("Subscription cancelled:", result.ID)

	tx.Commit()

	return nil
}

func (s *SubscriptionService) Webhook(payload []byte, sigHeader string) (*stripe.Event, error) {
	stripe.Key = s.stripeSecret

	event, err := webhook.ConstructEvent(payload, sigHeader, s.webhookSecret)
	if err != nil {
		return nil, err
	}

	var subscriptions []Subscription
	tx := s.db.Begin()

	switch event.Type {
	case "payment_intent.succeeded":
		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			return nil, err
		}
		fmt.Println("✅ PaymentIntent was successful! Payment Intent ID:", paymentIntent.ID)

	case "checkout.session.completed":
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			return nil, err
		}
		fmt.Println("✅ Checkout session completed! Session ID:", session.ID)

		if err := s.db.Where("subscription_id = ?", session.ID).Find(&subscriptions).Error; err != nil {
			return nil, err
		}

		for i := range subscriptions {
			subscriptions[i].PaymentStatus = "paymentSuccess"
		}

		s.sessionID = session.ID

		fmt.Println("Session ID:", s.sessionID)

		if err := tx.Save(&subscriptions).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

	case "customer.subscription.updated":
		var sub stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &sub)
		if err != nil {
			return nil, err
		}
		fmt.Println("✅ Subscription was updated! Subscription ID:", sub.ID)

		fmt.Println("Session ID:", s.sessionID)

		var subscriptions []Subscription
		if err := s.db.Where("subscription_id = ?", s.sessionID).Find(&subscriptions).Error; err != nil {
			fmt.Println("Error during query:", err)
			return nil, err
		}

		if len(subscriptions) == 0 {
			fmt.Println("No subscriptions found for ID:", sub.ID)
			return nil, fmt.Errorf("Subscription not found")
		}

		for i := range subscriptions {
			subscriptions[i].SubscriptionStatus = "active"
			subscriptions[i].CustomerID = sub.Customer.ID
			subscriptions[i].SubscriptionID = sub.ID
		}

		if err := tx.Save(&subscriptions).Error; err != nil {
			fmt.Println("Error saving updated subscriptions:", err)
			tx.Rollback()
			return nil, err
		}

		fmt.Println("Successfully updated subscription with ID:", sub.ID)
	}
	tx.Commit()
	return &event, nil
}

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
