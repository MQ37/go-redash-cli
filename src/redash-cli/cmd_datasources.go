package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"
)

func runDataSources(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: datasources <list|schema> [flags]")
	}
	sub, rest := args[0], args[1:]

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	c := newClient(cfg)

	switch sub {
	case "list":
		if err := flag.NewFlagSet("datasources list", flag.ExitOnError).Parse(rest); err != nil {
			return err
		}
		raw, err := c.Get("/api/data_sources", nil)
		if err != nil {
			return err
		}
		return printJSON(raw)

	case "schema":
		id, flagArgs, err := parseID(rest, "datasources schema <id> [-page N] [-page-size N] [-search STR]")
		if err != nil {
			return err
		}
		fs := flag.NewFlagSet("datasources schema", flag.ExitOnError)
		page := fs.Int("page", 1, "page number (starts at 1)")
		pageSize := fs.Int("page-size", 25, "tables per page")
		search := fs.String("search", "", "case-insensitive substring match on table name")
		if err := fs.Parse(flagArgs); err != nil {
			return err
		}

		raw, err := c.Get(fmt.Sprintf("/api/data_sources/%d/schema", id), nil)
		if err != nil {
			return err
		}
		return printSchemaPage(raw, *page, *pageSize, *search)

	default:
		return fmt.Errorf("unknown datasources subcommand: %s", sub)
	}
}

// printSchemaPage filters and paginates a Redash schema response client-side,
// since the API returns the full table list in one response.
func printSchemaPage(raw json.RawMessage, page, pageSize int, search string) error {
	tables, err := extractSchemaTables(raw)
	if err != nil {
		return err
	}

	if search != "" {
		needle := strings.ToLower(search)
		filtered := tables[:0]
		for _, t := range tables {
			var meta struct {
				Name string `json:"name"`
			}
			if err := json.Unmarshal(t, &meta); err != nil {
				continue
			}
			if strings.Contains(strings.ToLower(meta.Name), needle) {
				filtered = append(filtered, t)
			}
		}
		tables = filtered
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 25
	}
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > len(tables) {
		start = len(tables)
	}
	if end > len(tables) {
		end = len(tables)
	}

	out := struct {
		Tables   []json.RawMessage `json:"tables"`
		Page     int               `json:"page"`
		PageSize int               `json:"pageSize"`
		Total    int               `json:"total"`
		HasMore  bool              `json:"hasMore"`
	}{
		Tables:   tables[start:end],
		Page:     page,
		PageSize: pageSize,
		Total:    len(tables),
		HasMore:  end < len(tables),
	}
	return printJSONValue(out)
}

// extractSchemaTables handles the two response shapes seen across Redash
// versions: a bare array of tables, or {"schema": [...]}.
func extractSchemaTables(raw json.RawMessage) ([]json.RawMessage, error) {
	var tables []json.RawMessage
	if err := json.Unmarshal(raw, &tables); err == nil {
		return tables, nil
	}

	var wrapped struct {
		Schema []json.RawMessage `json:"schema"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, fmt.Errorf("unexpected schema response shape: %w", err)
	}
	return wrapped.Schema, nil
}
