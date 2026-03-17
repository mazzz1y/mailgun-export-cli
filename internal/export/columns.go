package export

type FieldMapping struct {
	Column string
	Paths  [][]string
}

var fieldMappings = []FieldMapping{
	{"timestamp", nil},
	{"event", [][]string{{"event"}}},
	{"from", [][]string{{"message", "headers", "from"}}},
	{"to", [][]string{{"message", "headers", "to"}}},
	{"subject", [][]string{{"message", "headers", "subject"}}},
	{"message_id", [][]string{{"message", "headers", "message-id"}}},
	{"recipient", [][]string{{"recipient"}}},
	{"recipient_domain", [][]string{{"recipient-domain"}}},
	{"recipient_provider", [][]string{{"recipient-provider"}}},
	{"severity", [][]string{{"severity"}}},
	{"delivery_status_code", [][]string{{"delivery-status", "code"}}},
	{"delivery_status_message", [][]string{{"delivery-status", "message"}}},
	{"tags", [][]string{{"tags"}}},
	{"ip", [][]string{{"ip"}, {"originating-ip"}}},
	{"country", [][]string{{"geolocation", "country"}}},
	{"device", [][]string{{"client-info", "device-type"}}},
	{"client_name", [][]string{{"client-info", "client-name"}}},
	{"user_agent", [][]string{{"client-info", "user-agent"}}},
}

var pathsByColumn = buildPathIndex()

func ColumnNames() []string {
	names := make([]string, len(fieldMappings))
	for i, fm := range fieldMappings {
		names[i] = fm.Column
	}
	return names
}

func buildPathIndex() map[string][][]string {
	m := make(map[string][][]string, len(fieldMappings))
	for _, fm := range fieldMappings {
		m[fm.Column] = fm.Paths
	}
	return m
}
