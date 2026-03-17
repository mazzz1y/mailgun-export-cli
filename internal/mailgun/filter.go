package mailgun

import (
	"fmt"
	"net/mail"
	"strings"
)

type clientFilter struct {
	to   string
	from string
}

func (cf clientFilter) hasFilters() bool {
	return cf.to != "" || cf.from != ""
}

func (cf clientFilter) matchEvent(raw map[string]any) bool {
	if cf.to != "" {
		to := extractHeader(raw, "to")
		if !strings.Contains(strings.ToLower(to), strings.ToLower(cf.to)) {
			recipient, _ := raw["recipient"].(string)
			if !strings.Contains(strings.ToLower(recipient), strings.ToLower(cf.to)) {
				return false
			}
		}
	}
	if cf.from != "" {
		from := extractHeader(raw, "from")
		if !strings.Contains(strings.ToLower(from), strings.ToLower(cf.from)) {
			return false
		}
	}
	return true
}

func (opts ExportOpts) Warnings() []string {
	const tpl = "--%s %q is not a complete email address;" +
		" filtering client-side (slower, all events will be fetched)"

	var warnings []string
	_, cf := opts.buildFilter()
	if cf.to != "" {
		warnings = append(warnings, fmt.Sprintf(tpl, "to", opts.To))
	}
	if cf.from != "" {
		warnings = append(warnings, fmt.Sprintf(tpl, "from", opts.From))
	}
	return warnings
}

func (opts ExportOpts) buildFilter() (map[string]string, clientFilter) {
	f := make(map[string]string)
	var cf clientFilter

	if opts.Events != "" {
		f["event"] = opts.Events
	}
	if opts.From != "" {
		if isEmailAddress(opts.From) {
			f["from"] = opts.From
		} else {
			cf.from = opts.From
		}
	}
	if opts.To != "" {
		if isEmailAddress(opts.To) {
			f["to"] = opts.To
		} else {
			cf.to = opts.To
		}
	}
	if opts.Subject != "" {
		f["subject"] = opts.Subject
	}
	if opts.Tag != "" {
		f["tags"] = opts.Tag
	}

	return f, cf
}

func isEmailAddress(s string) bool {
	_, err := mail.ParseAddress(s)
	return err == nil
}

func extractHeader(raw map[string]any, key string) string {
	msg, _ := raw["message"].(map[string]any)
	if msg == nil {
		return ""
	}
	headers, _ := msg["headers"].(map[string]any)
	if headers == nil {
		return ""
	}
	val, _ := headers[key].(string)
	return val
}
