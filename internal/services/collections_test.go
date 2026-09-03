package services

import (
	"errors"
	"testing"
	"time"

	"paperviz/internal/repository"
)

// TestGetCollection table-drives success, forbidden, not-found paths.
func TestGetCollection(t *testing.T) {
	db := openTestDB(t)
	userRepo := repository.NewUserRepo(db)
	owner := "user-col-owner"
	other := "user-col-other"
	for _, u := range []string{owner, other} {
		if err := userRepo.Insert(repository.User{ID: u, Email: u + "@test.com", PasswordHash: "hash", CreatedAt: time.Now().Unix()}); err != nil {
			t.Fatalf("insert user: %v", err)
		}
	}
	if err := CreateCollection(db, "col-1", owner, "Owned", time.Now().Unix()); err != nil {
		t.Fatalf("create collection: %v", err)
	}
	tests := []struct {
		name    string
		id      string
		userID  string
		wantErr error
	}{
		{name: "owner success", id: "col-1", userID: owner, wantErr: nil},
		{name: "other forbidden", id: "col-1", userID: other, wantErr: ErrForbidden},
		{name: "missing not-found", id: "col-missing", userID: owner, wantErr: repository.ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCollection(db, tt.id, tt.userID)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("get: %v", err)
				}
				if got.ID != tt.id {
					t.Errorf("ID = %q, want %q", got.ID, tt.id)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// TestRenameCollection table-drives success, forbidden, not-found paths.
func TestRenameCollection(t *testing.T) {
	db := openTestDB(t)
	userRepo := repository.NewUserRepo(db)
	owner := "user-ren-owner"
	other := "user-ren-other"
	for _, u := range []string{owner, other} {
		if err := userRepo.Insert(repository.User{ID: u, Email: u + "@test.com", PasswordHash: "hash", CreatedAt: time.Now().Unix()}); err != nil {
			t.Fatalf("insert user: %v", err)
		}
	}
	if err := CreateCollection(db, "col-ren", owner, "Old", time.Now().Unix()); err != nil {
		t.Fatalf("create collection: %v", err)
	}
	tests := []struct {
		name    string
		id      string
		userID  string
		wantErr error
	}{
		{name: "owner success", id: "col-ren", userID: owner, wantErr: nil},
		{name: "other forbidden", id: "col-ren", userID: other, wantErr: ErrForbidden},
		{name: "missing not-found", id: "col-missing", userID: owner, wantErr: repository.ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RenameCollection(db, tt.id, tt.userID, "New")
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("rename: %v", err)
				}
				got, err := GetCollection(db, tt.id, tt.userID)
				if err != nil {
					t.Fatalf("verify get: %v", err)
				}
				if got.Name != "New" {
					t.Errorf("Name = %q, want New", got.Name)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// TestDeleteCollection table-drives success, forbidden, not-found paths.
func TestDeleteCollection(t *testing.T) {
	db := openTestDB(t)
	userRepo := repository.NewUserRepo(db)
	owner := "user-del-owner"
	other := "user-del-other"
	for _, u := range []string{owner, other} {
		if err := userRepo.Insert(repository.User{ID: u, Email: u + "@test.com", PasswordHash: "hash", CreatedAt: time.Now().Unix()}); err != nil {
			t.Fatalf("insert user: %v", err)
		}
	}
	if err := CreateCollection(db, "col-del", owner, "Gone", time.Now().Unix()); err != nil {
		t.Fatalf("create collection: %v", err)
	}
	if err := CreateCollection(db, "col-del-keep", owner, "Keep", time.Now().Unix()); err != nil {
		t.Fatalf("create collection: %v", err)
	}
	tests := []struct {
		name    string
		id      string
		userID  string
		wantErr error
	}{
		{name: "other forbidden", id: "col-del-keep", userID: other, wantErr: ErrForbidden},
		{name: "missing not-found", id: "col-missing", userID: owner, wantErr: repository.ErrNotFound},
		{name: "owner success", id: "col-del", userID: owner, wantErr: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DeleteCollection(db, tt.id, tt.userID)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("delete: %v", err)
				}
				if _, err := GetCollection(db, tt.id, tt.userID); !errors.Is(err, repository.ErrNotFound) {
					t.Fatalf("verify deleted: got %v, want ErrNotFound", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// TestCollectionDocumentsAuth verifies Add/Remove/List enforce ownership.
func TestCollectionDocumentsAuth(t *testing.T) {
	db := openTestDB(t)
	userRepo := repository.NewUserRepo(db)
	owner := "user-doc-owner"
	other := "user-doc-other"
	for _, u := range []string{owner, other} {
		if err := userRepo.Insert(repository.User{ID: u, Email: u + "@test.com", PasswordHash: "hash", CreatedAt: time.Now().Unix()}); err != nil {
			t.Fatalf("insert user: %v", err)
		}
	}
	docRepo := repository.NewDocumentRepo(db)
	if err := docRepo.Insert(repository.Document{ID: "doc-c1", CreatedAt: 1000, LastAccessedAt: 1000, Status: repository.StatusComplete, SourceType: repository.SourceTypePDF, ReadingLevel: repository.ReadingLevelSimplified, Title: "Doc", OriginalText: "orig"}); err != nil {
		t.Fatalf("insert doc: %v", err)
	}
	if err := CreateCollection(db, "col-docs", owner, "Docs", time.Now().Unix()); err != nil {
		t.Fatalf("create collection: %v", err)
	}
	if err := AddDocumentToCollection(db, "col-docs", other, "doc-c1"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("add forbidden: got %v, want ErrForbidden", err)
	}
	if err := AddDocumentToCollection(db, "col-docs", owner, "doc-c1"); err != nil {
		t.Fatalf("add success: %v", err)
	}
	if _, err := ListCollectionDocuments(db, "col-docs", other); !errors.Is(err, ErrForbidden) {
		t.Fatalf("list forbidden: got %v, want ErrForbidden", err)
	}
	docs, err := ListCollectionDocuments(db, "col-docs", owner)
	if err != nil {
		t.Fatalf("list success: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("got %d docs, want 1", len(docs))
	}
	if err := RemoveDocumentFromCollection(db, "col-docs", other, "doc-c1"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("remove forbidden: got %v, want ErrForbidden", err)
	}
	if err := RemoveDocumentFromCollection(db, "col-docs", owner, "doc-c1"); err != nil {
		t.Fatalf("remove success: %v", err)
	}
}
