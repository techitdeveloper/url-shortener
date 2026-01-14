package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/techitdeveloper/url-shortener/internal/middleware"
	"github.com/techitdeveloper/url-shortener/internal/models"
	"github.com/techitdeveloper/url-shortener/internal/repositories"
	"github.com/techitdeveloper/url-shortener/internal/services"
)

type URLHandler struct {
	urlService       *services.URLService
	analyticsService *services.AnalyticsService // NEW
}

func NewURLHandler(urlService *services.URLService, analyticsService *services.AnalyticsService) *URLHandler {
	return &URLHandler{
		urlService:       urlService,
		analyticsService: analyticsService,
	}
}

// ShortenURL handles POST /api/v1/shorten
func (h *URLHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.URL == "" {
		respondWithError(w, http.StatusBadRequest, "URL is required")
		return
	}

	// Try to get user ID from context
	var userID *int
	if id, ok := middleware.GetUserID(r); ok {
		userID = &id
	}

	// Call service with new parameters
	response, err := h.urlService.ShortenURL(req.URL, userID, req.CustomAlias, req.ExpiresAt)

	if err != nil {
		if err == services.ErrInvalidURL {
			respondWithError(w, http.StatusBadRequest, "Invalid URL format")
			return
		}
		if err == services.ErrInvalidAlias {
			respondWithError(w, http.StatusBadRequest, "Invalid custom alias. Must be 3-30 characters, alphanumeric with hyphens/underscores only")
			return
		}
		if err == services.ErrAliasAlreadyExists {
			respondWithError(w, http.StatusConflict, "Custom alias already taken")
			return
		}
		if err == services.ErrInvalidExpiration {
			respondWithError(w, http.StatusBadRequest, "Invalid expiration date format. Use ISO 8601 format (e.g., 2024-12-31T23:59:59Z)")
			return
		}
		if err.Error() == "expiration date must be in the future" {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to shorten URL")
		return
	}

	respondWithJSON(w, http.StatusCreated, response)
}

func (h *URLHandler) RedirectURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	shortCode := r.URL.Path[1:]

	if shortCode == "" {
		http.Error(w, "Short code is required", http.StatusBadRequest)
		return
	}

	originalURL, err := h.urlService.GetOriginalURL(shortCode)
	if err != nil {
		if err == repositories.ErrURLNotFound {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}
		if err.Error() == "URL has expired" {
			http.Error(w, "URL has expired", http.StatusGone)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Record click analytics (async, don't block redirect)
	go func() {
		ip := getClientIP(r)
		userAgent := r.Header.Get("User-Agent")
		referer := r.Header.Get("Referer")

		if err := h.analyticsService.RecordClick(shortCode, ip, userAgent, referer); err != nil {
			// Log error but don't fail the redirect
			// In production, you'd log this properly
		}
	}()

	http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
}

func (h *URLHandler) GetMyURLs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	urls, err := h.urlService.GetUserURLs(userID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch URLs")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"urls":  urls,
		"count": len(urls),
	})
}

func (h *URLHandler) GetURLAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract short code from path
	// Path is like: /api/v1/urls/abc123/analytics
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	shortCode := parts[4]

	// Get days parameter (default 30)
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 {
			days = parsedDays
		}
	}

	// Get user ID if authenticated
	var userID *int
	if id, ok := middleware.GetUserID(r); ok {
		userID = &id
	}

	analytics, err := h.analyticsService.GetURLAnalytics(shortCode, days, userID)
	if err != nil {
		if err == services.ErrUnauthorized {
			respondWithError(w, http.StatusForbidden, "You don't have permission to view these analytics")
			return
		}
		if err == repositories.ErrURLNotFound {
			respondWithError(w, http.StatusNotFound, "URL not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch analytics")
		return
	}

	respondWithJSON(w, http.StatusOK, analytics)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}
