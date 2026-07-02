package models

type ExternalBook struct {
	Title       string   `json:"title"`
	Authors     []string `json:"authors"`
	Description string   `json:"description,omitempty"`
	CoverURL    string   `json:"cover_url,omitempty"`
}
