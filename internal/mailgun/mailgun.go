package mailgun

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	mg "github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/events"
)

type EventWriter interface {
	WriteEvent(events.Event) error
	Flush() error
}

type ExportOpts struct {
	Domain  string
	Begin   time.Time
	End     time.Time
	Events  string
	From    string
	To      string
	Subject string
	Tag     string
	Limit   int
}

type Exporter struct {
	client *mg.Client
	opts   ExportOpts
}

func NewExporter(client *mg.Client, opts ExportOpts) *Exporter {
	return &Exporter{client: client, opts: opts}
}

func (e *Exporter) Export(ctx context.Context, w EventWriter) (int, error) {
	apiFilter, cf := e.opts.buildFilter()

	it := e.client.ListEvents(e.opts.Domain, &mg.ListEventOptions{
		Begin:          e.opts.Begin,
		End:            e.opts.End,
		ForceAscending: true,
		Limit:          300,
		Filter:         apiFilter,
	})

	var page []events.Event
	total := 0

	for it.Next(ctx, &page) {
		for _, event := range page {
			if e.opts.Limit > 0 && total >= e.opts.Limit {
				return total, w.Flush()
			}

			if cf.hasFilters() {
				raw, err := eventToRaw(event)
				if err != nil {
					return total, fmt.Errorf("marshaling event for filter: %w", err)
				}
				if !cf.matchEvent(raw) {
					continue
				}
			}

			if err := w.WriteEvent(event); err != nil {
				return total, fmt.Errorf("writing event: %w", err)
			}
			total++
		}

		if err := w.Flush(); err != nil {
			return total, fmt.Errorf("flushing: %w", err)
		}
	}

	if it.Err() != nil {
		return total, fmt.Errorf("iterating events: %w", it.Err())
	}

	return total, nil
}

func eventToRaw(e events.Event) (map[string]any, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}
