package subdomainenum

// Request recebido no seu endpoint local
type Request struct {
	ApexDomain string `json:"apex_domain"`
}

// Estruturas mínimas para parsear o retorno da SecurityTrails
type STListResponse struct {
	Records []STRecord `json:"records"`
}

type STRecord struct {
	HostProvider []string   `json:"host_provider"`
	Hostname     string     `json:"hostname"`
	IPs          []string   `json:"ips"`
}
