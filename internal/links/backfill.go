package links

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/refs"
)

// BackfillProject scans variables for ${{slug.VAR}} and creates missing outgoing links (idempotent).
func BackfillProject(ctx context.Context, q *sqlc.Queries, masterKey []byte, projectID string) (created int, err error) {
	services, err := q.ListServicesByProject(ctx, projectID)
	if err != nil {
		return 0, err
	}
	slugToID := make(map[string]string, len(services))
	for _, s := range services {
		slugToID[s.Slug] = s.ID
	}
	now := time.Now().Unix()
	for _, consumer := range services {
		if consumer.Kind != "app" {
			continue
		}
		rows, err := q.ListVariablesForService(ctx, consumer.ID)
		if err != nil {
			return created, err
		}
		linked := make(map[string]struct{})
		existing, err := q.ListLinkedProviderSlugs(ctx, consumer.ID)
		if err != nil {
			return created, err
		}
		for _, sl := range existing {
			linked[sl] = struct{}{}
		}
		for _, v := range rows {
			plain, err := crypto.Decrypt(masterKey, v.ValueEnc)
			if err != nil {
				return created, err
			}
			for _, refSlug := range refs.RefSlugsUnique(string(plain)) {
				if refSlug == consumer.Slug {
					continue
				}
				if _, ok := linked[refSlug]; ok {
					continue
				}
				providerID, ok := slugToID[refSlug]
				if !ok {
					continue
				}
				provider, err := q.GetServiceByID(ctx, providerID)
				if err != nil {
					return created, err
				}
				if consumer.EnvironmentID != provider.EnvironmentID {
					continue
				}
				if err := ValidatePair(consumer.Kind, provider.Kind); err != nil {
					continue
				}
				if _, err := q.GetServiceLinkByPair(ctx, sqlc.GetServiceLinkByPairParams{
					FromServiceID: consumer.ID,
					ToServiceID:   providerID,
				}); err == nil {
					linked[refSlug] = struct{}{}
					continue
				} else if !errorsIsNoRows(err) {
					return created, err
				}
				linkID := ulid.Make().String()
				if err := q.CreateServiceLink(ctx, sqlc.CreateServiceLinkParams{
					ID:            linkID,
					FromServiceID: consumer.ID,
					ToServiceID:   providerID,
					CreatedBy:     sql.NullString{},
					CreatedAt:     now,
				}); err != nil {
					return created, fmt.Errorf("create link %s→%s: %w", consumer.Slug, refSlug, err)
				}
				linked[refSlug] = struct{}{}
				created++
			}
		}
	}
	return created, nil
}

func errorsIsNoRows(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
