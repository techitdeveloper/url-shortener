package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/techitdeveloper/url-shortener/internal/models"
	"github.com/techitdeveloper/url-shortener/internal/repositories"
	"github.com/techitdeveloper/url-shortener/internal/services"
)

type URLHandler struct {
	service *services.URLService
}

func NewURLHandler(service *services.URLService) *URLHandler {
	return &URLHandler{
		service: service,
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

	response, err := h.service.ShortenURL(req.URL)
	if err != nil {
		if err == services.ErrInvalidURL {
			respondWithError(w, http.StatusBadRequest, "Invalid URL format")
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

	originalURL, err := h.service.GetOriginalURL(shortCode)
	if err != nil {
		if err == repositories.ErrURLNotFound {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
