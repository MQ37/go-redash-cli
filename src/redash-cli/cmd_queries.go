package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"strconv"
)

func runQueries(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: queries <list|get|create|update|archive|run|run-csv> [flags]")
	}
	sub, rest := args[0], args[1:]

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	c := newClient(cfg)

	switch sub {
	case "list":
		fs := flag.NewFlagSet("queries list", flag.ExitOnError)
		page := fs.Int("page", 1, "page number")
		pageSize := fs.Int("page-size", 25, "results per page")
		if err := fs.Parse(rest); err != nil {
			return err
		}
		q := url.Values{"page": {strconv.Itoa(*page)}, "page_size": {strconv.Itoa(*pageSize)}}
		raw, err := c.Get("/api/queries", q)
		if err != nil {
			return err
		}
		return printJSON(raw)

	case "get":
		id, flagArgs, err := parseID(rest, "queries get <id>")
		if err != nil {
			return err
		}
		if err := flag.NewFlagSet("queries get", flag.ExitOnError).Parse(flagArgs); err != nil {
			return err
		}
		raw, err := c.Get(fmt.Sprintf("/api/queries/%d", id), nil)
		if err != nil {
			return err
		}
		return printJSON(raw)

	case "create":
		fs := flag.NewFlagSet("queries create", flag.ExitOnError)
		name := fs.String("name", "", "query name (required)")
		dataSourceID := fs.Int("data-source-id", 0, "data source ID (required)")
		query := fs.String("query", "", "SQL text (required)")
		description := fs.String("description", "", "description")
		draft := fs.Bool("draft", false, "leave the query as a draft (Redash always creates queries as drafts server-side; this un-drafts by default)")
		if err := fs.Parse(rest); err != nil {
			return err
		}
		if *name == "" || *dataSourceID == 0 || *query == "" {
			return fmt.Errorf("queries create requires -name, -data-source-id, -query")
		}
		body := map[string]any{
			"name":           *name,
			"data_source_id": *dataSourceID,
			"query":          *query,
			"description":    *description,
			"options":        map[string]any{},
			"schedule":       nil,
		}
		raw, err := c.Post("/api/queries", body)
		if err != nil {
			return err
		}
		if *draft {
			return printJSON(raw)
		}

		var created struct {
			ID int `json:"id"`
		}
		if err := json.Unmarshal(raw, &created); err != nil {
			return fmt.Errorf("decode created query: %w", err)
		}
		undrafted, err := c.Post(fmt.Sprintf("/api/queries/%d", created.ID), map[string]any{"is_draft": false})
		if err != nil {
			return fmt.Errorf("query %d created but failed to un-draft it: %w", created.ID, err)
		}
		return printJSON(undrafted)

	case "update":
		id, flagArgs, err := parseID(rest, "queries update <id> [flags]")
		if err != nil {
			return err
		}
		fs := flag.NewFlagSet("queries update", flag.ExitOnError)
		name := fs.String("name", "", "new name")
		dataSourceID := fs.Int("data-source-id", 0, "new data source ID")
		query := fs.String("query", "", "new SQL text")
		description := fs.String("description", "", "new description")
		archived := fs.Bool("archived", false, "set the query as archived")
		draft := fs.String("draft", "", "true or false - set the query's draft status")
		if err := fs.Parse(flagArgs); err != nil {
			return err
		}
		body := map[string]any{}
		if *name != "" {
			body["name"] = *name
		}
		if *dataSourceID != 0 {
			body["data_source_id"] = *dataSourceID
		}
		if *query != "" {
			body["query"] = *query
		}
		if *description != "" {
			body["description"] = *description
		}
		if *archived {
			body["is_archived"] = true
		}
		if *draft != "" {
			isDraft, err := strconv.ParseBool(*draft)
			if err != nil {
				return fmt.Errorf("invalid -draft value %q: must be true or false", *draft)
			}
			body["is_draft"] = isDraft
		}
		raw, err := c.Post(fmt.Sprintf("/api/queries/%d", id), body)
		if err != nil {
			return err
		}
		return printJSON(raw)

	case "archive":
		id, flagArgs, err := parseID(rest, "queries archive <id>")
		if err != nil {
			return err
		}
		if err := flag.NewFlagSet("queries archive", flag.ExitOnError).Parse(flagArgs); err != nil {
			return err
		}
		if err := c.Delete(fmt.Sprintf("/api/queries/%d", id)); err != nil {
			return err
		}
		fmt.Printf("query %d archived\n", id)
		return nil

	case "run":
		id, flagArgs, err := parseID(rest, "queries run <id> [-max-age N]")
		if err != nil {
			return err
		}
		fs := flag.NewFlagSet("queries run", flag.ExitOnError)
		maxAge := fs.Int("max-age", -1, "max cache age in seconds; 0 forces a refresh")
		if err := fs.Parse(flagArgs); err != nil {
			return err
		}
		q := url.Values{}
		if *maxAge >= 0 {
			q.Set("max_age", strconv.Itoa(*maxAge))
		}
		raw, err := c.Get(fmt.Sprintf("/api/queries/%d/results.json", id), q)
		if err != nil {
			return err
		}
		return printJSON(raw)

	case "run-csv":
		id, flagArgs, err := parseID(rest, "queries run-csv <id>")
		if err != nil {
			return err
		}
		if err := flag.NewFlagSet("queries run-csv", flag.ExitOnError).Parse(flagArgs); err != nil {
			return err
		}
		raw, err := c.Get(fmt.Sprintf("/api/queries/%d/results.csv", id), nil)
		if err != nil {
			return err
		}
		printRaw(raw)
		return nil

	default:
		return fmt.Errorf("unknown queries subcommand: %s", sub)
	}
}
