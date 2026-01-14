package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/techitdeveloper/url-shortener/database"
	"github.com/techitdeveloper/url-shortener/internal/config"
	"github.com/techitdeveloper/url-shortener/internal/handlers"
	"github.com/techitdeveloper/url-shortener/internal/middleware"
	"github.com/techitdeveloper/url-shortener/internal/repositories"
	"github.com/techitdeveloper/url-shortener/internal/services"
)

func main() {
	// Load configuration
	cfg := config.GetConfig()

	// Initialize PostgreSQL
	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize Redis
	redisClient, err := database.NewRedisClient(&cfg.Redis)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()

	// Initialize services
	cacheService := services.NewCacheService(redisClient, cfg.Redis.TTL)
	rateLimiter := services.NewRateLimiter(redisClient)

	// Initialize repositories
	urlRepo := repositories.NewPostgresURLRepository(db)
	userRepo := repositories.NewPostgresUserRepository(db)
	analyticsRepo := repositories.NewPostgresAnalyticsRepository(db)

	// Initialize business services
	urlService := services.NewURLService(urlRepo, cacheService, cfg.BaseURL)
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	analyticsService := services.NewAnalyticsService(analyticsRepo, urlRepo)
	cleanupService := services.NewCleanupService(urlRepo)

	// Start background cleanup job (runs every hour)
	cleanupService.StartCleanupJob(1 * time.Hour)

	// Initialize handlers
	urlHandler := handlers.NewURLHandler(urlService, analyticsService)
	authHandler := handlers.NewAuthHandler(authService)
	healthHandler := handlers.NewHealthHandler()

	// Create middlewares
	authMiddleware := middleware.AuthMiddleware(authService)

	var rateLimitMiddleware func(http.Handler) http.Handler
	if cfg.RateLimit.Enabled {
		rateLimitMiddleware = middleware.RateLimitMiddleware(
			rateLimiter,
			cfg.RateLimit.RequestsPerHour,
			1*time.Hour,
		)
	}

	// Setup routes
	mux := http.NewServeMux()

	// Public routes (no auth required)
	mux.HandleFunc("/health", healthHandler.Health)
	mux.HandleFunc("/api/v1/register", authHandler.Register)
	mux.HandleFunc("/api/v1/login", authHandler.Login)

	// Semi-protected route (auth optional, rate limited)
	if rateLimitMiddleware != nil {
		mux.Handle("/api/v1/shorten", rateLimitMiddleware(http.HandlerFunc(urlHandler.ShortenURL)))
	} else {
		mux.HandleFunc("/api/v1/shorten", urlHandler.ShortenURL)
	}

	// Protected routes (auth required)
	mux.Handle("/api/v1/urls", authMiddleware(http.HandlerFunc(urlHandler.GetMyURLs)))

	// Analytics route (auth required for owned URLs)
	mux.HandleFunc("/api/v1/urls/", func(w http.ResponseWriter, r *http.Request) {
		// Check if path ends with /analytics
		if len(r.URL.Path) > 10 && r.URL.Path[len(r.URL.Path)-10:] == "/analytics" {
			urlHandler.GetURLAnalytics(w, r)
			return
		}
		http.Error(w, "Not found", http.StatusNotFound)
	})

	// Redirect route (public, no auth needed)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("URL Shortener API - Production Ready"))
			return
		}
		urlHandler.RedirectURL(w, r)
	})

	// Start server
	addr := ":" + cfg.ServerPort
	fmt.Printf("Server starting on port %s...\n", cfg.ServerPort)
	fmt.Printf("Base URL: %s\n", cfg.BaseURL)
	fmt.Println("Using PostgreSQL database")
	fmt.Println("Using Redis cache (TTL:", cfg.Redis.TTL, "seconds)")
	fmt.Println("Authentication enabled with JWT")
	if cfg.RateLimit.Enabled {
		fmt.Printf("Rate limiting enabled (%d requests/hour)\n", cfg.RateLimit.RequestsPerHour)
	}
	fmt.Println("Background cleanup job started")
	fmt.Println("Analytics tracking enabled")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
