package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v79"
	stripeportal "github.com/stripe/stripe-go/v79/billingportal/session"
	"github.com/stripe/stripe-go/v79/checkout/session"
	"github.com/stripe/stripe-go/v79/customer"
)

// BillingHandler manages Stripe subscription billing.
type BillingHandler struct {
	db *sql.DB
}

// NewBillingHandler creates a new BillingHandler.
func NewBillingHandler(db *sql.DB) *BillingHandler {
	return &BillingHandler{db: db}
}

// checkoutRequest is the wire shape for POST /api/billing/checkout.
type checkoutRequest struct {
	PriceID string `json:"price_id"`
}

// checkoutResponse is the wire shape for checkout session response.
type checkoutResponse struct {
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
}

// CreateCheckoutSession handles POST /api/billing/checkout.
func (h *BillingHandler) CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	if stripe.Key == "" {
		writeError(w, http.StatusServiceUnavailable, "stripe_not_configured")
		return
	}

	var req checkoutRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request")
		return
	}

	if req.PriceID == "" {
		writeError(w, http.StatusBadRequest, "missing_price_id")
		return
	}

	// Get or create Stripe customer
	var stripeCustomerID string
	err := h.db.QueryRow(`SELECT stripe_customer_id FROM users WHERE id = ?`, userID).Scan(&stripeCustomerID)
	if err != nil && err != sql.ErrNoRows {
		slog.Error("get user stripe id failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	if stripeCustomerID == "" {
		// Get user email
		var email string
		err = h.db.QueryRow(`SELECT email FROM users WHERE id = ?`, userID).Scan(&email)
		if err != nil {
			slog.Error("get user email failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal_error")
			return
		}

		// Create Stripe customer
		custParams := &stripe.CustomerParams{
			Email: stripe.String(email),
			Metadata: map[string]string{
				"user_id": userID,
			},
		}
		cust, err := customer.New(custParams)
		if err != nil {
			slog.Error("create stripe customer failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal_error")
			return
		}
		stripeCustomerID = cust.ID

		// Save customer ID
		_, err = h.db.Exec(`UPDATE users SET stripe_customer_id = ? WHERE id = ?`, stripeCustomerID, userID)
		if err != nil {
			slog.Error("save stripe customer id failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal_error")
			return
		}
	}

	// Create checkout session
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(stripeCustomerID),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(req.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(os.Getenv("FRONTEND_URL") + "/account?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(os.Getenv("FRONTEND_URL") + "/account"),
		Metadata: map[string]string{
			"user_id": userID,
		},
	}

	sess, err := session.New(params)
	if err != nil {
		slog.Error("create checkout session failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, checkoutResponse{
		SessionID: sess.ID,
		URL:       sess.URL,
	})
}

// portalResponse is the wire shape for portal session response.
type portalResponse struct {
	URL string `json:"url"`
}

// CreatePortalSession handles POST /api/billing/portal.
func (h *BillingHandler) CreatePortalSession(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	if stripe.Key == "" {
		writeError(w, http.StatusServiceUnavailable, "stripe_not_configured")
		return
	}

	var stripeCustomerID string
	err := h.db.QueryRow(`SELECT stripe_customer_id FROM users WHERE id = ?`, userID).Scan(&stripeCustomerID)
	if err != nil || stripeCustomerID == "" {
		writeError(w, http.StatusBadRequest, "no_subscription")
		return
	}

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(stripeCustomerID),
		ReturnURL: stripe.String(os.Getenv("FRONTEND_URL") + "/account"),
	}

	sess, err := stripeportal.New(params)
	if err != nil {
		slog.Error("create portal session failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	writeJSON(w, http.StatusOK, portalResponse{URL: sess.URL})
}

// webhookEvent is the minimal structure for Stripe webhook events.
type webhookEvent struct {
	Type string `json:"type"`
	Data struct {
		Object struct {
			ID       string            `json:"id"`
			Customer string            `json:"customer"`
			Metadata map[string]string `json:"metadata"`
			Status   string            `json:"status"`
		} `json:"object"`
	} `json:"data"`
}

// HandleWebhook handles POST /api/billing/webhook.
func (h *BillingHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")

	if webhookSecret == "" {
		writeError(w, http.StatusServiceUnavailable, "stripe_not_configured")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body")
		return
	}

	// Verify webhook signature (simplified - in production use stripe.ConstructEvent)
	sigHeader := r.Header.Get("Stripe-Signature")
	if sigHeader == "" {
		writeError(w, http.StatusBadRequest, "missing_signature")
		return
	}

	// TODO: Implement proper signature verification with stripe.ConstructEvent
	// For now, parse the event directly
	var event webhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_event")
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		// Update user subscription
		customerID := event.Data.Object.Customer
		userID := event.Data.Object.Metadata["user_id"]

		if userID != "" {
			_, err := h.db.Exec(
				`UPDATE users SET subscription_tier = 'pro', subscription_status = 'active' WHERE id = ? AND stripe_customer_id = ?`,
				userID, customerID,
			)
			if err != nil {
				slog.Error("update subscription failed", "error", err)
			}
		}

	case "customer.subscription.deleted":
		customerID := event.Data.Object.Customer
		_, err := h.db.Exec(
			`UPDATE users SET subscription_tier = 'free', subscription_status = 'canceled' WHERE stripe_customer_id = ?`,
			customerID,
		)
		if err != nil {
			slog.Error("cancel subscription failed", "error", err)
		}
	}

	writeJSON(w, http.StatusOK, nil)
}
