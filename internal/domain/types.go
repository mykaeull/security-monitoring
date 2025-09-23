package domain

type Request struct {
	Domain []string `json:"domain"`
}

type Response struct {
	Inserted int      `json:"inserted"`
	Skipped  int      `json:"skipped"`
	Details  []string `json:"details"`
}
