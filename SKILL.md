---
name: go-redash-cli
description: Use the go-redash-cli (`redash-cli`) binary to talk to a Redash instance's REST API from the shell — list/run/create/update queries, execute ad-hoc SQL, inspect data-source schemas, and manage dashboards, widgets, and visualizations. Use when the user asks to query Redash, build or edit a Redash dashboard, or when an MCP Redash server isn't available and a CLI subprocess is preferred.
---

# go-redash-cli

Plain CLI wrapping the Redash REST API. No MCP, no server. Every command
prints raw Redash JSON to stdout; non-zero exit and stderr message on error.

## Setup

Required env vars before any command:

```
REDASH_URL=https://your-redash-instance.com
REDASH_API_KEY=your_api_key
```

Optional: `REDASH_TIMEOUT` (ms, default 30000).

## Commands

```
redash-cli queries list [-page N] [-page-size N]
redash-cli queries get <id>
redash-cli queries create -name X -data-source-id N -query "SQL" [-description X] [-draft]
redash-cli queries update <id> [-name X] [-data-source-id N] [-query SQL] [-description X] [-archived] [-draft true|false]
redash-cli queries archive <id>
redash-cli queries run <id> [-max-age N]       # cached results.json; -max-age 0 forces refresh
redash-cli queries run-csv <id>

redash-cli adhoc run -data-source-id N -query "SQL" [-max-age N] [-timeout 60s]

redash-cli datasources list
redash-cli datasources schema <id> [-page N] [-page-size N] [-search STR]

redash-cli dashboards list [-page N] [-page-size N]
redash-cli dashboards get <id-or-slug>
redash-cli dashboards create -name X
redash-cli dashboards delete <id>

redash-cli widgets list
redash-cli widgets get <id>
redash-cli widgets create -dashboard-id N [-visualization-id N] [-text X] [-width N] [-position-json '{"col":0,"row":0,"sizeX":3,"sizeY":8}']
redash-cli widgets update <id> [-visualization-id N] [-text X] [-width N] [-position-json '{}']
redash-cli widgets delete <id>

redash-cli visualizations get <id>
redash-cli visualizations create -query-id N -type X -name X [-options-json '{}']
redash-cli visualizations update <id> [-type X] [-name X] [-options-json '{}']
redash-cli visualizations delete <id>
```

## Building a dashboard end to end

A visualization only appears on a dashboard once attached as a widget —
creating a visualization does not place it anywhere by itself.

1. `queries create` — get `id` from the JSON response.
2. `visualizations create -query-id <id>` — get the visualization `id`.
3. `dashboards create` — get the dashboard `id`.
4. `widgets create -dashboard-id <id> -visualization-id <id> -position-json '...'` — attaches it.
5. `dashboards get <id>` — verify layout.

Extract IDs from JSON with `jq .id` (or equivalent) since every command
prints the full raw response, not just the ID.

## Notes

- Redash always creates queries as drafts server-side, no matter what the
  create request sends. `queries create` un-drafts automatically after
  creating; pass `-draft` to leave it as a draft. Use `queries update <id>
  -draft true|false` to flip draft status later (e.g. before attaching a
  visualization built on it to a dashboard).
- IDs are always the first positional argument, before any flags:
  `redash-cli queries update 5 -name "New name"`, not the reverse.
- `adhoc run` polls Redash's async job endpoint automatically when a query
  isn't cached; increase `-timeout` for slow queries.
- `datasources schema` fetches the full schema once, then paginates/filters
  it client-side — safe to call repeatedly with different `-search` values.
