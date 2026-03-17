package export

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/mailgun/mailgun-go/v5/events"
)

func TestColumnNames(t *testing.T) {
	names := ColumnNames()

	if len(names) != len(fieldMappings) {
		t.Fatalf("expected %d columns, got %d", len(fieldMappings), len(names))
	}

	if names[0] != "timestamp" {
		t.Errorf("expected first column to be timestamp, got %q", names[0])
	}

	if names[1] != "event" {
		t.Errorf("expected second column to be event, got %q", names[1])
	}
}

func TestColumnNamesMatchPathIndex(t *testing.T) {
	for _, name := range ColumnNames() {
		if _, ok := pathsByColumn[name]; !ok && name != "timestamp" {
			t.Errorf("column %q has no entry in pathsByColumn", name)
		}
	}
}

func TestNewCSVWriterDefaultColumns(t *testing.T) {
	var buf bytes.Buffer
	cw, err := NewCSVWriter(&buf, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := cw.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	header := strings.TrimSpace(buf.String())
	expected := strings.Join(ColumnNames(), ",")
	if header != expected {
		t.Errorf("expected header %q, got %q", expected, header)
	}
}

func TestNewCSVWriterSelectedColumns(t *testing.T) {
	var buf bytes.Buffer
	cw, err := NewCSVWriter(&buf, []string{"timestamp", "event", "from"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := cw.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	header := strings.TrimSpace(buf.String())
	if header != "timestamp,event,from" {
		t.Errorf("expected header %q, got %q", "timestamp,event,from", header)
	}
}

func TestNewCSVWriterInvalidColumn(t *testing.T) {
	var buf bytes.Buffer
	_, err := NewCSVWriter(&buf, []string{"timestamp", "bogus"})
	if err == nil {
		t.Fatal("expected error for invalid column")
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Errorf("error should mention invalid column name, got: %v", err)
	}
}

func TestWriteEventAccepted(t *testing.T) {
	var buf bytes.Buffer
	cw, err := NewCSVWriter(&buf, []string{"timestamp", "event", "from", "to", "subject", "recipient", "ip", "tags"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ev := &events.Accepted{}
	ev.Timestamp = float64(time.Date(2026, 3, 5, 12, 0, 0, 0, time.UTC).Unix())
	ev.SetName("accepted")
	ev.Message.Headers.From = "alice@example.com"
	ev.Message.Headers.To = "bob@example.com"
	ev.Message.Headers.Subject = "Hello"
	ev.Recipient = "bob@example.com"
	ev.OriginatingIP = "1.2.3.4"
	ev.Tags = []string{"welcome", "onboarding"}

	if err := cw.WriteEvent(ev); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := cw.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (header + row), got %d", len(lines))
	}

	row := lines[1]
	fields := strings.Split(row, ",")
	if fields[0] != "2026-03-05T12:00:00Z" {
		t.Errorf("timestamp: expected 2026-03-05T12:00:00Z, got %q", fields[0])
	}
	if fields[1] != "accepted" {
		t.Errorf("event: expected accepted, got %q", fields[1])
	}
	if fields[2] != "alice@example.com" {
		t.Errorf("from: expected alice@example.com, got %q", fields[2])
	}
	if fields[3] != "bob@example.com" {
		t.Errorf("to: expected bob@example.com, got %q", fields[3])
	}
	if fields[4] != "Hello" {
		t.Errorf("subject: expected Hello, got %q", fields[4])
	}
	if fields[5] != "bob@example.com" {
		t.Errorf("recipient: expected bob@example.com, got %q", fields[5])
	}
	if fields[6] != "1.2.3.4" {
		t.Errorf("ip: expected 1.2.3.4, got %q", fields[6])
	}
	if fields[7] != "welcome;onboarding" {
		t.Errorf("tags: expected welcome;onboarding, got %q", fields[7])
	}
}

func TestWriteEventDelivered(t *testing.T) {
	var buf bytes.Buffer
	cw, err := NewCSVWriter(&buf, []string{"event", "recipient_provider", "delivery_status_code", "delivery_status_message"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ev := &events.Delivered{}
	ev.SetName("delivered")
	ev.RecipientProvider = "gmail"
	ev.DeliveryStatus.Code = 250
	ev.DeliveryStatus.Message = "OK"

	if err := cw.WriteEvent(ev); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := cw.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	row := lines[1]
	fields := strings.Split(row, ",")
	if fields[0] != "delivered" {
		t.Errorf("event: expected delivered, got %q", fields[0])
	}
	if fields[1] != "gmail" {
		t.Errorf("recipient_provider: expected gmail, got %q", fields[1])
	}
	if fields[2] != "250" {
		t.Errorf("delivery_status_code: expected 250, got %q", fields[2])
	}
	if fields[3] != "OK" {
		t.Errorf("delivery_status_message: expected OK, got %q", fields[3])
	}
}

func TestWriteEventFailed(t *testing.T) {
	var buf bytes.Buffer
	cw, err := NewCSVWriter(&buf, []string{"event", "severity", "delivery_status_code"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ev := &events.Failed{}
	ev.SetName("failed")
	ev.Severity = "permanent"
	ev.DeliveryStatus.Code = 550

	if err := cw.WriteEvent(ev); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := cw.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	row := lines[1]
	fields := strings.Split(row, ",")
	if fields[1] != "permanent" {
		t.Errorf("severity: expected permanent, got %q", fields[1])
	}
	if fields[2] != "550" {
		t.Errorf("delivery_status_code: expected 550, got %q", fields[2])
	}
}

func TestWriteEventOpened(t *testing.T) {
	var buf bytes.Buffer
	cw, err := NewCSVWriter(&buf, []string{"event", "ip", "country", "device", "client_name", "user_agent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ev := &events.Opened{}
	ev.SetName("opened")
	ev.IP = "5.6.7.8"
	ev.GeoLocation.Country = "US"
	ev.ClientInfo.DeviceType = "desktop"
	ev.ClientInfo.ClientName = "Thunderbird"
	ev.ClientInfo.UserAgent = "Mozilla/5.0"

	if err := cw.WriteEvent(ev); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := cw.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	row := lines[1]
	fields := strings.Split(row, ",")
	if fields[1] != "5.6.7.8" {
		t.Errorf("ip: expected 5.6.7.8, got %q", fields[1])
	}
	if fields[2] != "US" {
		t.Errorf("country: expected US, got %q", fields[2])
	}
	if fields[3] != "desktop" {
		t.Errorf("device: expected desktop, got %q", fields[3])
	}
	if fields[4] != "Thunderbird" {
		t.Errorf("client_name: expected Thunderbird, got %q", fields[4])
	}
	if fields[5] != "Mozilla/5.0" {
		t.Errorf("user_agent: expected Mozilla/5.0, got %q", fields[5])
	}
}

func TestIPFallbackToOriginatingIP(t *testing.T) {
	var buf bytes.Buffer
	cw, err := NewCSVWriter(&buf, []string{"ip"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ev := &events.Accepted{}
	ev.SetName("accepted")
	ev.OriginatingIP = "10.0.0.1"

	if err := cw.WriteEvent(ev); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := cw.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	row := lines[1]
	if row != "10.0.0.1" {
		t.Errorf("ip fallback: expected 10.0.0.1, got %q", row)
	}
}

func TestMissingFieldsReturnEmpty(t *testing.T) {
	var buf bytes.Buffer
	cw, err := NewCSVWriter(&buf, []string{"severity", "country", "delivery_status_code"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ev := &events.Accepted{}
	ev.SetName("accepted")

	if err := cw.WriteEvent(ev); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := cw.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	row := lines[1]
	if row != ",," {
		t.Errorf("expected empty fields ',,', got %q", row)
	}
}

func TestMultipleEvents(t *testing.T) {
	var buf bytes.Buffer
	cw, err := NewCSVWriter(&buf, []string{"event", "recipient"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ev1 := &events.Accepted{}
	ev1.SetName("accepted")
	ev1.Recipient = "a@test.com"

	ev2 := &events.Delivered{}
	ev2.SetName("delivered")
	ev2.Recipient = "b@test.com"

	if err := cw.WriteEvent(ev1); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := cw.WriteEvent(ev2); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := cw.Flush(); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[1] != "accepted,a@test.com" {
		t.Errorf("row 1: expected 'accepted,a@test.com', got %q", lines[1])
	}
	if lines[2] != "delivered,b@test.com" {
		t.Errorf("row 2: expected 'delivered,b@test.com', got %q", lines[2])
	}
}

func TestResolveNestedPath(t *testing.T) {
	raw := map[string]any{
		"message": map[string]any{
			"headers": map[string]any{
				"from": "test@example.com",
			},
		},
	}

	val := resolve(raw, []string{"message", "headers", "from"})
	if val != "test@example.com" {
		t.Errorf("expected test@example.com, got %q", val)
	}
}

func TestResolveMissingPath(t *testing.T) {
	raw := map[string]any{
		"message": map[string]any{},
	}

	val := resolve(raw, []string{"message", "headers", "from"})
	if val != "" {
		t.Errorf("expected empty string, got %q", val)
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{"hello", "hello"},
		{float64(250), "250"},
		{float64(3.14), "3.14"},
		{true, "true"},
		{[]any{"a", "b", "c"}, "a;b;c"},
		{[]any{}, ""},
		{nil, ""},
	}

	for _, tt := range tests {
		got := toString(tt.input)
		if got != tt.expected {
			t.Errorf("toString(%v): expected %q, got %q", tt.input, tt.expected, got)
		}
	}
}
