package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	mg "github.com/mailgun/mailgun-go/v5"

	"github.com/mazzz1y/mailgun-export-csv/internal/export"
	"github.com/mazzz1y/mailgun-export-csv/internal/mailgun"
)

func run(
	apiKey, domain, region,
	begin, end,
	events, from, to,
	subject, tag,
	columns, output string,
	limit int,
) error {
	apiKey = envFallback(apiKey, "MAILGUN_API_KEY")
	domain = envFallback(domain, "MAILGUN_DOMAIN")
	region = envFallback(region, "MAILGUN_REGION")

	if apiKey == "" {
		return fmt.Errorf("--api-key or MAILGUN_API_KEY is required")
	}
	if domain == "" {
		return fmt.Errorf("--domain or MAILGUN_DOMAIN is required")
	}

	client := mg.NewMailgun(apiKey)
	if strings.EqualFold(region, "eu") {
		client.SetAPIBase(mg.APIBaseEU)
	}

	beginTime, err := parseTime(begin)
	if err != nil {
		return fmt.Errorf("invalid --begin: %w", err)
	}
	endTime, err := parseTime(end)
	if err != nil {
		return fmt.Errorf("invalid --end: %w", err)
	}

	opts := mailgun.ExportOpts{
		Domain:  domain,
		Begin:   beginTime,
		End:     endTime,
		Events:  events,
		From:    from,
		To:      to,
		Subject: subject,
		Tag:     tag,
		Limit:   limit,
	}

	for _, w := range opts.Warnings() {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}

	f, cleanup, err := openOutput(output)
	if err != nil {
		return err
	}
	defer cleanup()

	csvWriter, err := export.NewCSVWriter(f, splitCSV(columns))
	if err != nil {
		return err
	}

	exporter := mailgun.NewExporter(client, opts)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	total, err := exporter.Export(ctx, csvWriter)
	if err != nil {
		return err
	}

	if output != "" {
		fmt.Fprintf(os.Stderr, "exported %d events to %s\n", total, output)
	} else {
		fmt.Fprintf(os.Stderr, "exported %d events\n", total)
	}

	return nil
}

func openOutput(path string) (*os.File, func(), error) {
	if path == "" {
		return os.Stdout, func() {}, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, nil, fmt.Errorf("creating output file: %w", err)
	}
	return f, func() { f.Close() }, nil
}
