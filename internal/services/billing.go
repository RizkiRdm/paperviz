package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/stripe/stripe-go/v79"
	stripeportal "github.com/stripe/stripe-go/v79/billingportal/session"
	"github.com/stripe/stripe-go/v79/checkout/session"
	"github.com/stripe/stripe-go/v79/customer"
)

// EnsureStripeCustomer returns existing or creates Stripe customer for user.

func EnsureStripeCustomer(ctx context.Context, db *sql.DB, userID string) (string, error) {
	var stripeCustomerID sql.NullString
	err := db.QueryRow(`SELECT stripe_customer_id FROM users WHERE id = ?`, userID).Scan(&stripeCustomerID)
	if err != nil && err != sql.ErrNoRows {
		return "", fmt.Errorf("get stripe id: %w", err)
	}
	if stripeCustomerID.Valid && stripeCustomerID.String != "" {
		return stripeCustomerID.String, nil
	}
	var email sql.NullString
	if err := db.QueryRow(`SELECT email FROM users WHERE id = ?`, userID).Scan(&email); err != nil {
		return "", fmt.Errorf("get email: %w", err)
	}
	if !email.Valid || email.String == "" {
		return "", fmt.Errorf("email not found")
	}
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	custParams := &stripe.CustomerParams{
		Email: stripe.String(email.String),
		Metadata: map[string]string{
			"user_id": userID,
		},
	}
	cust, err := customer.New(custParams)
	if err != nil {
		return "", fmt.Errorf("create stripe customer: %w", err)
	}
	if _, err := db.Exec(`UPDATE users SET stripe_customer_id = ? WHERE id = ?`, cust.ID, userID); err != nil {
		return "", fmt.Errorf("save stripe id: %w", err)
	}
	return cust.ID, nil
}

// CreateCheckoutSession creates Stripe checkout session for subscription.

func CreateCheckoutSession(ctx context.Context, customerID, userID, priceID string) (string, string, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
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
		return "", "", fmt.Errorf("create checkout session: %w", err)
	}
	return sess.ID, sess.URL, nil
}

// CreatePortalSession creates Stripe billing portal session for user.

func CreatePortalSession(ctx context.Context, db *sql.DB, userID string) (string, error) {
	var stripeCustomerID sql.NullString
	err := db.QueryRow(`SELECT stripe_customer_id FROM users WHERE id = ?`, userID).Scan(&stripeCustomerID)
	if err != nil || !stripeCustomerID.Valid || stripeCustomerID.String == "" {
		return "", fmt.Errorf("no_subscription")
	}
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(stripeCustomerID.String),
		ReturnURL: stripe.String(os.Getenv("FRONTEND_URL") + "/account"),
	}
	sess, err := stripeportal.New(params)
	if err != nil {
		return "", fmt.Errorf("create portal session: %w", err)
	}
	return sess.URL, nil
}

// CompleteCheckout marks user subscription as pro after checkout.

func CompleteCheckout(ctx context.Context, db *sql.DB, userID, customerID string) error {
	if userID == "" {
		return nil
	}
	if _, err := db.Exec(`UPDATE users SET subscription_tier = 'pro', subscription_status = 'active' WHERE id = ? AND stripe_customer_id = ?`, userID, customerID); err != nil {
		slog.Error("update subscription failed", "error", err)
		return fmt.Errorf("update subscription: %w", err)
	}
	return nil
}

// CancelSubscription marks subscription as canceled for customer.

func CancelSubscription(ctx context.Context, db *sql.DB, customerID string) error {
	if _, err := db.Exec(`UPDATE users SET subscription_tier = 'free', subscription_status = 'canceled' WHERE stripe_customer_id = ?`, customerID); err != nil {
		slog.Error("cancel subscription failed", "error", err)
		return fmt.Errorf("cancel subscription: %w", err)
	}
	return nil
}
