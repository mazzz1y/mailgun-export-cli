package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/mailgun/mailgun-go/v5/events"
)

type CSVWriter struct {
	writer  *csv.Writer
	columns []string
}

func NewCSVWriter(w io.Writer, selected []string) (*CSVWriter, error) {
	if len(selected) == 0 {
		selected = ColumnNames()
	}

	for _, col := range selected {
		if _, ok := pathsByColumn[col]; !ok {
			return nil, fmt.Errorf("unknown column: %q (available: %s)", col, strings.Join(ColumnNames(), ", "))
		}
	}

	cw := &CSVWriter{
		writer:  csv.NewWriter(w),
		columns: selected,
	}

	if err := cw.writer.Write(selected); err != nil {
		return nil, fmt.Errorf("writing header: %w", err)
	}

	return cw, nil
}

func (cw *CSVWriter) WriteEvent(e events.Event) error {
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshaling event: %w", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("unmarshaling event: %w", err)
	}

	row := make([]string, len(cw.columns))
	for i, col := range cw.columns {
		row[i] = resolveColumn(raw, col, e)
	}
	return cw.writer.Write(row)
}

func (cw *CSVWriter) Flush() error {
	cw.writer.Flush()
	return cw.writer.Error()
}

func resolveColumn(raw map[string]any, col string, e events.Event) string {
	if col == "timestamp" {
		return e.GetTimestamp().Format("2006-01-02T15:04:05Z")
	}

	for _, path := range pathsByColumn[col] {
		if val := resolve(raw, path); val != "" {
			return val
		}
	}

	return ""
}

func resolve(raw map[string]any, path []string) string {
	var current any = raw

	for _, key := range path {
		m, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current, ok = m[key]
		if !ok {
			return ""
		}
	}

	return toString(current)
}

func toString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	case []any:
		parts := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, ";")
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
}
