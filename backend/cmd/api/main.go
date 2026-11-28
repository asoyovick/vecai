package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"vecai/backend/internal/api"
	"vecai/backend/internal/auth"
	"vecai/backend/internal/boq"
	"vecai/backend/internal/db"
	"vecai/backend/internal/documents"
	"vecai/backend/internal/materials"
	"vecai/backend/internal/newsletter"
	"vecai/backend/internal/project"
	"vecai/backend/internal/reports"
	"vecai/backend/internal/storage"
)

func main() {
	port := getEnv("PORT", "8080")
	dbConnStr := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5433/vecai?sslmode=disable")
	migrationsPath := getEnv("MIGRATIONS_PATH", "backend/db/migrations")
	uploadDir := getEnv("UPLOAD_DIR", "backend/uploads")
	jwtSecret := auth.SecretFromEnv()

	log.Println("[INIT] Connecting to PostgreSQL...")
	dbConn, err := db.InitDB(dbConnStr)
	if err != nil {
		log.Fatalf("[FATAL] Could not connect to database: %v", err)
	}
	defer dbConn.Close()

	log.Println("[INIT] Running database migrations...")
	if err := db.RunMigrations(dbConn, migrationsPath); err != nil {
		log.Fatalf("[FATAL] Migration failed: %v", err)
	}

	// ---- Domain modules (modular monolith, docs/architecture.md) ----
	userRepo := auth.NewPostgresUserRepository(dbConn)
	authService := auth.NewService(userRepo)
	authHandler := auth.NewHandler(authService)

	projectRepo := project.NewPostgresRepository(dbConn)
	projectService := project.NewService(projectRepo)
	projectHandler := project.NewHandler(projectService)

	materialsRepo := materials.NewPostgresRepository(dbConn)
	materialsService := materials.NewService(materialsRepo)
	materialsHandler := materials.NewHandler(materialsService)

	reportsRepo := reports.NewPostgresRepository(dbConn)
	reportsService := reports.NewService(reportsRepo)
	reportsHandler := reports.NewHandler(reportsService)

	documentsRepo := documents.NewPostgresRepository(dbConn)
	documentsStorer := storage.NewLocalStorage(uploadDir)
	documentsService := documents.NewService(documentsRepo, documentsStorer)
	documentsHandler := documents.NewHandler(documentsService)

	boqRepo := boq.NewPostgresRepository(dbConn)
	boqService := boq.NewService(boqRepo)
	boqHandler := boq.NewHandler(boqService)

	newsletterRepo := newsletter.NewPostgresRepository(dbConn)
	newsletterHandler := newsletter.NewHandler(newsletterRepo)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(api.CORSMiddleware(allowedOrigins()))

	apiGroup := router.Group("/api/v1")
	apiGroup.Use(auth.OptionalAuthMiddleware(jwtSecret))

	authHandler.RegisterRoutes(apiGroup)
	projectHandler.RegisterRoutes(apiGroup)
	materialsHandler.RegisterRoutes(apiGroup)
	reportsHandler.RegisterRoutes(apiGroup)
	documentsHandler.RegisterRoutes(apiGroup)
	boqHandler.RegisterRoutes(apiGroup)
	newsletterHandler.RegisterRoutes(apiGroup)

	// Protected group for future role-restricted endpoints.
	adminGroup := apiGroup.Group("")
	adminGroup.Use(auth.AuthMiddleware(jwtSecret))
	_ = adminGroup

	router.GET("/health", func(c *gin.Context) {
		api.OK(c, http.StatusOK, "ok", gin.H{"time": time.Now().UTC()})
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	go func() {
		log.Printf("[SERVER] Starting API server on port %s...", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[FATAL] Server startup failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[SERVER] Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("[SERVER] Forced shutdown: %v", err)
	}

	log.Println("[SERVER] Exited cleanly.")
}

// allowedOrigins reads CORS origins from the environment so production can
// lock down access without code changes. Defaults match local development.
func allowedOrigins() []string {
	raw := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:8080")
	origins := strings.Split(raw, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}
	return origins
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
