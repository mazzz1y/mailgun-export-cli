# mailgun-export-csv

CLI tool to export email events from the Mailgun Events API to CSV.

Uses the [mailgun-go](https://github.com/mailgun/mailgun-go) SDK.

## Install

```bash
go build -o mailgun-export-csv ./cmd/mailgun-export-csv
```

## Usage

```bash
mailgun-export-csv --api-key KEY --domain example.com --begin 72h
```

### Environment Variables

| Variable | Description |
|---|---|
| `MAILGUN_API_KEY` | API key (alternative to `--api-key`) |
| `MAILGUN_DOMAIN` | Domain (alternative to `--domain`) |
| `MAILGUN_REGION` | Region: `us` or `eu` (alternative to `--region`) |

### Flags

| Flag | Description |
|---|---|
| `--api-key` | Mailgun API key |
| `--domain` | Mailgun sending domain |
| `--region` | API region: `us` (default) or `eu` |
| `--begin` | Start time (RFC3339 or duration: `72h`, `3d`; default: `72h`) |
| `--end` | End time (RFC3339 or duration: `1h`) |
| `--events` | Event types, comma-separated (exact match) |
| `--from` | Filter by sender (full email: server-side, partial: client-side substring) |
| `--to` | Filter by recipient (full email: server-side, partial: client-side substring) |
| `--subject` | Filter by subject (substring match) |
| `--tag` | Filter by tag (exact match) |
| `--columns` | Columns to include, comma-separated |
| `-o, --output` | Output file (default: stdout) |
| `--limit` | Max events to export |

### Available Columns

`timestamp`, `event`, `from`, `to`, `subject`, `message_id`, `recipient`, `recipient_domain`, `recipient_provider`, `severity`, `delivery_status_code`, `delivery_status_message`, `tags`, `ip`, `country`, `device`, `client_name`, `user_agent`

### Event Types

`accepted`, `delivered`, `failed`, `rejected`, `clicked`, `opened`, `unsubscribed`, `stored`, `complained`

## Examples

```bash
# Export delivered events from the last 24 hours
mailgun-export-csv --begin 24h --events delivered -o delivered.csv

# Export specific columns for a date range
mailgun-export-csv \
  --begin 2026-03-01T00:00:00Z \
  --end 2026-03-05T00:00:00Z \
  --columns timestamp,event,from,to,subject \
  -o export.csv

# Filter by sender, pipe to stdout
mailgun-export-csv --from alice@example.com --begin 72h

# EU region, multiple event types, limit to 500
mailgun-export-csv --region eu --events accepted,delivered,failed --limit 500 -o events.csv

# Using environment variables
export MAILGUN_API_KEY=key-xxx
export MAILGUN_DOMAIN=example.com
mailgun-export-csv --begin 72h
```

## Filters

| Flag | Match type |
|---|---|
| `--events`, `--tag` | exact match (server-side) |
| `--subject` | substring match (server-side) |
| `--from`, `--to` | see below |

**`--from` and `--to` behavior:**

- **Full email address** (e.g. `alice@example.com`): filtered server-side via the Mailgun Events API (fast).
- **Partial string** (e.g. `example.com`, `alice`): filtered client-side with case-insensitive substring matching.

```bash
# Server-side filter (fast) -- full email address
mailgun-export-csv --to bob@example.com --begin 72h

# Client-side filter (slower) -- partial string / domain
mailgun-export-csv --to example.com --begin 72h
# warning: --to "example.com" is not a complete email address; filtering client-side (slower, all events will be fetched)
```