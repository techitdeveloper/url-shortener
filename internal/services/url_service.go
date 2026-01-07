package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/techitdeveloper/url-shortener/internal/models"
	"github.com/techitdeveloper/url-shortener/internal/repositories"
	"github.com/techitdeveloper/url-shortener/pkg/utils"
)

var (
	ErrInvalidURL = errors.New("invalid URL")
)

type URLService struct {
	repo    repositories.URLRepository
	cache   *CacheService
	baseURL string
}

func NewURLService(repo repositories.URLRepository, cache *CacheService, baseURL string) *URLService {
	return &URLService{
		repo:    repo,
		cache:   cache,
		baseURL: baseURL,
	}
}

func (s *URLService) ShortenURL(originalURL string, userID *int) (*models.ShortenResponse, error) {
	if !utils.IsValidURL(originalURL) {
		return nil, ErrInvalidURL
	}

	normalizedURL := utils.NormalizeURL(originalURL)

	existingURL, err := s.repo.FindByOriginalURL(normalizedURL)
	if err == nil {
		// URL already shortened
		// If it belongs to same user (or no user), return existing
		if (existingURL.UserID == nil && userID == nil) ||
			(existingURL.UserID != nil && userID != nil && *existingURL.UserID == *userID) {
			// Cache it
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			cacheKey := s.cache.BuildURLKey(existingURL.ShortCode)
			if err := s.cache.Set(ctx, cacheKey, existingURL.OriginalURL); err != nil {
				log.Printf("Failed to cache existing URL: %v", err)
			}

			return &models.ShortenResponse{
				ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, existingURL.ShortCode),
				OriginalURL: existingURL.OriginalURL,
				ShortCode:   existingURL.ShortCode,
			}, nil
		}
		// Different user shortened same URL, create new short code
	}

	var shortCode string
	maxAttempts := 5
	for i := 0; i < maxAttempts; i++ {
		shortCode = utils.GenerateShortCode(6)
		_, err := s.repo.FindByShortCode(shortCode)
		if err == repositories.ErrURLNotFound {
			break
		}
	}

	url := &models.URL{
		OriginalURL: normalizedURL,
		ShortCode:   shortCode,
		UserID:      userID,
	}

	err = s.repo.Save(url)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cacheKey := s.cache.BuildURLKey(shortCode)
	if err := s.cache.Set(ctx, cacheKey, url.OriginalURL); err != nil {
		log.Printf("Failed to cache new URL: %v", err)
		// Continue anyway, not critical
	}

	return &models.ShortenResponse{
		ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, shortCode),
		OriginalURL: normalizedURL,
		ShortCode:   shortCode,
	}, nil
}

func (s *URLService) GetOriginalURL(shortCode string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cacheKey := s.cache.BuildURLKey(shortCode)
	cachedURL, err := s.cache.Get(ctx, cacheKey)

	if err != nil {
		log.Printf("Cache error: %v", err)
	} else if cachedURL != "" {
		// Cache hit! Return immediately
		log.Printf("Cache HIT for short code: %s", shortCode)
		return cachedURL, nil
	}

	log.Printf("Cache MISS for short code: %s", shortCode)
	url, err := s.repo.FindByShortCode(shortCode)
	if err != nil {
		return "", err
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()

	if err := s.cache.Set(ctx2, cacheKey, url.OriginalURL); err != nil {
		log.Printf("Failed to cache URL after database lookup: %v", err)
		// Continue anyway, we have the URL
	}

	return url.OriginalURL, nil
}

func (s *URLService) GetUserURLs(userID int) ([]*models.URL, error) {
	return s.repo.FindByUserID(userID)
}
