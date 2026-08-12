# go-redash-cli

<img src="logo.svg" alt="go-redash-cli" width="240">

Plain CLI for the [Redash](https://redash.io) REST API. No MCP, no server,
zero dependencies outside the Go standard library — for AI agents and scripts
that prefer a subprocess call over an MCP tool call.

## Install

```bash
go install github.com/MQ37/go-redash-cli@latest
```

This installs the `go-redash-cli` binary. Build it as `redash-cli` instead if
you prefer a shorter name:

```bash
git clone https://github.com/MQ37/go-redash-cli
cd go-redash-cli
go build -o redash-cli .
```

## Configuration

Set these environment variables before running any command:

| Variable          | Required | Description                              |
|--------------------|----------|-------------------------------------------|
| `REDASH_URL`       | yes      | Redash instance URL, e.g. `https://redash.example.com` |
| `REDASH_API_KEY`   | yes      | Redash API key                            |
| `REDASH_TIMEOUT`   | no       | Request timeout in milliseconds (default `30000`) |

Every command prints the raw Redash JSON response to stdout and exits
non-zero with an error on stderr on failure.

## Commands

```
redash-cli queries list [-page N] [-page-size N]
redash-cli queries get <id>
redash-cli queries create -name X -data-source-id N -query "SQL" [-description X]
redash-cli queries update <id> [-name X] [-data-source-id N] [-query SQL] [-description X] [-archived]
redash-cli queries archive <id>
redash-cli queries run <id> [-max-age N]
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

Run `redash-cli help` for a summary, or `redash-cli <command>` with no
subcommand to see the usage line for that command's flags.

## Example: build a dashboard end to end

```bash
export REDASH_URL=https://redash.example.com
export REDASH_API_KEY=your_api_key

# 1. create a query
query_id=$(redash-cli queries create -name "Signups" -data-source-id 1 \
  -query "select date_trunc('day', created_at), count(*) from users group by 1" \
  | jq .id)

# 2. create a chart visualization for it
viz_id=$(redash-cli visualizations create -query-id "$query_id" -type CHART \
  -name "Signups over time" | jq .id)

# 3. create a dashboard
dashboard_id=$(redash-cli dashboards create -name "Growth" | jq .id)

# 4. attach the visualization to the dashboard as a widget
redash-cli widgets create -dashboard-id "$dashboard_id" -visualization-id "$viz_id" \
  -position-json '{"col":0,"row":0,"sizeX":3,"sizeY":8}'

# 5. verify
redash-cli dashboards get "$dashboard_id"
```

## Notes

- `queries run` reads `/api/queries/{id}/results.json`; pass `-max-age 0` to
  force a refresh instead of using Redash's cache.
- `adhoc run` posts to `/api/query_results` and, if Redash queues an async
  job, polls it until it finishes or `-timeout` elapses.
- `datasources schema` fetches the full schema from Redash and paginates or
  filters it client-side, since the API returns it in one response.

## License

MIT
