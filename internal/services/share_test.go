package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"paperviz/internal/repository"
)

func shareStrPtr(s string) *string { return &s }

// insertShareFixture creates a user plus one complete document owned by that
// user, returning (db, docID, ownerID) ready for share-flow assertions.
func insertShareFixture(t *testing.T, visibility string) (*sql.DB, string, string) {
	t.Helper()
	db := openTestServicesDB(t)
	userRepo := repository.NewUserRepo(db)
	userID := "share-owner"
	if err := userRepo.Insert(repository.User{ID: userID, Email: "share@test.com", PasswordHash: "hash", CreatedAt: 1}); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	docRepo := repository.NewDocumentRepo(db)
	docID, err := repository.NewID()
	if err != nil {
		t.Fatalf("new id: %v", err)
	}
	doc := repository.Document{
		ID: docID, CreatedAt: 1000, LastAccessedAt: 1000,
		Status: repository.StatusComplete, SourceType: repository.SourceTypePDF,
		ReadingLevel: repository.ReadingLevelSimplified, Title: "Shared Paper",
		OriginalText: "secret original text", SimplifiedText: shareStrPtr("simple text"),
		UserID: &userID,
	}
	if visibility != "" {
		doc.Visibility = visibility
	}
	if err := docRepo.Insert(doc); err != nil {
		t.Fatalf("insert document: %v", err)
	}
	return db, docID, userID
}

// insertShareChart adds one chart to a document, optionally with an image blob.
func insertShareChart(t *testing.T, db *sql.DB, docID, chartID string, imageBlob []byte) {
	t.Helper()
	chartRepo := repository.NewChartRepo(db)
	if err := chartRepo.Insert(repository.Chart{
		ID: chartID, DocumentID: docID, SourceMethod: repository.ChartSourceImageFallback,
		ImageBlob: imageBlob,
	}); err != nil {
		t.Fatalf("insert chart %s: %v", chartID, err)
	}
}

func TestGenerateDocumentShareToken(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func(t *testing.T) (*sql.DB, string, string)
		userID  func(ownerID string) string
		wantErr string
		check   func(t *testing.T, db *sql.DB, token1, token2, docID string)
	}{
		{
			name:   "creates token and bumps private to unlisted",
			setup:  func(t *testing.T) (*sql.DB, string, string) { return insertShareFixture(t, "") },
			userID: func(owner string) string { return owner },
			check: func(t *testing.T, db *sql.DB, token1, token2, docID string) {
				if token1 == "" {
					t.Fatal("expected non-empty token")
				}
				d, err := repository.NewDocumentRepo(db).Get(docID)
				if err != nil {
					t.Fatalf("get doc: %v", err)
				}
				if d.Visibility != "unlisted" {
					t.Errorf("visibility: got %q, want unlisted", d.Visibility)
				}
				if d.ShareToken == nil || *d.ShareToken != token1 {
					t.Errorf("stored token mismatch: got %v", d.ShareToken)
				}
			},
		},
		{
			name:   "second call returns same token idempotently",
			setup:  func(t *testing.T) (*sql.DB, string, string) { return insertShareFixture(t, "") },
			userID: func(owner string) string { return owner },
			check: func(t *testing.T, db *sql.DB, token1, token2, docID string) {
				if token2 != token1 {
					t.Errorf("idempotency broken: got %q then %q", token1, token2)
				}
			},
		},
		{
			name:    "other user gets unauthorized",
			setup:   func(t *testing.T) (*sql.DB, string, string) { return insertShareFixture(t, "") },
			userID:  func(string) string { return "attacker" },
			wantErr: "unauthorized",
		},
		{
			name: "missing document errors",
			setup: func(t *testing.T) (*sql.DB, string, string) {
				db, _, owner := insertShareFixture(t, "")
				return db, "no-such-doc", owner
			},
			userID:  func(owner string) string { return owner },
			wantErr: "document not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, docID, owner := tt.setup(t)
			callerID := tt.userID(owner)

			token1, err := GenerateDocumentShareToken(ctx, db, docID, callerID)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("error: got %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("generate: %v", err)
			}
			token2, err := GenerateDocumentShareToken(ctx, db, docID, callerID)
			if err != nil {
				t.Fatalf("generate second: %v", err)
			}

			if tt.check != nil {
				tt.check(t, db, token1, token2, docID)
			}
		})
	}
}

func TestRevokeDocumentShareToken(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		public   bool
		userID   func(owner string) string
		wantErr  string
		checkVis string
	}{
		{
			name:     "clears token and reverts unlisted to private",
			userID:   func(owner string) string { return owner },
			checkVis: "private",
		},
		{
			name:     "public stays public after revoke",
			public:   true,
			userID:   func(owner string) string { return owner },
			checkVis: "public",
		},
		{
			name:    "other user unauthorized",
			userID:  func(string) string { return "attacker" },
			wantErr: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, docID, owner := insertShareFixture(t, "")

			token, err := GenerateDocumentShareToken(ctx, db, docID, owner)
			if err != nil {
				t.Fatalf("setup generate: %v", err)
			}
			if tt.public {
				if err := SetDocumentVisibility(ctx, db, docID, owner, "public"); err != nil {
					t.Fatalf("setup public: %v", err)
				}
			}

			err = RevokeDocumentShareToken(ctx, db, docID, tt.userID(owner))
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("error: got %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("revoke: %v", err)
			}

			_, err = GetSharedPaper(ctx, db, token)
			if err == nil || err.Error() != "not found" {
				t.Errorf("revoked token still resolvable: got %v, want not found", err)
			}

			d, err := repository.NewDocumentRepo(db).Get(docID)
			if err != nil {
				t.Fatalf("get doc: %v", err)
			}
			if d.ShareToken != nil {
				t.Errorf("share_token not cleared: got %v", *d.ShareToken)
			}
			if d.Visibility != tt.checkVis {
				t.Errorf("visibility after revoke: got %q, want %q", d.Visibility, tt.checkVis)
			}
		})
	}
}

func TestSetDocumentVisibility(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		visibility    string
		wantErr       string
		wantVis       string
		wantTokenKept bool
	}{
		{name: "to public keeps token", visibility: "public", wantVis: "public", wantTokenKept: true},
		{name: "to unlisted keeps token", visibility: "unlisted", wantVis: "unlisted", wantTokenKept: true},
		{name: "to private clears token", visibility: "private", wantVis: "private"},
		{name: "invalid value errors", visibility: "friends-only", wantErr: "invalid visibility"},
		{name: "empty value errors", visibility: "", wantErr: "invalid visibility"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, docID, owner := insertShareFixture(t, "")

			token, err := GenerateDocumentShareToken(ctx, db, docID, owner)
			if err != nil {
				t.Fatalf("setup generate: %v", err)
			}

			err = SetDocumentVisibility(ctx, db, docID, owner, tt.visibility)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("error: got %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("set visibility: %v", err)
			}

			d, err := repository.NewDocumentRepo(db).Get(docID)
			if err != nil {
				t.Fatalf("get doc: %v", err)
			}
			if d.Visibility != tt.wantVis {
				t.Errorf("visibility: got %q, want %q", d.Visibility, tt.wantVis)
			}
			tokenKept := d.ShareToken != nil && *d.ShareToken == token
			if tokenKept != tt.wantTokenKept {
				t.Errorf("token kept = %v, want %v", tokenKept, tt.wantTokenKept)
			}
		})
	}
}

func TestGetSharedPaper(t *testing.T) {
	ctx := context.Background()
	blob := []byte{0x89, 0x50, 0x4E, 0x47}

	t.Run("payload includes charts and image_url only when blob present", func(t *testing.T) {
		db, docID, owner := insertShareFixture(t, "")
		insertShareChart(t, db, docID, "chart-blob", blob)
		insertShareChart(t, db, docID, "chart-noblob", nil)

		token, err := GenerateDocumentShareToken(ctx, db, docID, owner)
		if err != nil {
			t.Fatalf("generate: %v", err)
		}

		paper, err := GetSharedPaper(ctx, db, token)
		if err != nil {
			t.Fatalf("get shared paper: %v", err)
		}
		if paper.DocumentID != docID || paper.Title != "Shared Paper" || paper.CreatedAt != 1000 {
			t.Errorf("paper header mismatch: %+v", paper)
		}
		if paper.SimplifiedText == nil || *paper.SimplifiedText != "simple text" {
			t.Errorf("simplified_text: got %v", paper.SimplifiedText)
		}
		if len(paper.Charts) != 2 {
			t.Fatalf("charts: got %d, want 2", len(paper.Charts))
		}
		for _, c := range paper.Charts {
			hasBlob := !strings.Contains(c.ChartID, "noblob")
			if hasBlob && c.ImageURL == "" {
				t.Errorf("chart %s missing image_url despite blob", c.ChartID)
			}
			if !hasBlob && c.ImageURL != "" {
				t.Errorf("chart %s has image_url without blob: %q", c.ChartID, c.ImageURL)
			}
		}
	})

	t.Run("marshaled JSON never leaks original text or tokens", func(t *testing.T) {
		db, docID, owner := insertShareFixture(t, "")
		insertShareChart(t, db, docID, "chart-1", blob)

		token, err := GenerateDocumentShareToken(ctx, db, docID, owner)
		if err != nil {
			t.Fatalf("generate: %v", err)
		}
		paper, err := GetSharedPaper(ctx, db, token)
		if err != nil {
			t.Fatalf("get shared paper: %v", err)
		}
		raw, err := json.Marshal(paper)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		for _, leak := range []string{"original_text", "user_id", "share_token", "OriginalText", "UserID"} {
			if strings.Contains(string(raw), leak) {
				t.Errorf("payload leaks %q: %s", leak, raw)
			}
		}
	})

	t.Run("unknown token is not found", func(t *testing.T) {
		db, _, _ := insertShareFixture(t, "")
		_, err := GetSharedPaper(ctx, db, "no-such-token")
		if err == nil || err.Error() != "not found" {
			t.Fatalf("error: got %v, want not found", err)
		}
	})

	t.Run("revoked token is not found", func(t *testing.T) {
		db, docID, owner := insertShareFixture(t, "")
		token, err := GenerateDocumentShareToken(ctx, db, docID, owner)
		if err != nil {
			t.Fatalf("generate: %v", err)
		}
		if err := RevokeDocumentShareToken(ctx, db, docID, owner); err != nil {
			t.Fatalf("revoke: %v", err)
		}
		_, err = GetSharedPaper(ctx, db, token)
		if err == nil || err.Error() != "not found" {
			t.Fatalf("error: got %v, want not found", err)
		}
	})
}
