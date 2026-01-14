package services

import (
	"log"
	"time"

	"github.com/techitdeveloper/url-shortener/internal/repositories"
)

type CleanupService struct {
	urlRepo repositories.URLRepository
}

func NewCleanupService(urlRepo repositories.URLRepository) *CleanupService {
	return &CleanupService{urlRepo: urlRepo}
}

// StartCleanupJob starts a background job that periodically cleans expired URLs
func (s *CleanupService) StartCleanupJob(interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		log.Println("Cleanup job started")

		for range ticker.C {
			s.cleanupExpiredURLs()
		}
	}()
}

func (s *CleanupService) cleanupExpiredURLs() {
	log.Println("Running cleanup job...")

	count, err := s.urlRepo.DeleteExpired()
	if err != nil {
		log.Printf("Error cleaning up expired URLs: %v", err)
		return
	}

	if count > 0 {
		log.Printf("Cleaned up %d expired URLs", count)
	} else {
		log.Println("No expired URLs to clean up")
	}
}
