package deploy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Joeydekruis/dockyrd/internal/appconfig"
	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/dbkind"
	"github.com/Joeydekruis/dockyrd/internal/refs"
	"github.com/Joeydekruis/dockyrd/internal/serviceops"
)

// projectCatalog builds published variables for every service in a project (briefing §7.3).
func (rt *Runtime) projectCatalog(ctx context.Context, projectID string) (map[string]map[string]string, error) {
	key, err := crypto.LoadMasterKey(rt.Cfg.SecretKeyPath())
	if err != nil {
		return nil, err
	}
	rows, err := rt.Q.ListServicesByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	out := make(map[string]map[string]string, len(rows))
	for _, row := range rows {
		raw := appconfig.RawConfig(row.Config)
		vars, err := publishedForService(ctx, rt.Q, key, serviceops.ServiceFromListRow(row), raw)
		if err != nil {
			return nil, fmt.Errorf("service %s: %w", row.Slug, err)
		}
		if len(vars) > 0 {
			out[row.Slug] = vars
		}
	}
	return out, nil
}

func publishedForService(ctx context.Context, q *sqlc.Queries, masterKey []byte, row sqlc.Service, raw json.RawMessage) (map[string]string, error) {
	switch row.Kind {
	case "postgres":
		pg, pw, err := dbkind.ParsePostgres(raw, masterKey)
		if err != nil {
			return nil, err
		}
		return dbkind.Published("postgres", row.Slug, pw, pg), nil
	case "redis":
		_, pw, err := dbkind.ParseRedis(raw, masterKey)
		if err != nil {
			return nil, err
		}
		return dbkind.Published("redis", row.Slug, pw, dbkind.PostgresConfig{}), nil
	case "app":
		cfg, err := appconfig.Parse(raw)
		if err != nil {
			return nil, err
		}
		if cfg.Port <= 0 {
			cfg.Port = 80
		}
		vars := map[string]string{
			"PRIVATE_HOST": fmt.Sprintf("%s:%d", row.Slug, cfg.Port),
		}
		if q != nil {
			domains, err := q.ListDomainsForService(ctx, row.ID)
			if err != nil {
				return nil, err
			}
			for _, d := range domains {
				if d.IsPrimary != 0 {
					vars["PUBLIC_URL"] = "https://" + d.Hostname
					break
				}
			}
			if vars["PUBLIC_URL"] == "" && len(domains) > 0 {
				vars["PUBLIC_URL"] = "https://" + domains[0].Hostname
			}
		}
		return vars, nil
	case "bucket":
		return publishedBucket(raw, masterKey, row.Slug)
	default:
		return nil, nil
	}
}

// consumerCatalog builds published variables only for services linked to the consumer.
func (rt *Runtime) consumerCatalog(ctx context.Context, consumerServiceID, projectID string) (refs.Catalog, error) {
	full, err := rt.projectCatalog(ctx, projectID)
	if err != nil {
		return nil, err
	}
	slugs, err := rt.Q.ListLinkedProviderSlugs(ctx, consumerServiceID)
	if err != nil {
		return nil, err
	}
	allowed := make(map[string]struct{}, len(slugs))
	for _, s := range slugs {
		allowed[s] = struct{}{}
	}
	out := make(refs.Catalog, len(allowed))
	for slug, vars := range full {
		if _, ok := allowed[slug]; ok {
			out[slug] = vars
		}
	}
	return out, nil
}

func publishedBucket(raw json.RawMessage, masterKey []byte, slug string) (map[string]string, error) {
	cfg, err := dbkind.ParseBucket(raw)
	if err != nil {
		return nil, err
	}
	accessKeyID := ""
	secret := ""
	if len(cfg.Keys) > 0 {
		accessKeyID = cfg.Keys[0].AccessKeyID
		sec, err := dbkind.DecryptKeySecret(masterKey, cfg.Keys[0])
		if err != nil {
			return nil, err
		}
		secret = sec
	}
	endpoint := "http://dy-garage:3900"
	vars := map[string]string{
		"S3_ENDPOINT":          endpoint,
		"S3_REGION":            "garage",
		"S3_BUCKET":            cfg.BucketName,
		"S3_ACCESS_KEY_ID":     accessKeyID,
		"S3_SECRET_ACCESS_KEY": secret,
	}
	if cfg.PublicRead {
		vars["S3_PUBLIC_URL"] = endpoint + "/" + cfg.BucketName
	}
	return vars, nil
}
