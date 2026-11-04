package dbkind

import (
	"fmt"
	"net/url"
)

// Published returns connection variables for a postgres or redis service (briefing §7.3).
func Published(kind, serviceSlug, password string, pg PostgresConfig) map[string]string {
	switch kind {
	case "postgres":
		host := serviceSlug
		port := "5432"
		user := pg.User
		db := pg.Database
		u := url.URL{
			Scheme: "postgres",
			User:   url.UserPassword(user, password),
			Host:   fmt.Sprintf("%s:%s", host, port),
			Path:   db,
		}
		q := u.Query()
		q.Set("sslmode", "disable")
		u.RawQuery = q.Encode()
		return map[string]string{
			"DATABASE_URL": u.String(),
			"PGHOST":       host,
			"PGPORT":       port,
			"PGUSER":       user,
			"PGPASSWORD":   password,
			"PGDATABASE":   db,
		}
	case "redis":
		host := serviceSlug
		port := "6379"
		u := &url.URL{
			Scheme: "redis",
			Host:   fmt.Sprintf("%s:%s", host, port),
			User:   url.UserPassword("", password),
		}
		return map[string]string{
			"REDIS_URL":      u.String(),
			"REDIS_HOST":     host,
			"REDIS_PORT":     port,
			"REDIS_PASSWORD": password,
		}
	default:
		return nil
	}
}

// PublishedPostgresAt builds connection vars for an external host:port.
func PublishedPostgresAt(host, port, password string, pg PostgresConfig) map[string]string {
	user := pg.User
	db := pg.Database
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   db,
	}
	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()
	return map[string]string{
		"DATABASE_URL": u.String(),
		"PGHOST":       host,
		"PGPORT":       port,
		"PGUSER":       user,
		"PGPASSWORD":   password,
		"PGDATABASE":   db,
	}
}
