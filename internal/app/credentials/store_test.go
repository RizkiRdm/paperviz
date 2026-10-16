package credentials

import (
	"database/sql"
	"errors"
	"testing"

	"paperviz/internal/external"
	"paperviz/internal/testutil"
)

const storeTestKey = "sk-test-SUPER-SECRET-1234"

// newTestStore builds a Store on a migrated database with two users.
func newTestStore(t *testing.T) (*Store, *sql.DB) {
	t.Helper()
	db := testutil.OpenTestDB(t)
	testutil.SeedUser(t, db, "u1")
	testutil.SeedUser(t, db, "u2")
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	cipher, err := external.NewCipher(key)
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}
	return NewStore(db, cipher), db
}

// TestStoreCreateDefaultsAndScope covers the first-key-becomes-default rule
// and that every read stays scoped to one user.
func TestStoreCreateDefaultsAndScope(t *testing.T) {
	store, _ := newTestStore(t)

	tests := []struct {
		name          string
		userID        string
		provider      string
		wantDefault   bool
		wantListCount int
	}{
		{name: "first key becomes default", userID: "u1", provider: "gemini", wantDefault: true, wantListCount: 1},
		{name: "second key is not default", userID: "u1", provider: "openai", wantDefault: false, wantListCount: 2},
		{name: "other user's keys are separate", userID: "u2", provider: "anthropic", wantDefault: true, wantListCount: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred, err := store.Create(tt.userID, tt.provider, "m1", storeTestKey)
			if err != nil {
				t.Fatalf("create: %v", err)
			}
			if cred.IsDefault != tt.wantDefault {
				t.Fatalf("IsDefault = %v, want %v", cred.IsDefault, tt.wantDefault)
			}
			if cred.KeyHint == "" || cred.KeyHint == storeTestKey {
				t.Fatalf("bad key hint %q", cred.KeyHint)
			}
			list, err := store.List(tt.userID)
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			if len(list) != tt.wantListCount {
				t.Fatalf("list len = %d, want %d", len(list), tt.wantListCount)
			}
		})
	}
}

// TestStoreErrorPaths covers the missing-row answers callers map to 404.
func TestStoreErrorPaths(t *testing.T) {
	store, _ := newTestStore(t)
	if _, err := store.Create("u1", "gemini", "m1", storeTestKey); err != nil {
		t.Fatalf("seed credential: %v", err)
	}

	tests := []struct {
		name    string
		call    func() error
		wantErr error
	}{
		{
			name: "set default of unknown id",
			call: func() error {
				return store.SetDefault("u1", "ghost")
			},
			wantErr: ErrNotFound,
		},
		{
			name: "get default without any key",
			call: func() error {
				_, err := store.GetDefault("u2")
				return err
			},
			wantErr: sql.ErrNoRows,
		},
		{
			name: "cross-user get by id is invisible",
			call: func() error {
				cred, err := store.List("u1")
				if err != nil || len(cred) == 0 {
					t.Fatalf("seed list: %v", err)
				}
				_, err = store.GetByID("u2", cred[0].ID)
				return err
			},
			wantErr: sql.ErrNoRows,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
