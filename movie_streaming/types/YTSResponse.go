package types

type MoviesResponse struct {
	Status        string        `json:"status"`
	StatusMessage string        `json:"status_message"`
	Data          DataForMovies `json:"data"`
	Meta          Meta          `json:"@meta"` // Handles the special '@' character
}

// Data holds the payload container.
type DataForMovies struct {
	Movie      []Movie `json:"movies"`
	Limit      int     `json:"limit"`
	PageNumber int     `json:"page_number"`
}

// MovieResponse represents the root level of the JSON response.
type MovieResponse struct {
	Status        string `json:"status"`
	StatusMessage string `json:"status_message"`
	Data          Data   `json:"data"`
	Meta          Meta   `json:"@meta"` // Handles the special '@' character
}

// Data holds the payload container.
type Data struct {
	Movie Movie `json:"movie"`
}

// Movie contains all the granular details about the film.
type Movie struct {
	ID                      uint64    `json:"id"`
	URL                     string    `json:"url"`
	IMDBCode                string    `json:"imdb_code"`
	Title                   string    `json:"title"`
	TitleEnglish            string    `json:"title_english"`
	TitleLong               string    `json:"title_long"`
	Slug                    string    `json:"slug"`
	Year                    int       `json:"year"`
	Rating                  float64   `json:"rating"`
	Runtime                 int       `json:"runtime"`
	Genres                  []string  `json:"genres"`
	LikeCount               int       `json:"like_count"`
	DescriptionIntro        string    `json:"description_intro"`
	DescriptionFull         string    `json:"description_full"`
	YTOfficialTrailerCode   string    `json:"yt_trailer_code"`
	Language                string    `json:"language"`
	MPARating               string    `json:"mpa_rating"`
	BackgroundImage         string    `json:"background_image"`
	BackgroundImageOriginal string    `json:"background_image_original"`
	SmallCoverImage         string    `json:"small_cover_image"`
	MediumCoverImage        string    `json:"medium_cover_image"`
	LargeCoverImage         string    `json:"large_cover_image"`
	Torrents                []Torrent `json:"torrents"`
	DateUploaded            string    `json:"date_uploaded"`
	DateUploadedUnix        int64     `json:"date_uploaded_unix"` // int64 is safer for timestamps
}

// Torrent holds the data for individual downloadable links.
type Torrent struct {
	URL              string `json:"url"`
	Hash             string `json:"hash"`
	Quality          string `json:"quality"`
	Type             string `json:"type"`
	IsRepack         string `json:"is_repack"` // Stored as a string ("0") in JSON
	VideoCodec       string `json:"video_codec"`
	BitDepth         string `json:"bit_depth"`
	AudioChannels    string `json:"audio_channels"`
	Seeds            int    `json:"seeds"`
	Peers            int    `json:"peers"`
	Size             string `json:"size"`
	SizeBytes        int64  `json:"size_bytes"` // int64 used for large file sizes in bytes
	DateUploaded     string `json:"date_uploaded"`
	DateUploadedUnix int64  `json:"date_uploaded_unix"`
}

// Meta captures the API meta information block.
type Meta struct {
	APIVersion    int    `json:"api_version"`
	ExecutionTime string `json:"execution_time"`
}
