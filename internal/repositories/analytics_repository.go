package repositories

import (
	"database/sql"
	"time"

	"github.com/techitdeveloper/url-shortener/internal/models"
)

// AnalyticsRepository defines the interface for analytics storage
type AnalyticsRepository interface {
	RecordClick(click *models.URLClick) error
	GetURLAnalytics(urlID string, days int) (*models.URLAnalytics, error)
	GetClicksByURL(urlID string, limit int) ([]*models.URLClick, error)
}

// PostgresAnalyticsRepository implements AnalyticsRepository with PostgreSQL
type PostgresAnalyticsRepository struct {
	db *sql.DB
}

// NewPostgresAnalyticsRepository creates a new PostgreSQL analytics repository
func NewPostgresAnalyticsRepository(db *sql.DB) *PostgresAnalyticsRepository {
	return &PostgresAnalyticsRepository{db: db}
}

func (r *PostgresAnalyticsRepository) RecordClick(click *models.URLClick) error {
	query := `
        INSERT INTO url_clicks (url_id, clicked_at, ip_address, user_agent, referer, country, city)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id
    `

	err := r.db.QueryRow(
		query,
		click.URLID,
		click.ClickedAt,
		click.IPAddress,
		click.UserAgent,
		click.Referer,
		click.Country,
		click.City,
	).Scan(&click.ID)

	return err
}

func (r *PostgresAnalyticsRepository) GetURLAnalytics(urlID string, days int) (*models.URLAnalytics, error) {
	// Get URL details
	urlQuery := `
        SELECT short_code, original_url
        FROM urls
        WHERE id = $1
    `

	analytics := &models.URLAnalytics{
		ClicksByDay:  make(map[string]int),
		TopCountries: make(map[string]int),
		TopReferers:  make(map[string]int),
	}

	err := r.db.QueryRow(urlQuery, urlID).Scan(
		&analytics.ShortCode,
		&analytics.OriginalURL,
	)
	if err != nil {
		return nil, err
	}

	// Calculate date range
	startDate := time.Now().AddDate(0, 0, -days)

	// Get total clicks
	totalClicksQuery := `
        SELECT COUNT(*)
        FROM url_clicks
        WHERE url_id = $1 AND clicked_at >= $2
    `

	err = r.db.QueryRow(totalClicksQuery, urlID, startDate).Scan(&analytics.TotalClicks)
	if err != nil {
		return nil, err
	}

	// Get unique IPs
	uniqueIPsQuery := `
        SELECT COUNT(DISTINCT ip_address)
        FROM url_clicks
        WHERE url_id = $1 AND clicked_at >= $2
    `

	err = r.db.QueryRow(uniqueIPsQuery, urlID, startDate).Scan(&analytics.UniqueIPs)
	if err != nil {
		return nil, err
	}

	// Get clicks by day
	clicksByDayQuery := `
        SELECT DATE(clicked_at) as day, COUNT(*) as count
        FROM url_clicks
        WHERE url_id = $1 AND clicked_at >= $2
        GROUP BY DATE(clicked_at)
        ORDER BY day DESC
    `

	rows, err := r.db.Query(clicksByDayQuery, urlID, startDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var day time.Time
		var count int
		if err := rows.Scan(&day, &count); err != nil {
			return nil, err
		}
		analytics.ClicksByDay[day.Format("2006-01-02")] = count
	}

	// Get top countries
	topCountriesQuery := `
        SELECT country, COUNT(*) as count
        FROM url_clicks
        WHERE url_id = $1 AND clicked_at >= $2 AND country IS NOT NULL
        GROUP BY country
        ORDER BY count DESC
        LIMIT 5
    `

	rows, err = r.db.Query(topCountriesQuery, urlID, startDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var country string
		var count int
		if err := rows.Scan(&country, &count); err != nil {
			return nil, err
		}
		analytics.TopCountries[country] = count
	}

	// Get top referers
	topReferersQuery := `
        SELECT referer, COUNT(*) as count
        FROM url_clicks
        WHERE url_id = $1 AND clicked_at >= $2 AND referer IS NOT NULL AND referer != ''
        GROUP BY referer
        ORDER BY count DESC
        LIMIT 5
    `

	rows, err = r.db.Query(topReferersQuery, urlID, startDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var referer string
		var count int
		if err := rows.Scan(&referer, &count); err != nil {
			return nil, err
		}
		analytics.TopReferers[referer] = count
	}

	// Get recent clicks
	recentClicks, err := r.GetClicksByURL(urlID, 10)
	if err != nil {
		return nil, err
	}
	analytics.RecentClicks = recentClicks

	return analytics, nil
}

func (r *PostgresAnalyticsRepository) GetClicksByURL(urlID string, limit int) ([]*models.URLClick, error) {
	query := `
        SELECT id, url_id, clicked_at, ip_address, user_agent, referer, country, city
        FROM url_clicks
        WHERE url_id = $1
        ORDER BY clicked_at DESC
        LIMIT $2
    `

	rows, err := r.db.Query(query, urlID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clicks []*models.URLClick
	for rows.Next() {
		click := &models.URLClick{}
		err := rows.Scan(
			&click.ID,
			&click.URLID,
			&click.ClickedAt,
			&click.IPAddress,
			&click.UserAgent,
			&click.Referer,
			&click.Country,
			&click.City,
		)
		if err != nil {
			return nil, err
		}
		clicks = append(clicks, click)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return clicks, nil
}
