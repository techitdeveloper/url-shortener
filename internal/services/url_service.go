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
	ErrInvalidURL         = errors.New("invalid URL")
	ErrInvalidAlias       = errors.New("invalid custom alias")
	ErrAliasAlreadyExists = errors.New("custom alias already exists")
	ErrInvalidExpiration  = errors.New("invalid expiration date")
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

// Updated ShortenURL with custom alias and expiration support
func (s *URLService) ShortenURL(originalURL string, userID *int, customAlias *string, expiresAtStr *string) (*models.ShortenResponse, error) {
	// Validate URL
	if !utils.IsValidURL(originalURL) {
		return nil, ErrInvalidURL
	}

	// Normalize URL
	normalizedURL := utils.NormalizeURL(originalURL)

	// Parse expiration date if provided
	var expiresAt *time.Time
	if expiresAtStr != nil && *expiresAtStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, *expiresAtStr)
		if err != nil {
			return nil, ErrInvalidExpiration
		}
		// Check if expiration is in the future
		if parsedTime.Before(time.Now()) {
			return nil, errors.New("expiration date must be in the future")
		}
		expiresAt = &parsedTime
	}

	var shortCode string

	// Handle custom alias
	if customAlias != nil && *customAlias != "" {
		// Validate custom alias
		if !utils.IsValidAlias(*customAlias) {
			return nil, ErrInvalidAlias
		}

		// Normalize alias
		shortCode = utils.NormalizeAlias(*customAlias)

		// Check if alias already exists
		_, err := s.repo.FindByShortCode(shortCode)
		if err == nil {
			return nil, ErrAliasAlreadyExists
		}
		if err != repositories.ErrURLNotFound {
			return nil, err
		}
	} else {
		// Generate random short code
		maxAttempts := 5
		for i := 0; i < maxAttempts; i++ {
			shortCode = utils.GenerateShortCode(6)
			_, err := s.repo.FindByShortCode(shortCode)
			if err == repositories.ErrURLNotFound {
				break
			}
		}
	}

	// Create new URL entry
	url := &models.URL{
		OriginalURL: normalizedURL,
		ShortCode:   shortCode,
		UserID:      userID,
		ExpiresAt:   expiresAt,
	}

	err := s.repo.Save(url)
	if err != nil {
		return nil, err
	}

	// Cache the new URL
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cacheKey := s.cache.BuildURLKey(shortCode)
	if err := s.cache.Set(ctx, cacheKey, url.OriginalURL); err != nil {
		log.Printf("Failed to cache new URL: %v", err)
	}

	return &models.ShortenResponse{
		ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, shortCode),
		OriginalURL: normalizedURL,
		ShortCode:   shortCode,
		ExpiresAt:   expiresAt,
	}, nil
}

// GetOriginalURL now checks expiration
func (s *URLService) GetOriginalURL(shortCode string) (string, error) {
	// Try cache first
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cacheKey := s.cache.BuildURLKey(shortCode)
	cachedURL, err := s.cache.Get(ctx, cacheKey)

	if err != nil {
		log.Printf("Cache error: %v", err)
	} else if cachedURL != "" {
		log.Printf("Cache HIT for short code: %s", shortCode)
		return cachedURL, nil
	}

	// Cache miss, query database
	log.Printf("Cache MISS for short code: %s", shortCode)
	url, err := s.repo.FindByShortCode(shortCode)
	if err != nil {
		return "", err
	}

	// Check if URL has expired
	if url.ExpiresAt != nil && url.ExpiresAt.Before(time.Now()) {
		return "", errors.New("URL has expired")
	}

	// Store in cache for next time
	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()

	if err := s.cache.Set(ctx2, cacheKey, url.OriginalURL); err != nil {
		log.Printf("Failed to cache URL after database lookup: %v", err)
	}

	return url.OriginalURL, nil
}

func (s *URLService) GetUserURLs(userID int) ([]*models.URL, error) {
	return s.repo.FindByUserID(userID)
}
