package bypass

type SiteConfig struct {
	Pass bool `json:"pass"`
	C    struct {
		Type string `json:"type"`
		Req  string `json:"req"`
	} `json:"c"`
}
