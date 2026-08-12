package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

// printJSON pretty-prints a raw JSON response to stdout.
func printJSON(raw json.RawMessage) error {
	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		fmt.Println(string(raw))
		return nil
	}
	fmt.Println(buf.String())
	return nil
}

// printRaw writes non-JSON payloads (e.g. CSV) verbatim to stdout.
func printRaw(data []byte) {
	os.Stdout.Write(data)
	if len(data) == 0 || data[len(data)-1] != '\n' {
		fmt.Println()
	}
}

func printJSONValue(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("encode output: %w", err)
	}
	fmt.Println(string(b))
	return nil
}
