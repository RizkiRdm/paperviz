package users

import (
	"database/sql"
	"errors"
	"testing"

	"paperviz/internal/testutil"
)

// TestGetByID covers the read path AuthMiddleware and /api/auth/me share.
func TestGetByID(t *testing.T) {
	db := testutil.OpenTestDB(t)
	testutil.SeedClerkUser(t, db, "user_a", "a@example.com", 1700000000)
	svc := New(db)

	t.Run("existing subject resolves", func(t *testing.T) {
		u, err := svc.GetByID("user_a")
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if u.ID != "user_a" || u.Email != "a@example.com" {
			t.Fatalf("unexpected user: %+v", u)
		}
	})

	t.Run("unknown subject is sql.ErrNoRows", func(t *testing.T) {
		if _, err := svc.GetByID("ghost"); !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("err = %v, want sql.ErrNoRows", err)
		}
	})
}

// TestProvisionClerk covers first-sight provisioning and idempotency.
func TestProvisionClerk(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := New(db)

	tests := []struct {
		name  string
		email string
	}{
		{name: "first sight provisions", email: "a@example.com"},
		{name: "repeat sight upserts", email: "a@example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := svc.ProvisionClerk("user_a", tt.email, 1700000000); err != nil {
				t.Fatalf("provision: %v", err)
			}
		})
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE id = ?`, "user_a").Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("user count = %d, want 1", count)
	}
}
