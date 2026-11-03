package admin

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/auth"
	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
)

// ResetPassword sets a new password for the user (server-side emergency).
func ResetPassword(cfg config.Config, email, password string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("email required")
	}
	if len(password) < 12 {
		return errors.New("password must be at least 12 characters")
	}
	conn, q, err := db.Open(cfg.DatabasePath())
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx := context.Background()
	user, err := q.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("user lookup: %w", err)
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	if err := q.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		PasswordHash: sql.NullString{String: hash, Valid: true},
		UpdatedAt:    now,
		ID:           user.ID,
	}); err != nil {
		return err
	}
	return nil
}

// Disable2FA clears TOTP for the user.
func Disable2FA(cfg config.Config, email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("email required")
	}
	conn, q, err := db.Open(cfg.DatabasePath())
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx := context.Background()
	user, err := q.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("user lookup: %w", err)
	}
	return q.DisableUserTOTP(ctx, sqlc.DisableUserTOTPParams{
		UpdatedAt: time.Now().Unix(),
		ID:        user.ID,
	})
}

// ReadPasswordLine reads one line from stdin (for CLI use when no arg given).
func ReadPasswordLine(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	sc := bufio.NewScanner(os.Stdin)
	if !sc.Scan() {
		return "", sc.Err()
	}
	return strings.TrimSpace(sc.Text()), nil
}
