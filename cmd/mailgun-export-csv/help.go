package main

const longDescription = `Export email events from the Mailgun Events API to CSV format.

Supports filtering by event type, sender, recipient, subject, and tags.
Choose which columns to include in the output.

Filter modes:
  --events      exact match (server-side)
  --from        full email: server-side; partial string: client-side substring
  --to          full email: server-side; partial string: client-side substring
  --subject     substring match (server-side)
  --tag         exact match (server-side)
  Multiple values per flag are comma-separated.

Authentication via --api-key flag or MAILGUN_API_KEY environment variable.
Domain via --domain flag or MAILGUN_DOMAIN environment variable.
Region via --region flag or MAILGUN_REGION environment variable (default: us).`

const usageExamples = `  # Export all events from the last 3 days
  mailgun-export-csv --api-key KEY --domain example.com --begin 72h

  # Export delivered events for a specific date range
  mailgun-export-csv --events delivered --begin 2026-03-01T00:00:00Z --end 2026-03-05T00:00:00Z

  # Export only specific columns to a file
  mailgun-export-csv --columns timestamp,event,from,to,subject -o export.csv

  # Filter by sender and recipient
  mailgun-export-csv --from alice@example.com --to bob@example.com

  # Use environment variables
  export MAILGUN_API_KEY=key-xxx
  export MAILGUN_DOMAIN=example.com
  mailgun-export-csv --begin 24h --events accepted,delivered`
