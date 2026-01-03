package repositories

import (
	"database/sql"
	"time"

	"github.com/techitdeveloper/url-shortener/internal/models"
)

type PostgresURLRepository struct {
	db *sql.DB
}

func NewPostgresURLRepository(db *sql.DB) *PostgresURLRepository {
	return &PostgresURLRepository{db: db}
}

func (r *PostgresURLRepository) Save(url *models.URL) error {
	query := `
		INSERT INTO urls (original_url, short_code, created_at)
		VALUES ($1, $2, $3)
		RETURNING id
		`
	url.CreatedAt = time.Now()

	err := r.db.QueryRow(query, url.OriginalURL, url.ShortCode, url.CreatedAt).Scan(&url.ID)

	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresURLRepository) FindByShortCode(shotCode string) (*models.URL, error) {
	query := `
		SELECT id, original_url, short_code, created_at
		FROM urls
		WHERE short_code = $1
		`

	url := &models.URL{}
	err := r.db.QueryRow(query, shotCode).Scan(
		&url.ID,
		&url.OriginalURL,
		&url.ShortCode,
		&url.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return url, nil
}

func (r *PostgresURLRepository) FindByOriginalURL(originalURL string) (*models.URL, error) {
	query := `
		SELECT id, original_url, short_code, created_at
		FROM urls
		WHERE original_url = $1
		`
	url := &models.URL{}
	err := r.db.QueryRow(query, originalURL).Scan(
		&url.ID,
		&url.OriginalURL,
		&url.ShortCode,
		&url.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return url, nil
}
