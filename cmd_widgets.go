package main

import (
	"encoding/json"
	"flag"
	"fmt"
)

func runWidgets(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: widgets <list|get|create|update|delete> [flags]")
	}
	sub, rest := args[0], args[1:]

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	c := newClient(cfg)

	switch sub {
	case "list":
		if err := flag.NewFlagSet("widgets list", flag.ExitOnError).Parse(rest); err != nil {
			return err
		}
		raw, err := c.Get("/api/widgets", nil)
		if err != nil {
			return err
		}
		return printJSON(raw)

	case "get":
		id, flagArgs, err := parseID(rest, "widgets get <id>")
		if err != nil {
			return err
		}
		if err := flag.NewFlagSet("widgets get", flag.ExitOnError).Parse(flagArgs); err != nil {
			return err
		}
		raw, err := c.Get(fmt.Sprintf("/api/widgets/%d", id), nil)
		if err != nil {
			return err
		}
		return printJSON(raw)

	case "create":
		fs := flag.NewFlagSet("widgets create", flag.ExitOnError)
		dashboardID := fs.Int("dashboard-id", 0, "dashboard ID (required)")
		visualizationID := fs.Int("visualization-id", 0, "visualization ID (omit for a text widget)")
		text := fs.String("text", "", "text widget content")
		width := fs.Int("width", 1, "legacy widget width (1 or 2)")
		positionJSON := fs.String("position-json", "", `grid position, e.g. {"col":0,"row":0,"sizeX":3,"sizeY":8}`)
		if err := fs.Parse(rest); err != nil {
			return err
		}
		if *dashboardID == 0 {
			return fmt.Errorf("widgets create requires -dashboard-id")
		}

		options, err := widgetOptions(*positionJSON)
		if err != nil {
			return err
		}
		body := map[string]any{
			"dashboard_id": *dashboardID,
			"width":        *width,
			"options":      options,
		}
		if *visualizationID != 0 {
			body["visualization_id"] = *visualizationID
		}
		if *text != "" {
			body["text"] = *text
		}
		raw, err := c.Post("/api/widgets", body)
		if err != nil {
			return err
		}
		return printJSON(raw)

	case "update":
		id, flagArgs, err := parseID(rest, "widgets update <id> [flags]")
		if err != nil {
			return err
		}
		fs := flag.NewFlagSet("widgets update", flag.ExitOnError)
		visualizationID := fs.Int("visualization-id", 0, "new visualization ID")
		text := fs.String("text", "", "new text widget content")
		width := fs.Int("width", 0, "new legacy widget width (1 or 2)")
		positionJSON := fs.String("position-json", "", `new grid position, e.g. {"col":0,"row":0,"sizeX":3,"sizeY":8}`)
		if err := fs.Parse(flagArgs); err != nil {
			return err
		}

		body := map[string]any{}
		if *visualizationID != 0 {
			body["visualization_id"] = *visualizationID
		}
		if *text != "" {
			body["text"] = *text
		}
		if *width != 0 {
			body["width"] = *width
		}
		if *positionJSON != "" {
			options, err := widgetOptions(*positionJSON)
			if err != nil {
				return err
			}
			body["options"] = options
		}
		raw, err := c.Post(fmt.Sprintf("/api/widgets/%d", id), body)
		if err != nil {
			return err
		}
		return printJSON(raw)

	case "delete":
		id, flagArgs, err := parseID(rest, "widgets delete <id>")
		if err != nil {
			return err
		}
		if err := flag.NewFlagSet("widgets delete", flag.ExitOnError).Parse(flagArgs); err != nil {
			return err
		}
		if err := c.Delete(fmt.Sprintf("/api/widgets/%d", id)); err != nil {
			return err
		}
		fmt.Printf("widget %d deleted\n", id)
		return nil

	default:
		return fmt.Errorf("unknown widgets subcommand: %s", sub)
	}
}

func widgetOptions(positionJSON string) (map[string]any, error) {
	options := map[string]any{}
	if positionJSON == "" {
		return options, nil
	}
	var position any
	if err := json.Unmarshal([]byte(positionJSON), &position); err != nil {
		return nil, fmt.Errorf("invalid -position-json: %w", err)
	}
	options["position"] = position
	return options, nil
}
