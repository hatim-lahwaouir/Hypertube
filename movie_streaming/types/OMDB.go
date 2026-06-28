package types

// Movie represents an individual movie item inside the Search array.
type MovieOMDB struct {
	Title  string `json:"Title"`
	Year   string `json:"Year"`
	ImdbID string `json:"imdbID"`
	Type   string `json:"Type"`
	Poster string `json:"Poster"`
}

// MovieSearchResponse represents the top-level wrapper JSON object.
type MovieSearchResponseOMDB struct {
	Search       []MovieOMDB `json:"Search"`
	TotalResults string      `json:"totalResults"`
	Response     string      `json:"Response"` // Handled as a string since the value is "True"
}
