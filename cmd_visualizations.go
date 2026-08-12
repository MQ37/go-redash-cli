package main

import (
	"encoding/json"
	"flag"
	"fmt"
)

func runVisualizations(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: visualizations <get|create|update|delete> [flags]")
	}
	sub, rest := args[0], args[1:]

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	c := newClient(cfg)

	switch sub {
	case "get":
		id, flagArgs, err := parseID(rest, "visualizations get <id>")
		if err != nil {
			return err
		}
		if err := flag.NewFlagSet("visualizations get", flag.ExitOnError).Parse(flagArgs); err != nil {
			return err
		}
		raw, err := c.Get(fmt.Sprintf("/api/visualizations/%d", id), nil)
		if err != nil {
			return err
		}
		return printJSON(raw)

	case "create":
		fs := flag.NewFlagSet("visualizations create", flag.ExitOnError)
		queryID := fs.Int("query-id", 0, "query ID (required)")
		vizType := fs.String("type", "", "visualization type, e.g. CHART, TABLE (required)")
		name := fs.String("name", "", "visualization name (required)")
		description := fs.String("description", "", "description")
		optionsJSON := fs.String("options-json", "{}", "visualization options as a JSON object")
		if err := fs.Parse(rest); err != nil {
			return err
		}
		if *queryID == 0 || *vizType == "" || *name == "" {
			return fmt.Errorf("visualizations create requires -query-id, -type, -name")
		}
		options, err := decodeOptions(*optionsJSON)
		if err != nil {
			return err
		}
		body := map[string]any{
			"query_id":    *queryID,
			"type":        *vizType,
			"name":        *name,
			"description": *description,
			"options":     options,
		}
		raw, err := c.Post("/api/visualizations", body)
		if err != nil {
			return err
		}
		return printJSON(raw)

	case "update":
		id, flagArgs, err := parseID(rest, "visualizations update <id> [flags]")
		if err != nil {
			return err
		}
		fs := flag.NewFlagSet("visualizations update", flag.ExitOnError)
		vizType := fs.String("type", "", "new visualization type")
		name := fs.String("name", "", "new name")
		description := fs.String("description", "", "new description")
		optionsJSON := fs.String("options-json", "", "new options as a JSON object")
		if err := fs.Parse(flagArgs); err != nil {
			return err
		}
		body := map[string]any{}
		if *vizType != "" {
			body["type"] = *vizType
		}
		if *name != "" {
			body["name"] = *name
		}
		if *description != "" {
			body["description"] = *description
		}
		if *optionsJSON != "" {
			options, err := decodeOptions(*optionsJSON)
			if err != nil {
				return err
			}
			body["options"] = options
		}
		raw, err := c.Post(fmt.Sprintf("/api/visualizations/%d", id), body)
		if err != nil {
			return err
		}
		return printJSON(raw)

	case "delete":
		id, flagArgs, err := parseID(rest, "visualizations delete <id>")
		if err != nil {
			return err
		}
		if err := flag.NewFlagSet("visualizations delete", flag.ExitOnError).Parse(flagArgs); err != nil {
			return err
		}
		if err := c.Delete(fmt.Sprintf("/api/visualizations/%d", id)); err != nil {
			return err
		}
		fmt.Printf("visualization %d deleted\n", id)
		return nil

	default:
		return fmt.Errorf("unknown visualizations subcommand: %s", sub)
	}
}

func decodeOptions(raw string) (map[string]any, error) {
	options := map[string]any{}
	if err := json.Unmarshal([]byte(raw), &options); err != nil {
		return nil, fmt.Errorf("invalid -options-json: %w", err)
	}
	return options, nil
}
