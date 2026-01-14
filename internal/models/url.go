package models

import "time"

type URL struct {
	ID          string     `json:"id"`
	OriginalURL string     `json:"original_url"`
	ShortCode   string     `json:"short_code"`
	UserID      *int       `json:"user_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	ClickCount  int        `json:"click_count,omitempty"` // NEW
}

type ShortenRequest struct {
	URL         string  `json:"url"`
	CustomAlias *string `json:"custom_alias,omitempty"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
}

type ShortenResponse struct {
	ShortURL    string     `json:"short_url"`
	OriginalURL string     `json:"original_url"`
	ShortCode   string     `json:"short_code"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// NEW: Click analytics model
type URLClick struct {
	ID        int       `json:"id"`
	URLID     string    `json:"url_id"`
	ClickedAt time.Time `json:"clicked_at"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	Referer   string    `json:"referer,omitempty"`
	Country   string    `json:"country,omitempty"`
	City      string    `json:"city,omitempty"`
}

// NEW: Analytics summary
type URLAnalytics struct {
	ShortCode    string         `json:"short_code"`
	OriginalURL  string         `json:"original_url"`
	TotalClicks  int            `json:"total_clicks"`
	UniqueIPs    int            `json:"unique_ips"`
	RecentClicks []*URLClick    `json:"recent_clicks"`
	ClicksByDay  map[string]int `json:"clicks_by_day"`
	TopCountries map[string]int `json:"top_countries"`
	TopReferers  map[string]int `json:"top_referers"`
}
