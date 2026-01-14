package services

import (
	"errors"
	"time"

	"github.com/techitdeveloper/url-shortener/internal/models"
	"github.com/techitdeveloper/url-shortener/internal/repositories"
)

type AnalyticsService struct {
	analyticsRepo repositories.AnalyticsRepository
	urlRepo       repositories.URLRepository
}

func NewAnalyticsService(analyticsRepo repositories.AnalyticsRepository, urlRepo repositories.URLRepository) *AnalyticsService {
	return &AnalyticsService{
		analyticsRepo: analyticsRepo,
		urlRepo:       urlRepo,
	}
}

// RecordClick records a click event
func (s *AnalyticsService) RecordClick(shortCode, ipAddress, userAgent, referer string) error {
	// Find URL by short code
	url, err := s.urlRepo.FindByShortCode(shortCode)
	if err != nil {
		return err
	}

	// TODO: Get country/city from IP (we'll skip this for now, can add GeoIP later)

	click := &models.URLClick{
		URLID:     url.ID,
		ClickedAt: time.Now(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Referer:   referer,
	}

	return s.analyticsRepo.RecordClick(click)
}

// GetURLAnalytics gets analytics for a URL
func (s *AnalyticsService) GetURLAnalytics(shortCode string, days int, userID *int) (*models.URLAnalytics, error) {
	// Find URL
	url, err := s.urlRepo.FindByShortCode(shortCode)
	if err != nil {
		return nil, err
	}

	// Check authorization: only owner can see analytics
	if userID != nil && url.UserID != nil {
		if *userID != *url.UserID {
			return nil, ErrUnauthorized
		}
	} else if url.UserID != nil {
		// URL has owner but request is anonymous
		return nil, ErrUnauthorized
	}

	return s.analyticsRepo.GetURLAnalytics(url.ID, days)
}

var ErrUnauthorized = errors.New("unauthorized to view analytics")
