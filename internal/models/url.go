package models

import "time"

type URL struct {
	ID          string    `json:"id"`
	OriginalURL string    `json:"original_url"`
	ShortCode   string    `json:"short_code"`
	UserID      *int      `json:"user_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	ShortCode   string `json:"short_code"`
}
