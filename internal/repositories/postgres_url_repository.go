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
		INSERT INTO urls (original_url, short_code, user_id, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
		`
	url.CreatedAt = time.Now()

	err := r.db.QueryRow(query, url.OriginalURL, url.ShortCode, url.UserID, url.CreatedAt).Scan(&url.ID)

	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresURLRepository) FindByShortCode(shotCode string) (*models.URL, error) {
	query := `
		SELECT id, original_url, short_code, user_id, created_at
		FROM urls
		WHERE short_code = $1
		`

	url := &models.URL{}
	err := r.db.QueryRow(query, shotCode).Scan(
		&url.ID,
		&url.OriginalURL,
		&url.ShortCode,
		&url.UserID,
		&url.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return url, nil
}

func (r *PostgresURLRepository) FindByOriginalURL(originalURL string) (*models.URL, error) {
	query := `
		SELECT id, original_url, short_code, user_id, created_at
		FROM urls
		WHERE original_url = $1
		`
	url := &models.URL{}
	err := r.db.QueryRow(query, originalURL).Scan(
		&url.ID,
		&url.OriginalURL,
		&url.ShortCode,
		&url.UserID,
		&url.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return url, nil
}

func (r *PostgresURLRepository) FindByUserID(userID int) ([]*models.URL, error) {
	query := `
		SELECT id, original_url, short_code, user_id, created_at
		FROM urls
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var urls []*models.URL
	for rows.Next() {
		url := &models.URL{}
		err := rows.Scan(
			&url.ID,
			&url.OriginalURL,
			&url.ShortCode,
			&url.UserID,
			&url.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		urls = append(urls, url)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}
