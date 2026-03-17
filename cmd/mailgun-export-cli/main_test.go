package main

import (
	"testing"
	"time"
)

func TestSplitCSV(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", nil},
		{"a", []string{"a"}},
		{"a,b,c", []string{"a", "b", "c"}},
		{" a , b , c ", []string{"a", "b", "c"}},
		{"a,,b", []string{"a", "b"}},
		{",", nil},
	}

	for _, tt := range tests {
		got := splitCSV(tt.input)
		if len(got) != len(tt.expected) {
			t.Errorf("splitCSV(%q): expected %v, got %v", tt.input, tt.expected, got)
			continue
		}
		for i := range got {
			if got[i] != tt.expected[i] {
				t.Errorf("splitCSV(%q)[%d]: expected %q, got %q", tt.input, i, tt.expected[i], got[i])
			}
		}
	}
}

func TestParseTimeEmpty(t *testing.T) {
	result, err := parseTime("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsZero() {
		t.Errorf("expected zero time, got %v", result)
	}
}

func TestParseTimeRFC3339(t *testing.T) {
	result, err := parseTime("2026-03-05T12:00:00Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Date(2026, 3, 5, 12, 0, 0, 0, time.UTC)
	if !result.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestParseTimeDays(t *testing.T) {
	before := time.Now()
	result, err := parseTime("3d")
	after := time.Now()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedEarliest := before.Add(-3 * 24 * time.Hour)
	expectedLatest := after.Add(-3 * 24 * time.Hour)

	if result.Before(expectedEarliest) || result.After(expectedLatest) {
		t.Errorf("3d should be ~72h ago, got %v", result)
	}
}

func TestParseTimeHours(t *testing.T) {
	before := time.Now()
	result, err := parseTime("24h")
	after := time.Now()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedEarliest := before.Add(-24 * time.Hour)
	expectedLatest := after.Add(-24 * time.Hour)

	if result.Before(expectedEarliest) || result.After(expectedLatest) {
		t.Errorf("24h should be ~24h ago, got %v", result)
	}
}

func TestParseTimeInvalid(t *testing.T) {
	_, err := parseTime("not-a-time")
	if err == nil {
		t.Fatal("expected error for invalid time")
	}
}

func TestEnvFallback(t *testing.T) {
	if got := envFallback("flag", "SOME_VAR"); got != "flag" {
		t.Errorf("expected flag value, got %q", got)
	}

	t.Setenv("TEST_MAILGUN_VAR", "envval")
	if got := envFallback("", "TEST_MAILGUN_VAR"); got != "envval" {
		t.Errorf("expected envval, got %q", got)
	}

	if got := envFallback("", "NONEXISTENT_VAR_12345"); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}
