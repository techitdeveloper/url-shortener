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
        INSERT INTO urls (original_url, short_code, user_id, created_at, expires_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `

	url.CreatedAt = time.Now()

	err := r.db.QueryRow(
		query,
		url.OriginalURL,
		url.ShortCode,
		url.UserID,
		url.CreatedAt,
		url.ExpiresAt,
	).Scan(&url.ID)

	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"urls_short_code_key\"" {
			return ErrShortCodeExists
		}
		return err
	}

	return nil
}

func (r *PostgresURLRepository) FindByShortCode(shortCode string) (*models.URL, error) {
	query := `
        SELECT id, original_url, short_code, user_id, created_at, expires_at
        FROM urls
        WHERE short_code = $1
    `

	url := &models.URL{}
	err := r.db.QueryRow(query, shortCode).Scan(
		&url.ID,
		&url.OriginalURL,
		&url.ShortCode,
		&url.UserID,
		&url.CreatedAt,
		&url.ExpiresAt, // NEW
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrURLNotFound
		}
		return nil, err
	}

	return url, nil
}

func (r *PostgresURLRepository) FindByOriginalURL(originalURL string) (*models.URL, error) {
	query := `
        SELECT id, original_url, short_code, user_id, created_at, expires_at
        FROM urls
        WHERE original_url = $1
        LIMIT 1
    `

	url := &models.URL{}
	err := r.db.QueryRow(query, originalURL).Scan(
		&url.ID,
		&url.OriginalURL,
		&url.ShortCode,
		&url.UserID,
		&url.CreatedAt,
		&url.ExpiresAt, // NEW
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrURLNotFound
		}
		return nil, err
	}

	return url, nil
}

func (r *PostgresURLRepository) FindByUserID(userID int) ([]*models.URL, error) {
	query := `
        SELECT id, original_url, short_code, user_id, created_at, expires_at
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
			&url.ExpiresAt, // NEW
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

func (r *PostgresURLRepository) DeleteExpired() (int64, error) {
	query := `
		DELETE FROM urls
		WHERE expires_at IS NOT NULL AND expires_at < NOW()
	`
	result, err := r.db.Exec(query)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}
