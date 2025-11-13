package models

// SearchResult represents a search result
type SearchResult struct {
	ID       string                 `json:"id"`
	Score    float32                `json:"score"`
	Text     string                 `json:"text"`
	Source   string                 `json:"source"`   // document name or slack thread
	Metadata map[string]interface{} `json:"metadata"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query  string `json:"query" binding:"required"`
	TopK   int    `json:"top_k"`
	Filter string `json:"filter"` // 'documents', 'slack', or empty for all
}
