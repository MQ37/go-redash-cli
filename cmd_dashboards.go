package main

import (
	"flag"
	"fmt"
	"net/url"
	"strconv"
)

func runDashboards(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: dashboards <list|get|create|delete> [flags]")
	}
	sub, rest := args[0], args[1:]

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	c := newClient(cfg)

	switch sub {
	case "list":
		fs := flag.NewFlagSet("dashboards list", flag.ExitOnError)
		page := fs.Int("page", 1, "page number")
		pageSize := fs.Int("page-size", 25, "results per page")
		if err := fs.Parse(rest); err != nil {
			return err
		}
		q := url.Values{"page": {strconv.Itoa(*page)}, "page_size": {strconv.Itoa(*pageSize)}}
		raw, err := c.Get("/api/dashboards", q)
		if err != nil {
			return err
		}
		return printJSON(raw)

	case "get":
		idOrSlug, flagArgs, err := parseArg(rest, "dashboards get <id-or-slug>")
		if err != nil {
			return err
		}
		if err := flag.NewFlagSet("dashboards get", flag.ExitOnError).Parse(flagArgs); err != nil {
			return err
		}
		raw, err := c.Get(fmt.Sprintf("/api/dashboards/%s", idOrSlug), nil)
		if err != nil {
			return err
		}
		return printJSON(raw)

	case "create":
		fs := flag.NewFlagSet("dashboards create", flag.ExitOnError)
		name := fs.String("name", "", "dashboard name (required)")
		if err := fs.Parse(rest); err != nil {
			return err
		}
		if *name == "" {
			return fmt.Errorf("dashboards create requires -name")
		}
		raw, err := c.Post("/api/dashboards", map[string]any{"name": *name})
		if err != nil {
			return err
		}
		return printJSON(raw)

	case "delete":
		id, flagArgs, err := parseID(rest, "dashboards delete <id>")
		if err != nil {
			return err
		}
		if err := flag.NewFlagSet("dashboards delete", flag.ExitOnError).Parse(flagArgs); err != nil {
			return err
		}
		if err := c.Delete(fmt.Sprintf("/api/dashboards/%d", id)); err != nil {
			return err
		}
		fmt.Printf("dashboard %d deleted\n", id)
		return nil

	default:
		return fmt.Errorf("unknown dashboards subcommand: %s", sub)
	}
}
