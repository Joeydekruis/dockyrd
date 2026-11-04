package dbquery

import (
	"context"
	"fmt"
	"strings"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/docker"
)

// SchemaColumn is one column on a table.
type SchemaColumn struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
}

// SchemaTable groups columns under a schema-qualified table name.
type SchemaTable struct {
	Schema  string         `json:"schema"`
	Name    string         `json:"name"`
	Columns []SchemaColumn `json:"columns"`
}

// ListPostgresSchema returns user tables and columns from information_schema.
func ListPostgresSchema(ctx context.Context, cfg config.Config, dc *docker.Client, svc sqlc.Service) ([]SchemaTable, error) {
	if svc.Kind != "postgres" {
		return nil, fmt.Errorf("schema browser supports postgres only")
	}
	res, err := RunPostgres(ctx, cfg, dc, svc, `
SELECT table_schema, table_name, column_name, data_type
FROM information_schema.columns
WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
ORDER BY table_schema, table_name, ordinal_position
`, 5000)
	if err != nil {
		return nil, err
	}
	type key struct{ schema, name string }
	grouped := map[key][]SchemaColumn{}
	order := []key{}
	for _, row := range res.Rows {
		if len(row) < 4 {
			continue
		}
		k := key{schema: row[0], name: row[1]}
		if _, ok := grouped[k]; !ok {
			order = append(order, k)
		}
		grouped[k] = append(grouped[k], SchemaColumn{Name: row[2], DataType: row[3]})
	}
	out := make([]SchemaTable, 0, len(order))
	for _, k := range order {
		out = append(out, SchemaTable{
			Schema:  k.schema,
			Name:    k.name,
			Columns: grouped[k],
		})
	}
	return out, nil
}

// QualifiedTable returns "schema"."table" for SQL insertion.
func QualifiedTable(schema, name string) string {
	if strings.TrimSpace(schema) == "" {
		return quoteIdent(name)
	}
	return quoteIdent(schema) + "." + quoteIdent(name)
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
