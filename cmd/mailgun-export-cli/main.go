package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mazzz1y/mailgun-export-cli/internal/export"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	var (
		apiKey, domain, region string
		begin, end             string
		events, from, to       string
		subject, tag           string
		columns, output        string
		limit                  int
	)

	cmd := &cobra.Command{
		Use:          "mailgun-export-cli",
		Short:        "Export email events from Mailgun to CSV",
		Long:         longDescription,
		Example:      usageExamples,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(
				apiKey, domain, region,
				begin, end,
				events, from, to,
				subject, tag,
				columns, output,
				limit,
			)
		},
	}

	f := cmd.Flags()
	f.StringVar(&apiKey, "api-key", "",
		fmt.Sprintf("%s", "Mailgun API key [$MAILGUN_API_KEY]"))
	f.StringVar(&domain, "domain", "",
		fmt.Sprintf("%s", "Mailgun domain [$MAILGUN_DOMAIN]"))
	f.StringVar(&region, "region", "",
		fmt.Sprintf("%s", "API region: us or eu [$MAILGUN_REGION]"))
	f.StringVar(&begin, "begin", "",
		fmt.Sprintf("%s", "Start time (RFC3339 or duration like 72h, 3d; default: 72h)"))
	f.StringVar(&end, "end", "",
		fmt.Sprintf("%s", "End time (RFC3339 or duration like 1h)"))
	f.StringVar(&events, "events", "",
		fmt.Sprintf("%s", "Event types, exact match (comma-separated: accepted,delivered,failed,...)"))
	f.StringVar(&from, "from", "",
		fmt.Sprintf("%s", "Filter by sender (full email: server-side, partial: client-side substring)"))
	f.StringVar(&to, "to", "",
		fmt.Sprintf("%s", "Filter by recipient (full email: server-side, partial: client-side substring)"))
	f.StringVar(&subject, "subject", "",
		fmt.Sprintf("%s", "Filter by subject, substring match"))
	f.StringVar(&tag, "tag", "",
		fmt.Sprintf("%s", "Filter by tag, exact match"))
	f.StringVar(&columns, "columns", "",
		fmt.Sprintf(
			"Columns to include (comma-separated, default: all) Available: %s",
			strings.Join(export.ColumnNames(), ", ")))
	f.StringVarP(&output, "output", "o", "",
		fmt.Sprintf("%s", "Output file path (default: stdout)"))
	f.IntVar(&limit, "limit", 0,
		fmt.Sprintf("%s", "Maximum number of events to export (0 = no limit)"))

	return cmd
}
