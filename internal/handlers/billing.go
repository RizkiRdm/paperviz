package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"

	"paperviz/internal/services"
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
	if os.Getenv("STRIPE_SECRET_KEY") == "" {
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
	stripeCustomerID, err := services.EnsureStripeCustomer(r.Context(), h.db, userID)
	if err != nil {
		slog.Error("ensure stripe customer failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	sessID, sessURL, err := services.CreateCheckoutSession(r.Context(), stripeCustomerID, userID, req.PriceID)
	if err != nil {
		slog.Error("create checkout session failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, checkoutResponse{
		SessionID: sessID,
		URL:       sessURL,
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
	if os.Getenv("STRIPE_SECRET_KEY") == "" {
		writeError(w, http.StatusServiceUnavailable, "stripe_not_configured")
		return
	}
	portalURL, err := services.CreatePortalSession(r.Context(), h.db, userID)
	if err != nil {
		if err.Error() == "no_subscription" {
			writeError(w, http.StatusBadRequest, "no_subscription")
			return
		}
		slog.Error("create portal session failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}
	writeJSON(w, http.StatusOK, portalResponse{URL: portalURL})
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
	sigHeader := r.Header.Get("Stripe-Signature")
	if sigHeader == "" {
		writeError(w, http.StatusBadRequest, "missing_signature")
		return
	}
	var event webhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_event")
		return
	}
	switch event.Type {
	case "checkout.session.completed":
		customerID := event.Data.Object.Customer
		userID := event.Data.Object.Metadata["user_id"]
		if err := services.CompleteCheckout(r.Context(), h.db, userID, customerID); err != nil {
			slog.Error("complete checkout failed", "error", err)
		}
	case "customer.subscription.deleted":
		customerID := event.Data.Object.Customer
		if err := services.CancelSubscription(r.Context(), h.db, customerID); err != nil {
			slog.Error("cancel subscription failed", "error", err)
		}
	}
	writeJSON(w, http.StatusOK, nil)
}
