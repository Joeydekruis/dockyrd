package preview

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/proxy"
)

// SyncOptions carries host-level settings for preview provisioning.
type SyncOptions struct {
	SecretKeyPath string
	PublicIPv4    string
}

// CopyVariables clones parent service variables onto a preview child.
func CopyVariables(ctx context.Context, q *sqlc.Queries, keyPath, parentID, childID string) error {
	key, err := crypto.LoadMasterKey(keyPath)
	if err != nil {
		return err
	}
	rows, err := q.ListVariablesForService(ctx, parentID)
	if err != nil {
		return err
	}
	for _, row := range rows {
		val := row.ValueEnc
		if row.IsSecret != 0 {
			plain, err := crypto.Decrypt(key, row.ValueEnc)
			if err != nil {
				return err
			}
			val, err = crypto.Encrypt(key, plain)
			if err != nil {
				return err
			}
		}
		if err := q.InsertVariable(ctx, sqlc.InsertVariableParams{
			ID:        ulid.Make().String(),
			ServiceID: childID,
			Key:       row.Key,
			ValueEnc:  val,
			IsSecret:  row.IsSecret,
			AtBuild:   row.AtBuild,
			AtRuntime: row.AtRuntime,
		}); err != nil {
			return err
		}
	}
	return nil
}

// EnsurePlatformDomain creates a primary platform hostname when the preview has none.
func EnsurePlatformDomain(ctx context.Context, q *sqlc.Queries, serviceID, slug, publicIPv4 string) error {
	existing, err := q.ListDomainsForService(ctx, serviceID)
	if err != nil {
		return err
	}
	for _, d := range existing {
		if d.IsPrimary != 0 {
			return nil
		}
	}
	publicIPv4 = strings.TrimSpace(publicIPv4)
	if publicIPv4 == "" {
		publicIPv4 = "127.0.0.1"
	}
	host, err := proxy.PlatformHostname(slug, publicIPv4)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	domainID := ulid.Make().String()
	return q.CreateDomain(ctx, sqlc.CreateDomainParams{
		ID:           domainID,
		ServiceID:    serviceID,
		Hostname:     host,
		IsPlatform:   1,
		IsPrimary:    1,
		DnsStatus:    sql.NullString{String: "active", Valid: true},
		DnsCheckedAt: sql.NullInt64{Int64: now, Valid: true},
		CertStatus:   sql.NullString{String: "active", Valid: true},
	})
}
