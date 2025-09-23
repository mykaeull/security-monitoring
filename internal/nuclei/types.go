package subdomainenum

type Request struct {
	ApexDomain string `json:"apex_domain"`
}

type STListResponse struct {
	Records []STRecord `json:"records"`
}

type STRecord struct {
	HostProvider []string `json:"host_provider"`
	Hostname     string   `json:"hostname"`
	IPs          []string `json:"ips"`
}
