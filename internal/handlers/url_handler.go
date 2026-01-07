package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/techitdeveloper/url-shortener/internal/middleware"
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
	log.Printf("ShortenURL: Received request: %s %s", r.Method, r.URL.Path)

	if r.Method != http.MethodPost {
		log.Printf("ShortenURL: Method not allowed: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ShortenURL: Invalid request body: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	log.Printf("ShortenURL: Request body: %+v", req)

	if req.URL == "" {
		log.Printf("ShortenURL: URL is empty")
		respondWithError(w, http.StatusBadRequest, "URL is required")
		return
	}

	var userID *int
	if id, ok := middleware.GetUserID(r); ok {
		userID = &id
		log.Printf("ShortenURL: User ID: %d", id)
	} else {
		log.Printf("ShortenURL: No user ID found in request")
	}

	response, err := h.service.ShortenURL(req.URL, userID)
	if err != nil {
		log.Printf("ShortenURL: Failed to shorten URL: %v", err)
		if err == services.ErrInvalidURL {
			respondWithError(w, http.StatusBadRequest, "Invalid URL format")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to shorten URL")
		return
	}

	log.Printf("ShortenURL: Success. Response: %+v", response)
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

	urls, err := h.service.GetUserURLs(userID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch URLs")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"urls":  urls,
		"count": len(urls),
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
