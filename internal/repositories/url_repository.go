package repositories

import (
	"errors"
	"sync"
	"time"

	"github.com/techitdeveloper/url-shortener/internal/models"
)

var (
	ErrURLNotFound     = errors.New("url not found")
	ErrShortCodeExists = errors.New("short code already exists")
)

type URLRepository interface {
	Save(url *models.URL) error
	FindByShortCode(shortCode string) (*models.URL, error)
	FindByOriginalURL(originalURL string) (*models.URL, error)
}

type InMemoryURLRepository struct {
	urls map[string]*models.URL // shortCode -> URL
	mu   sync.RWMutex
}

func NewInMemoryURLRepository() *InMemoryURLRepository {
	return &InMemoryURLRepository{
		urls: make(map[string]*models.URL),
	}
}

func (r *InMemoryURLRepository) Save(url *models.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.urls[url.ShortCode]; exists {
		return ErrShortCodeExists
	}

	url.CreatedAt = time.Now()
	r.urls[url.ShortCode] = url

	return nil
}

func (r *InMemoryURLRepository) FindByShortCode(shortCode string) (*models.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	url, exists := r.urls[shortCode]
	if !exists {
		return nil, ErrURLNotFound
	}

	return url, nil
}

func (r *InMemoryURLRepository) FindByOriginalURL(originalURL string) (*models.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, url := range r.urls {
		if url.OriginalURL == originalURL {
			return url, nil
		}
	}

	return nil, ErrURLNotFound
}
