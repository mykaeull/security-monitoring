package httpxscan

// HttpxRecord representa o que será salvo na tabela httpx
type HttpxRecord struct {
	Host         string
	Status       string
	Title        string
	Location     string
	URL          string
	Technologies []string
}
