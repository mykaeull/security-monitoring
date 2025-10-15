package katana

import "time"

type STRecord struct {
	OccurredAt time.Time `json:"occurred_at"`
	Request    struct {
		Method   string `json:"method"`
		Endpoint string `json:"endpoint"`
	} `json:"request"`
	Response struct {
		StatusCode    int    `json:"status_code"`
		ContentLength int64  `json:"content_length"`
		ContentType   string `json:"content-type"`
	} `json:"response"`
}
