package mailgun

import (
	"testing"
)

func TestBuildFilterEmpty(t *testing.T) {
	f, cf := ExportOpts{}.buildFilter()
	if len(f) != 0 {
		t.Errorf("expected empty filter, got %v", f)
	}
	if cf.hasFilters() {
		t.Errorf("expected no client filters, got %+v", cf)
	}
}

func TestBuildFilterAllFieldsFullEmails(t *testing.T) {
	f, cf := ExportOpts{
		Events:  "delivered",
		From:    "alice@example.com",
		To:      "bob@example.com",
		Subject: "Hello",
		Tag:     "newsletter",
	}.buildFilter()

	expected := map[string]string{
		"event":   "delivered",
		"from":    "alice@example.com",
		"to":      "bob@example.com",
		"subject": "Hello",
		"tags":    "newsletter",
	}

	if len(f) != len(expected) {
		t.Fatalf("expected %d filters, got %d: %v", len(expected), len(f), f)
	}

	for k, v := range expected {
		if f[k] != v {
			t.Errorf("filter[%q]: expected %q, got %q", k, v, f[k])
		}
	}

	if cf.hasFilters() {
		t.Errorf("expected no client filters for full emails, got %+v", cf)
	}
}

func TestBuildFilterPartialFullEmail(t *testing.T) {
	f, cf := ExportOpts{
		Events: "failed",
		To:     "user@test.com",
	}.buildFilter()

	if len(f) != 2 {
		t.Fatalf("expected 2 filters, got %d", len(f))
	}
	if f["event"] != "failed" {
		t.Errorf("event: expected failed, got %q", f["event"])
	}
	if f["to"] != "user@test.com" {
		t.Errorf("to: expected user@test.com, got %q", f["to"])
	}
	if _, ok := f["from"]; ok {
		t.Error("from should not be in filter")
	}
	if cf.hasFilters() {
		t.Errorf("expected no client filters for full email, got %+v", cf)
	}
}

func TestBuildFilterPartialToGoesClientSide(t *testing.T) {
	f, cf := ExportOpts{
		Events: "delivered",
		To:     "example.com",
	}.buildFilter()

	if _, ok := f["to"]; ok {
		t.Error("partial --to should not be in API filter")
	}
	if f["event"] != "delivered" {
		t.Errorf("event: expected delivered, got %q", f["event"])
	}
	if cf.to != "example.com" {
		t.Errorf("client filter to: expected example.com, got %q", cf.to)
	}
	if cf.from != "" {
		t.Errorf("client filter from should be empty, got %q", cf.from)
	}
}

func TestBuildFilterPartialFromGoesClientSide(t *testing.T) {
	_, cf := ExportOpts{
		From: "alice",
	}.buildFilter()

	if cf.from != "alice" {
		t.Errorf("client filter from: expected alice, got %q", cf.from)
	}
}

func TestBuildFilterMixedServerAndClientSide(t *testing.T) {
	f, cf := ExportOpts{
		From: "alice@example.com",
		To:   "example.com",
	}.buildFilter()

	if f["from"] != "alice@example.com" {
		t.Errorf("from should be in API filter, got %v", f)
	}
	if _, ok := f["to"]; ok {
		t.Error("partial to should not be in API filter")
	}
	if cf.to != "example.com" {
		t.Errorf("client filter to: expected example.com, got %q", cf.to)
	}
	if cf.from != "" {
		t.Errorf("client filter from should be empty, got %q", cf.from)
	}
}

func TestBuildFilterTagMapsToTags(t *testing.T) {
	f, _ := ExportOpts{Tag: "promo"}.buildFilter()

	if f["tags"] != "promo" {
		t.Errorf("expected tags=promo, got %q", f["tags"])
	}
	if _, ok := f["tag"]; ok {
		t.Error("should use 'tags' key, not 'tag'")
	}
}

func TestIsEmailAddress(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"alice@example.com", true},
		{"user@test.co.uk", true},
		{"a@b", true},
		{"\"Bob Smith\" <bob@example.com>", true},

		{"example.com", false},
		{"alice", false},
		{"@example.com", false},
		{"alice@", false},
		{"", false},
		{"not an email at all", false},
	}

	for _, tt := range tests {
		got := isEmailAddress(tt.input)
		if got != tt.want {
			t.Errorf("isEmailAddress(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestClientFilterMatchEvent(t *testing.T) {
	makeEvent := func(from, to, recipient string) map[string]any {
		raw := map[string]any{}
		if recipient != "" {
			raw["recipient"] = recipient
		}
		msg := map[string]any{}
		headers := map[string]any{}
		if from != "" {
			headers["from"] = from
		}
		if to != "" {
			headers["to"] = to
		}
		msg["headers"] = headers
		raw["message"] = msg
		return raw
	}

	tests := []struct {
		name   string
		cf     clientFilter
		raw    map[string]any
		expect bool
	}{
		{
			name:   "no filters matches everything",
			cf:     clientFilter{},
			raw:    makeEvent("alice@example.com", "bob@test.com", "bob@test.com"),
			expect: true,
		},
		{
			name:   "to domain match in header",
			cf:     clientFilter{to: "test.com"},
			raw:    makeEvent("alice@example.com", "bob@test.com", "bob@test.com"),
			expect: true,
		},
		{
			name:   "to domain no match",
			cf:     clientFilter{to: "other.com"},
			raw:    makeEvent("alice@example.com", "bob@test.com", "bob@test.com"),
			expect: false,
		},
		{
			name:   "to matches in recipient field when header missing",
			cf:     clientFilter{to: "test.com"},
			raw:    makeEvent("alice@example.com", "", "bob@test.com"),
			expect: true,
		},
		{
			name:   "to case insensitive",
			cf:     clientFilter{to: "TEST.COM"},
			raw:    makeEvent("alice@example.com", "bob@test.com", "bob@test.com"),
			expect: true,
		},
		{
			name:   "from substring match",
			cf:     clientFilter{from: "example"},
			raw:    makeEvent("alice@example.com", "bob@test.com", "bob@test.com"),
			expect: true,
		},
		{
			name:   "from no match",
			cf:     clientFilter{from: "other.com"},
			raw:    makeEvent("alice@example.com", "bob@test.com", "bob@test.com"),
			expect: false,
		},
		{
			name:   "both filters must match - both pass",
			cf:     clientFilter{from: "alice", to: "test.com"},
			raw:    makeEvent("alice@example.com", "bob@test.com", "bob@test.com"),
			expect: true,
		},
		{
			name:   "both filters must match - to fails",
			cf:     clientFilter{from: "alice", to: "other.com"},
			raw:    makeEvent("alice@example.com", "bob@test.com", "bob@test.com"),
			expect: false,
		},
		{
			name:   "to username match",
			cf:     clientFilter{to: "bob"},
			raw:    makeEvent("alice@example.com", "bob@test.com", "bob@test.com"),
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cf.matchEvent(tt.raw)
			if got != tt.expect {
				t.Errorf("matchEvent() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestClientFilterHasFilters(t *testing.T) {
	if (clientFilter{}).hasFilters() {
		t.Error("empty clientFilter should not have filters")
	}
	if !(clientFilter{to: "x"}).hasFilters() {
		t.Error("clientFilter with to should have filters")
	}
	if !(clientFilter{from: "x"}).hasFilters() {
		t.Error("clientFilter with from should have filters")
	}
}

func TestWarningsPartialTo(t *testing.T) {
	opts := ExportOpts{To: "example.com"}
	warnings := opts.Warnings()
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
}

func TestWarningsPartialFrom(t *testing.T) {
	opts := ExportOpts{From: "alice"}
	warnings := opts.Warnings()
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(warnings))
	}
}

func TestWarningsFullEmail(t *testing.T) {
	opts := ExportOpts{To: "bob@example.com", From: "alice@example.com"}
	warnings := opts.Warnings()
	if len(warnings) != 0 {
		t.Fatalf("expected 0 warnings for full emails, got %d: %v", len(warnings), warnings)
	}
}

func TestWarningsNone(t *testing.T) {
	opts := ExportOpts{}
	warnings := opts.Warnings()
	if len(warnings) != 0 {
		t.Fatalf("expected 0 warnings, got %d", len(warnings))
	}
}
