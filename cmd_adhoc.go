package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"
)

// Redash job status codes.
const (
	jobStatusSuccess = 3
	jobStatusFailure = 4
)

func runAdhoc(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: adhoc run -data-source-id N -query \"SQL\" [-max-age N] [-timeout 60s]")
	}
	sub, rest := args[0], args[1:]
	if sub != "run" {
		return fmt.Errorf("unknown adhoc subcommand: %s", sub)
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	c := newClient(cfg)

	fs := flag.NewFlagSet("adhoc run", flag.ExitOnError)
	dataSourceID := fs.Int("data-source-id", 0, "data source ID (required)")
	query := fs.String("query", "", "SQL text (required)")
	maxAge := fs.Int("max-age", 0, "max cache age in seconds; 0 forces a fresh run")
	timeout := fs.Duration("timeout", 60*time.Second, "how long to wait for an async query job")
	if err := fs.Parse(rest); err != nil {
		return err
	}
	if *dataSourceID == 0 || *query == "" {
		return fmt.Errorf("adhoc run requires -data-source-id and -query")
	}

	body := map[string]any{
		"query":          *query,
		"data_source_id": *dataSourceID,
		"max_age":        *maxAge,
	}
	raw, err := c.Post("/api/query_results", body)
	if err != nil {
		return err
	}

	var probe struct {
		Job *struct {
			ID string `json:"id"`
		} `json:"job"`
		QueryResult json.RawMessage `json:"query_result"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return printJSON(raw) // not the shape we expect; hand back the raw body
	}
	if probe.QueryResult != nil && string(probe.QueryResult) != "null" {
		return printJSON(raw)
	}
	if probe.Job == nil {
		return printJSON(raw)
	}

	resultID, err := c.pollJob(probe.Job.ID, *timeout)
	if err != nil {
		return err
	}
	result, err := c.Get(fmt.Sprintf("/api/query_results/%d", resultID), nil)
	if err != nil {
		return err
	}
	return printJSON(result)
}

// pollJob waits for a Redash async query job to finish and returns its
// query_result ID.
func (c *Client) pollJob(jobID string, timeout time.Duration) (int, error) {
	deadline := time.Now().Add(timeout)
	for {
		raw, err := c.Get(fmt.Sprintf("/api/jobs/%s", jobID), nil)
		if err != nil {
			return 0, err
		}
		var jr struct {
			Job struct {
				Status        int    `json:"status"`
				QueryResultID *int   `json:"query_result_id"`
				Error         string `json:"error"`
			} `json:"job"`
		}
		if err := json.Unmarshal(raw, &jr); err != nil {
			return 0, fmt.Errorf("decode job status: %w", err)
		}

		switch jr.Job.Status {
		case jobStatusSuccess:
			if jr.Job.QueryResultID == nil {
				return 0, fmt.Errorf("query job succeeded without a result ID")
			}
			return *jr.Job.QueryResultID, nil
		case jobStatusFailure:
			return 0, fmt.Errorf("query job failed: %s", jr.Job.Error)
		}

		if time.Now().After(deadline) {
			return 0, fmt.Errorf("timed out waiting for query job %s", jobID)
		}
		time.Sleep(time.Second)
	}
}
