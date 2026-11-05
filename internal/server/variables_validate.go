package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/refs"
)

func validateVariableRefs(ctx context.Context, q *sqlc.Queries, consumerID string, values []string) error {
	linked, err := q.ListLinkedProviderSlugs(ctx, consumerID)
	if err != nil {
		return err
	}
	allowed := make(map[string]struct{}, len(linked))
	for _, s := range linked {
		allowed[s] = struct{}{}
	}
	for _, val := range values {
		for _, slug := range refs.RefSlugsUnique(val) {
			if _, ok := allowed[slug]; !ok {
				return fmt.Errorf("service %q is not linked to this app; connect it first", slug)
			}
		}
	}
	return nil
}

func inferSecretFromKey(key string) bool {
	u := strings.ToUpper(key)
	return strings.Contains(u, "PASSWORD") || strings.Contains(u, "SECRET") || strings.HasSuffix(u, "_URL")
}
