package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/techitdeveloper/url-shortener/internal/handlers"
	"github.com/techitdeveloper/url-shortener/internal/repositories"
	"github.com/techitdeveloper/url-shortener/internal/services"
)

func main() {
	port := "8080"
	baseURL := "http://localhost:" + port

	urlRepo := repositories.NewInMemoryURLRepository()

	urlService := services.NewURLService(urlRepo, baseURL)

	urlHandler := handlers.NewURLHandler(urlService)
	healthHandler := handlers.NewHealthHandler()

	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler.Health)
	mux.HandleFunc("/api/v1/shorten", urlHandler.ShortenURL)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("URL shorten API"))
			return
		}
		urlHandler.RedirectURL(w, r)
	})

	fmt.Printf("Server starting on port %s...\n", port)
	fmt.Printf("Base URL: %s\n", baseURL)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
