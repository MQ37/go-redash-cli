// Command redash-cli is a plain CLI for the Redash REST API, meant for
// AI agents and scripts that prefer a subprocess call over an MCP server.
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd, args := os.Args[1], os.Args[2:]
	var err error
	switch cmd {
	case "queries":
		err = runQueries(args)
	case "adhoc":
		err = runAdhoc(args)
	case "datasources":
		err = runDataSources(args)
	case "dashboards":
		err = runDashboards(args)
	case "widgets":
		err = runWidgets(args)
	case "visualizations":
		err = runVisualizations(args)
	case "help", "-h", "--help":
		usage()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `go-redash-cli - plain CLI for the Redash REST API

Usage: redash-cli <command> <subcommand> [flags]

Commands:
  queries         list, get, create, update, archive, run, run-csv
  adhoc           run
  datasources     list, schema
  dashboards      list, get, create, delete
  widgets         list, get, create, update, delete
  visualizations  get, create, update, delete

Environment:
  REDASH_URL       Redash instance URL (required)
  REDASH_API_KEY   Redash API key (required)
  REDASH_TIMEOUT   Request timeout in milliseconds (default 30000)

Every command prints the raw Redash JSON response to stdout.
Run "redash-cli <command>" with no further args to see its flags.
`)
}
