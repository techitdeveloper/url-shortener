package services

import (
	"errors"
	"fmt"

	"github.com/techitdeveloper/url-shortener/internal/models"
	"github.com/techitdeveloper/url-shortener/internal/repositories"
	"github.com/techitdeveloper/url-shortener/pkg/utils"
)

var (
	ErrInvalidURL = errors.New("invalid URL")
)

type URLService struct {
	repo    repositories.URLRepository
	baseURL string
}

func NewURLService(repo repositories.URLRepository, baseURL string) *URLService {
	return &URLService{
		repo:    repo,
		baseURL: baseURL,
	}
}

func (s *URLService) ShortenURL(originalURL string) (*models.ShortenResponse, error) {
	if !utils.IsValidURL(originalURL) {
		return nil, ErrInvalidURL
	}

	normalizedURL := utils.NormalizeURL(originalURL)

	existingURL, err := s.repo.FindByOriginalURL(normalizedURL)
	if err == nil {
		return &models.ShortenResponse{
			ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, existingURL.ShortCode),
			OriginalURL: existingURL.OriginalURL,
			ShortCode:   existingURL.ShortCode,
		}, nil
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
	}

	err = s.repo.Save(url)
	if err != nil {
		return nil, err
	}

	return &models.ShortenResponse{
		ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, shortCode),
		OriginalURL: normalizedURL,
		ShortCode:   shortCode,
	}, nil
}

func (s *URLService) GetOriginalURL(shortCode string) (string, error) {
	url, err := s.repo.FindByShortCode(shortCode)
	if err != nil {
		return "", err
	}
	return url.OriginalURL, nil
}
