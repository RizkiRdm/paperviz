// Command admin is PaperViz's break-glass account tool.
//
// Clerk owns authentication, so PaperViz holds no password to reset and no
// session to revoke locally. What remains is operator access to the two things
// that still live on this machine: the local profile rows, and a way to cut off
// a compromised account when the Clerk dashboard is not reachable.
//
// Password recovery is Clerk's self-service flow, which is strictly better than
// the generated-password procedure this replaced: it needed no mail provider,
// and the user chooses their own secret.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	clerkSession "github.com/clerk/clerk-sdk-go/v2/session"
	clerkUser "github.com/clerk/clerk-sdk-go/v2/user"

	"paperviz/internal/repository"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run dispatches a subcommand. It returns an error rather than exiting so the
// open/close of the database is handled in one place.
func run(args []string) error {
	if len(args) == 0 {
		usage()
		return errors.New("a subcommand is required")
	}

	if args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		usage()
		return nil
	}

	// Only list-users needs the database. Session revocation talks to Clerk and
	// must keep working even when the local file is missing or migrating.
	if args[0] == "revoke-sessions" {
		return revokeSessions(args[1:])
	}

	db, closeDB, err := openDB()
	if err != nil {
		return err
	}
	defer closeDB()

	switch args[0] {
	case "list-users":
		return listUsers(db)
	default:
		usage()
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}

// usage prints the available subcommands.
func usage() {
	fmt.Fprint(os.Stderr, `usage: admin <command> [args]

  list-users                 list account emails
  revoke-sessions <email>    revoke every Clerk session for an account

Reads DATABASE_PATH and MIGRATIONS_DIR from the environment.
revoke-sessions additionally reads CLERK_SECRET_KEY.
`)
}

// openDB opens the database the same way the server does, so a change here is
// visible to a running server.
func openDB() (*sql.DB, func(), error) {
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "paperviz.db"
	}
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}

	db, err := repository.Open(dbPath, migrationsDir)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	return db, func() { db.Close() }, nil
}

// revokeSessions kills every live Clerk session for an account, which is the
// only way to lock someone out when Clerk's dashboard cannot be reached.
//
// It deliberately does not consult the local users table: the authoritative
// session list is Clerk's, and a local row can be missing while a valid Clerk
// session still exists.
func revokeSessions(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: revoke-sessions <email>")
	}
	secretKey := os.Getenv("CLERK_SECRET_KEY")
	if secretKey == "" {
		return errors.New("CLERK_SECRET_KEY environment variable is required")
	}
	clerk.SetKey(secretKey)

	ctx := context.Background()
	email := normalizeEmail(args[0])

	identifier, err := lookupClerkUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	list, err := clerkSession.List(ctx, &clerkSession.ListParams{UserID: &identifier})
	if err != nil {
		return fmt.Errorf("list clerk sessions: %w", err)
	}
	if list.TotalCount == 0 {
		fmt.Printf("No active sessions for %s.\n", email)
		return nil
	}

	revoked := 0
	for _, s := range list.Sessions {
		if _, err := clerkSession.Revoke(ctx, &clerkSession.RevokeParams{ID: s.ID}); err != nil {
			// Report which ones failed rather than abandoning the rest: a partial
			// revoke is still worth having, and a silent partial revoke is not.
			fmt.Fprintf(os.Stderr, "could not revoke session %s: %v\n", s.ID, err)
			continue
		}
		revoked++
	}
	fmt.Printf("Revoked %d of %d sessions for %s.\n", revoked, list.TotalCount, email)
	if revoked < int(list.TotalCount) {
		return fmt.Errorf("%d sessions remain active", int(list.TotalCount)-revoked)
	}
	return nil
}

// listUsers prints every account email and creation time.
func listUsers(db *sql.DB) error {
	rows, err := db.Query(`SELECT email, created_at FROM users ORDER BY created_at`)
	if err != nil {
		return fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var email string
		var createdAt int64
		if err := rows.Scan(&email, &createdAt); err != nil {
			return fmt.Errorf("scan user: %w", err)
		}
		fmt.Printf("%s\t%s\n", email, time.Unix(createdAt, 0).UTC().Format(time.RFC3339))
	}
	return rows.Err()
}

// normalizeEmail lowercases and trims an email so operator input matches the
// address Clerk stored.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// lookupClerkUserByEmail resolves an operator-typed email to a Clerk user id.
//
// An ambiguous match is an error rather than a best guess: revoking sessions
// for the wrong account is not recoverable from the terminal.
func lookupClerkUserByEmail(ctx context.Context, email string) (string, error) {
	list, err := clerkUser.List(ctx, &clerkUser.ListParams{EmailAddresses: []string{email}})
	if err != nil {
		return "", fmt.Errorf("look up clerk user: %w", err)
	}
	switch len(list.Users) {
	case 0:
		return "", fmt.Errorf("no Clerk account for %s", email)
	case 1:
		return list.Users[0].ID, nil
	default:
		return "", fmt.Errorf("%d Clerk accounts match %s; revoke them in the Clerk dashboard", len(list.Users), email)
	}
}
