package nuclei

import "time"

// STRecord representa o JSON inteiro que você mostrou.
type STRecord struct {
	TemplateID string    `json:"template-id,omitempty"`
	OccurredAt time.Time `json:"timestamp,omitempty"` // top-level timestamp -> time.Time
	Info       struct {
		Name           string   `json:"name,omitempty"`
		Description    string   `json:"description,omitempty"`
		Reference      []string `json:"reference,omitempty"`
		Severity       string   `json:"severity,omitempty"`
		Classification struct {
			CVEID []string `json:"cve-id,omitempty"`
			CWEID []string `json:"cwe-id,omitempty"`
		} `json:"classification,omitempty"`
	} `json:"info,omitempty"`

	Type        string `json:"type,omitempty"`
	Host        string `json:"host,omitempty"`
	Port        string `json:"port,omitempty"`
	Scheme      string `json:"scheme,omitempty"`
	URL         string `json:"url,omitempty"`
	MatchedAt   string `json:"matched-at,omitempty"`
	IP          string `json:"ip,omitempty"`
	CurlCommand string `json:"curl-command,omitempty"`
	Interaction struct {
		Protocol      string    `json:"protocol,omitempty"`
		UniqueID      string    `json:"unique-id,omitempty"`
		FullID        string    `json:"full-id,omitempty"`
		QType         string    `json:"q-type,omitempty"`
		RawRequest    string    `json:"raw-request,omitempty"`
		RawResponse   string    `json:"raw-response,omitempty"`
		RemoteAddress string    `json:"remote-address,omitempty"`
		Timestamp     time.Time `json:"timestamp,omitempty"`
	} `json:"interaction,omitempty"`
}
