package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func envFallback(flagVal, envKey string) string {
	if flagVal != "" {
		return flagVal
	}
	return os.Getenv(envKey)
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			result = append(result, p)
		}
	}
	return result
}

func parseTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}

	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}

	s = strings.TrimSpace(s)
	if days, ok := strings.CutSuffix(s, "d"); ok {
		var n int
		if _, err := fmt.Sscanf(days, "%d", &n); err == nil {
			return time.Now().Add(-time.Duration(n) * 24 * time.Hour), nil
		}
	}

	if d, err := time.ParseDuration(s); err == nil {
		return time.Now().Add(-d), nil
	}

	return time.Time{}, fmt.Errorf("expected RFC3339 (2006-01-02T15:04:05Z) or duration (72h, 3d)")
}
