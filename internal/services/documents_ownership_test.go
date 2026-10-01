package services

import (
	"testing"
	"time"

	"paperviz/internal/apperr"
	"paperviz/internal/repository"
)

// TestDocumentOwnership covers the cross-user IDOR class on documents: a user
// must not be able to rename, delete, flag-save, or read a document owned by
// someone else. A document uploaded while signed out has no owner and stays
// reachable by anyone holding its URL, which preserves the no-account flow.
func TestDocumentOwnership(t *testing.T) {
	db := openTestDB(t)
	userRepo := repository.NewUserRepo(db)
	docRepo := repository.NewDocumentRepo(db)

	owner := "doc-owner"
	other := "doc-other"
	for _, u := range []string{owner, other} {
		if err := userRepo.Insert(repository.User{ID: u, Email: u + "@test.com", PasswordHash: "hash", CreatedAt: time.Now().Unix()}); err != nil {
			t.Fatalf("insert user %s: %v", u, err)
		}
	}

	now := time.Now().Unix()
	ownerPtr := owner
	seedDoc := func(id string, userID *string) {
		t.Helper()
		err := docRepo.Insert(repository.Document{
			ID: id, CreatedAt: now, LastAccessedAt: now,
			Status: repository.StatusComplete, SourceType: repository.SourceTypePastedText,
			ReadingLevel: repository.ReadingLevelSimplified, OriginalText: "body", UserID: userID,
		})
		if err != nil {
			t.Fatalf("insert document %s: %v", id, err)
		}
	}
	seedDoc("doc-owned", &ownerPtr)
	seedDoc("doc-anon", nil)

	t.Run("owner may rename", func(t *testing.T) {
		if err := RenameDocument(db, "doc-owned", owner, "Mine"); err != nil {
			t.Fatalf("owner rename: %v", err)
		}
		got, err := docRepo.Get("doc-owned")
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if got.Title != "Mine" {
			t.Errorf("title = %q, want %q", got.Title, "Mine")
		}
	})

	t.Run("other user may not rename", func(t *testing.T) {
		err := RenameDocument(db, "doc-owned", other, "Pwned")
		if err == nil {
			t.Fatal("expected cross-user rename to be rejected")
		}
		if apperr.HTTPStatus(err) != 403 {
			t.Errorf("status = %d, want 403", apperr.HTTPStatus(err))
		}
		got, _ := docRepo.Get("doc-owned")
		if got.Title == "Pwned" {
			t.Fatal("cross-user rename mutated the document")
		}
	})

	t.Run("other user may not toggle saved", func(t *testing.T) {
		if err := ToggleDocumentSaved(db, "doc-owned", other, true); err == nil {
			t.Fatal("expected cross-user toggle-save to be rejected")
		}
		got, _ := docRepo.Get("doc-owned")
		if got.Saved {
			t.Fatal("cross-user toggle-save mutated the document")
		}
	})

	t.Run("other user may not delete", func(t *testing.T) {
		if err := DeleteDocument(db, "doc-owned", other); err == nil {
			t.Fatal("expected cross-user delete to be rejected")
		}
		if _, err := docRepo.Get("doc-owned"); err != nil {
			t.Fatal("cross-user delete removed the document")
		}
	})

	t.Run("other user may not read", func(t *testing.T) {
		if err := DocumentAccessAllowed(db, "doc-owned", other); err == nil {
			t.Fatal("expected cross-user read to be rejected")
		}
	})

	t.Run("owner may read", func(t *testing.T) {
		if err := DocumentAccessAllowed(db, "doc-owned", owner); err != nil {
			t.Fatalf("owner read rejected: %v", err)
		}
	})

	t.Run("anonymous document stays reachable", func(t *testing.T) {
		if err := DocumentAccessAllowed(db, "doc-anon", ""); err != nil {
			t.Fatalf("anonymous document should be readable, got: %v", err)
		}
		if err := DocumentAccessAllowed(db, "doc-anon", other); err != nil {
			t.Fatalf("anonymous document should be readable by any signed-in user, got: %v", err)
		}
	})

	t.Run("missing document reports not-found", func(t *testing.T) {
		err := DocumentAccessAllowed(db, "doc-missing", owner)
		if !IsDocumentNotFound(err) {
			t.Fatalf("err = %v, want repository.ErrNotFound", err)
		}
	})
}

// TestSharedDocumentReachableByToken proves the ownership rule does not break
// public sharing: share links resolve by token, not by owner session.
func TestSharedDocumentReachableByToken(t *testing.T) {
	db := openTestDB(t)
	docRepo := repository.NewDocumentRepo(db)
	now := time.Now().Unix()
	owner := "share-owner"
	if err := repository.NewUserRepo(db).Insert(repository.User{ID: owner, Email: owner + "@test.com", PasswordHash: "hash", CreatedAt: now}); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	token := "tok-123"
	err := docRepo.Insert(repository.Document{
		ID: "doc-shared", CreatedAt: now, LastAccessedAt: now,
		Status: repository.StatusComplete, SourceType: repository.SourceTypePastedText,
		ReadingLevel: repository.ReadingLevelSimplified, OriginalText: "body",
		UserID: &owner,
	})
	if err != nil {
		t.Fatalf("insert shared document: %v", err)
	}
	if err := docRepo.SetShareToken("doc-shared", token); err != nil {
		t.Fatalf("set share token: %v", err)
	}
	got, err := docRepo.GetByShareToken("tok-123")
	if err != nil {
		t.Fatalf("share token lookup failed: %v", err)
	}
	if got.ID != "doc-shared" {
		t.Errorf("ID = %q, want doc-shared", got.ID)
	}
}
