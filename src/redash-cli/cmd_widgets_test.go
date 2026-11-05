package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestCreateWidgetVisualizationField(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"text", []string{"-text", "# OUTDATED"}, "null"},
		{"visualization", []string{"-visualization-id", "42"}, "42"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var called atomic.Bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called.Store(true)
				if r.Method != http.MethodPost || r.URL.Path != "/api/widgets" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					http.Error(w, "unexpected request", http.StatusBadRequest)
					return
				}
				var body map[string]json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
					http.Error(w, "bad JSON", http.StatusBadRequest)
					return
				}
				// Older Redash handlers pop this key without a default, returning 500 if absent.
				value, ok := body["visualization_id"]
				if !ok {
					http.Error(w, "missing visualization_id", http.StatusInternalServerError)
					return
				}
				if string(value) != tc.want {
					t.Errorf("visualization_id = %s, want %s", value, tc.want)
				}
				if string(body["dashboard_id"]) != "123" {
					t.Errorf("dashboard_id = %s, want 123", body["dashboard_id"])
				}
				if tc.name == "text" && string(body["text"]) != `"# OUTDATED"` {
					t.Errorf("text = %s", body["text"])
				}
				fmt.Fprint(w, `{"id":1}`)
			}))
			defer server.Close()
			t.Setenv("REDASH_URL", server.URL)
			t.Setenv("REDASH_API_KEY", "test-api-key")
			args := append([]string{"create", "-dashboard-id", "123"}, tc.args...)
			if err := runWidgets(args); err != nil {
				t.Fatal(err)
			}
			if !called.Load() {
				t.Fatal("no widget request sent")
			}
		})
	}
}
