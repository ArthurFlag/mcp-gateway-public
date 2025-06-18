package backup

type Backup struct {
	Config       string            `json:"config"`
	Registry     string            `json:"registry"`
	Catalog      string            `json:"catalog"`
	CatalogFiles map[string]string `json:"catalogFiles"`
	Secrets      map[string]string `json:"secrets"`
}
