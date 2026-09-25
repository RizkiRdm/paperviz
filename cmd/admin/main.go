// Command admin is PaperViz's account recovery tool.
//
// PaperViz has no self-service password reset, because delivering reset mail
// requires an email provider account, and holding no third-party account is a
// deliberate constraint of this product. Recovery is therefore an operator
// action against the database.
//
// Generated passwords are printed once to stdout and never logged or stored in
// the clear. Losing one is not a problem: run reset-password again.
package main

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"paperviz/internal/repository"
)

const passwordLength = 20

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

	db, closeDB, err := openDB()
	if err != nil {
		return err
	}
	defer closeDB()

	switch args[0] {
	case "create-user":
		return createUser(db, args[1:])
	case "reset-password":
		return resetPassword(db, args[1:])
	case "list-users":
		return listUsers(db)
	case "-h", "--help", "help":
		usage()
		return nil
	default:
		usage()
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}

// usage prints the available subcommands.
func usage() {
	fmt.Fprint(os.Stderr, `usage: admin <command> [args]

  create-user <email>     create an account and print its generated password
  reset-password <email>  replace a password and invalidate existing sessions
  list-users              list account emails

Reads DATABASE_PATH and MIGRATIONS_DIR from the environment.
`)
}

// openDB opens the database the same way the server does, so a reset here is
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

	migrations, err := repository.LoadMigrations(migrationsDir)
	if err != nil {
		return nil, nil, fmt.Errorf("load migrations: %w", err)
	}
	db, err := repository.Open(dbPath, migrations)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	return db, func() { db.Close() }, nil
}

// createUser inserts an account with a generated password.
func createUser(db *sql.DB, args []string) error {
	if len(args) != 1 {
		return errors.New("usage: create-user <email>")
	}
	email := normalizeEmail(args[0])

	users := repository.NewUserRepo(db)
	if _, err := users.GetByEmail(email); err == nil {
		return fmt.Errorf("account %s already exists", email)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("look up account: %w", err)
	}

	password, err := generatePassword()
	if err != nil {
		return err
	}
	hash, err := hashPassword(password)
	if err != nil {
		return err
	}
	id, err := repository.NewID()
	if err != nil {
		return fmt.Errorf("generate id: %w", err)
	}

	if err := users.Insert(repository.User{
		ID:           id,
		Email:        email,
		PasswordHash: hash,
		CreatedAt:    time.Now().Unix(),
	}); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	reportNewPassword(email, password)
	return nil
}

// resetPassword replaces a password and drops every existing session, so a
// reset actually revokes access rather than leaving a live session behind.
func resetPassword(db *sql.DB, args []string) error {
	if len(args) != 1 {
		return errors.New("usage: reset-password <email>")
	}
	email := normalizeEmail(args[0])

	users := repository.NewUserRepo(db)
	user, err := users.GetByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("no account for %s", email)
		}
		return fmt.Errorf("look up account: %w", err)
	}

	password, err := generatePassword()
	if err != nil {
		return err
	}
	hash, err := hashPassword(password)
	if err != nil {
		return err
	}

	if _, err := db.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, hash, user.ID); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if err := repository.NewSessionRepo(db).DeleteByUserID(user.ID); err != nil {
		// The password is already changed at this point, so failing the whole
		// command would be misleading. Report it loudly and carry on.
		slog.Error("session revocation failed after password reset",
			"user_id", user.ID, "error", err)
		fmt.Fprintln(os.Stderr,
			"warning: password changed but existing sessions could not be revoked")
		return nil
	}

	reportNewPassword(email, password)
	fmt.Println("Revoked all existing sessions for this account.")
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

// reportNewPassword prints a generated password exactly once.
func reportNewPassword(email, password string) {
	fmt.Printf("Account: %s\nPassword: %s\n\n", email, password)
	fmt.Fprintln(os.Stderr,
		"This password is shown once and cannot be retrieved. Store it now.")
}

// normalizeEmail lowercases and trims an email to match how signup stores it.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// hashPassword uses the same bcrypt cost as the signup handler, so an
// admin-created account is not weaker than a self-registered one.
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// generatePassword builds a random password that satisfies the same complexity
// rule the signup handler enforces, by drawing at least one character from each
// required class rather than rejecting and retrying.
func generatePassword() (string, error) {
	classes := []string{
		"ABCDEFGHJKLMNPQRSTUVWXYZ",  // upper, minus look-alikes
		"abcdefghijkmnopqrstuvwxyz", // lower, minus 'l'
		"23456789",                  // digits, minus 0 and 1
		"!@#$%^&*-_=+",              // special
	}

	out := make([]byte, 0, passwordLength)
	for _, class := range classes {
		c, err := randomFrom(class)
		if err != nil {
			return "", err
		}
		out = append(out, c)
	}
	all := strings.Join(classes, "")
	for len(out) < passwordLength {
		c, err := randomFrom(all)
		if err != nil {
			return "", err
		}
		out = append(out, c)
	}

	// Shuffle so the guaranteed characters are not always in the first
	// positions, which would make the shape predictable.
	for i := len(out) - 1; i > 0; i-- {
		j, err := randomInt(i + 1)
		if err != nil {
			return "", err
		}
		out[i], out[j] = out[j], out[i]
	}
	return string(out), nil
}

// randomFrom picks one character from set using a uniform random source.
func randomFrom(set string) (byte, error) {
	i, err := randomInt(len(set))
	if err != nil {
		return 0, err
	}
	return set[i], nil
}

// randomInt returns a uniform value in [0, n) using crypto/rand.
func randomInt(n int) (int, error) {
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0, fmt.Errorf("read random: %w", err)
	}
	return int(v.Int64()), nil
}
