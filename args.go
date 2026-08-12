package main

import (
	"fmt"
	"strconv"
)

// parseID splits a leading positional "<id>" from the remaining flag
// arguments, e.g. ["5", "-name", "x"] -> (5, ["-name", "x"]).
func parseID(rest []string, usage string) (int, []string, error) {
	if len(rest) < 1 {
		return 0, nil, fmt.Errorf("usage: %s", usage)
	}
	id, err := strconv.Atoi(rest[0])
	if err != nil {
		return 0, nil, fmt.Errorf("invalid id %q: %w", rest[0], err)
	}
	return id, rest[1:], nil
}

// parseArg splits a leading positional string argument (e.g. a slug) from
// the remaining flag arguments.
func parseArg(rest []string, usage string) (string, []string, error) {
	if len(rest) < 1 {
		return "", nil, fmt.Errorf("usage: %s", usage)
	}
	return rest[0], rest[1:], nil
}
