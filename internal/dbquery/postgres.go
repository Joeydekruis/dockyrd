package dbquery

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Joeydekruis/dockyrd/internal/config"
	"github.com/Joeydekruis/dockyrd/internal/crypto"
	"github.com/Joeydekruis/dockyrd/internal/db/sqlc"
	"github.com/Joeydekruis/dockyrd/internal/dbkind"
	"github.com/Joeydekruis/dockyrd/internal/docker"
	"github.com/Joeydekruis/dockyrd/internal/serviceops"
)

const defaultMaxRows = 200

// Result is a tabular query result for the dashboard.
type Result struct {
	Columns    []string   `json:"columns"`
	Rows       [][]string `json:"rows"`
	RowCount   int        `json:"row_count"`
	Truncated  bool       `json:"truncated"`
	DurationMS int64      `json:"duration_ms"`
}

// RunPostgres executes a read-only SQL statement via psql inside the database container.
func RunPostgres(ctx context.Context, cfg config.Config, dc *docker.Client, svc sqlc.Service, sqlText string, maxRows int) (Result, error) {
	if maxRows <= 0 || maxRows > 1000 {
		maxRows = defaultMaxRows
	}
	trimmed := strings.TrimSpace(sqlText)
	if trimmed == "" {
		return Result{}, fmt.Errorf("sql is required")
	}
	if err := validateReadOnlySQL(trimmed); err != nil {
		return Result{}, err
	}
	if svc.Kind != "postgres" {
		return Result{}, fmt.Errorf("query studio supports postgres only")
	}
	containerID, err := serviceops.FindDatabaseContainerID(ctx, dc, svc.ID)
	if err != nil {
		return Result{}, err
	}
	key, err := crypto.LoadMasterKey(cfg.SecretKeyPath())
	if err != nil {
		return Result{}, err
	}
	raw, err := configJSON(svc.Config)
	if err != nil {
		return Result{}, err
	}
	pg, _, err := dbkind.ParsePostgres(raw, key)
	if err != nil {
		return Result{}, err
	}
	limit := maxRows + 1
	wrapped := fmt.Sprintf(
		"COPY (SELECT * FROM (%s) AS dyq LIMIT %d) TO STDOUT WITH (FORMAT csv, HEADER true)",
		trimmed,
		limit,
	)
	start := time.Now()
	var buf strings.Builder
	cmd := []string{"psql", "-U", pg.User, "-d", pg.Database, "-v", "ON_ERROR_STOP=1", "-c", wrapped}
	if err := dc.ExecStream(ctx, containerID, cmd, &buf); err != nil {
		return Result{}, fmt.Errorf("query failed: %w", err)
	}
	parsed, truncated, err := parseCSV(buf.String(), maxRows)
	if err != nil {
		return Result{}, err
	}
	return Result{
		Columns:    parsed.columns,
		Rows:       parsed.rows,
		RowCount:   len(parsed.rows),
		Truncated:  truncated,
		DurationMS: time.Since(start).Milliseconds(),
	}, nil
}

func validateReadOnlySQL(sql string) error {
	upper := strings.ToUpper(strings.TrimSpace(sql))
	if strings.Contains(sql, ";") {
		return fmt.Errorf("only a single statement is allowed")
	}
	for _, prefix := range []string{"INSERT", "UPDATE", "DELETE", "DROP", "ALTER", "CREATE", "TRUNCATE", "GRANT", "REVOKE", "COPY"} {
		if strings.HasPrefix(upper, prefix+" ") || upper == prefix {
			return fmt.Errorf("only read-only SELECT queries are allowed")
		}
	}
	if !strings.HasPrefix(upper, "SELECT") && !strings.HasPrefix(upper, "WITH") && !strings.HasPrefix(upper, "TABLE") && !strings.HasPrefix(upper, "EXPLAIN") {
		return fmt.Errorf("only SELECT, WITH, TABLE, or EXPLAIN queries are allowed")
	}
	return nil
}

type csvParsed struct {
	columns []string
	rows    [][]string
}

func parseCSV(raw string, maxRows int) (csvParsed, bool, error) {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return csvParsed{}, false, nil
	}
	columns := splitCSVLine(lines[0])
	var rows [][]string
	truncated := false
	for i := 1; i < len(lines); i++ {
		if lines[i] == "" {
			continue
		}
		if len(rows) >= maxRows {
			truncated = true
			break
		}
		rows = append(rows, splitCSVLine(lines[i]))
	}
	return csvParsed{columns: columns, rows: rows}, truncated, nil
}

func splitCSVLine(line string) []string {
	var out []string
	var cur strings.Builder
	inQuote := false
	for i := 0; i < len(line); i++ {
		c := line[i]
		if c == '"' {
			inQuote = !inQuote
			continue
		}
		if c == ',' && !inQuote {
			out = append(out, cur.String())
			cur.Reset()
			continue
		}
		cur.WriteByte(c)
	}
	out = append(out, cur.String())
	return out
}

func configJSON(v any) (json.RawMessage, error) {
	switch x := v.(type) {
	case json.RawMessage:
		return x, nil
	case []byte:
		return json.RawMessage(x), nil
	case string:
		return json.RawMessage(x), nil
	default:
		return json.Marshal(x)
	}
}
