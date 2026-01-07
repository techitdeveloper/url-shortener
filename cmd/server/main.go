package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/techitdeveloper/url-shortener/database"
	"github.com/techitdeveloper/url-shortener/internal/config"
	"github.com/techitdeveloper/url-shortener/internal/handlers"
	"github.com/techitdeveloper/url-shortener/internal/middleware"
	"github.com/techitdeveloper/url-shortener/internal/repositories"
	"github.com/techitdeveloper/url-shortener/internal/services"
)

func main() {
	cfg := config.GetConfig()

	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	defer db.Close()

	redisClient, err := database.NewRedisClient(&cfg.Redis)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()

	cacheService := services.NewCacheService(redisClient, cfg.Redis.TTL)

	urlRepo := repositories.NewPostgresURLRepository(db)
	userRepo := repositories.NewPostgresUserRepository(db)

	urlService := services.NewURLService(urlRepo, cacheService, cfg.BaseURL)
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)

	urlHandler := handlers.NewURLHandler(urlService)
	authHandler := handlers.NewAuthHandler(authService)
	healthHandler := handlers.NewHealthHandler()

	authMiddleware := middleware.AuthMiddleware(authService)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler.Health)
	mux.HandleFunc("/api/v1/register", authHandler.Register) // NEW
	mux.HandleFunc("/api/v1/login", authHandler.Login)       // NEW
	mux.HandleFunc("/api/v1/shorten", urlHandler.ShortenURL)

	mux.Handle("/api/v1/urls", authMiddleware(http.HandlerFunc(urlHandler.GetMyURLs)))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("URL Shortener API - PostgreSQL Edition"))
			return
		}
		urlHandler.RedirectURL(w, r)
	})

	addr := ":" + cfg.ServerPort
	fmt.Printf("Server starting on port %s...\n", cfg.ServerPort)
	fmt.Printf("Base URL: %s\n", cfg.BaseURL)
	fmt.Println("Using PostgreSQL database")
	fmt.Println("Using Redis cache (TTL:", cfg.Redis.TTL, "seconds)")
	fmt.Println("Authentication enabled with JWT")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}

}
