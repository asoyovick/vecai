// Command api is the VECAI application entry point.
package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"vecai/internal/api"
	"vecai/internal/auth"
	"vecai/internal/boq"
	"vecai/internal/db"
	"vecai/internal/documents"
	"vecai/internal/materials"
	"vecai/internal/newsletter"
	"vecai/internal/project"
	"vecai/internal/reports"
)

type config struct {
	Port          string
	DatabaseURL   string
	MigrationsDir string
	StaticDir     string
	UploadsDir    string
	SeedPath      string
	SeedDemo      bool
}

func loadConfig() config {
	return config{
		Port:          envOr("PORT", "8080"),
		DatabaseURL:   envOr("DATABASE_URL", "postgres://postgres:vecai@localhost:5432/vecai?sslmode=disable"),
		MigrationsDir: envOr("MIGRATIONS_DIR", "db/migrations"),
		StaticDir:     envOr("STATIC_DIR", "dist"),
		UploadsDir:    envOr("UPLOADS_DIR", "data/uploads"),
		SeedPath:      envOr("SEED_PATH", "db/seed.sql"),
		SeedDemo:      os.Getenv("SEED_DEMO") != "0",
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	cfg := loadConfig()

	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer database.Close()

	handler, err := buildApp(cfg, database)
	if err != nil {
		log.Fatalf("startup: %v", err)
	}

	server := &http.Server{Addr: ":" + cfg.Port, Handler: handler, ReadHeaderTimeout: 10 * time.Second}
	log.Printf("vecai api listening on :%s", cfg.Port)
	log.Fatal(server.ListenAndServe())
}

// buildApp wires repositories, services and handlers for every domain.
func buildApp(cfg config, database *db.Pool) (http.Handler, error) {
	if err := db.Migrate(database.DB, cfg.MigrationsDir); err != nil {
		return nil, err
	}
	if cfg.SeedDemo {
		if err := db.SeedIfEmpty(database.DB, cfg.SeedPath); err != nil {
			return nil, err
		}
	}
	if err := os.MkdirAll(cfg.UploadsDir, 0o755); err != nil {
		return nil, err
	}

	authRepo := auth.NewPostgres(database.DB)
	authSvc := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authSvc)

	projectRepo := project.NewPostgres(database.DB)
	projectSvc := project.NewService(projectRepo)
	projectHandler := project.NewHandler(projectSvc)

	materialsRepo := materials.NewPostgres(database.DB)
	materialsSvc := materials.NewService(materialsRepo)
	materialsHandler := materials.NewHandler(materialsSvc)

	documentsRepo := documents.NewPostgres(database.DB)
	documentsSvc := documents.NewService(documentsRepo, cfg.UploadsDir)
	documentsHandler := documents.NewHandler(documentsSvc)

	boqRepo := boq.NewPostgres(database.DB)
	boqSvc := boq.NewService(boqRepo)
	boqHandler := boq.NewHandler(boqSvc)

	reportsRepo := reports.NewPostgres(database.DB)
	reportsSvc := reports.NewService(reportsRepo)
	reportsHandler := reports.NewHandler(reportsSvc)

	newsletterHandler := newsletter.NewHandler(database.DB)

	mux := http.NewServeMux()
	authHandler.Routes(mux)
	projectHandler.Routes(mux)
	materialsHandler.Routes(mux)
	documentsHandler.Routes(mux)
	boqHandler.Routes(mux)
	reportsHandler.Routes(mux)
	newsletterHandler.Routes(mux)
	mux.Handle("/", spaHandler(cfg.StaticDir))

	return api.Recover(api.CORS(api.JSONGuard(auth.Middleware(authSvc)(mux)))), nil
}

// spaHandler serves the built frontend with SPA fallback.
func spaHandler(staticDir string) http.Handler {
	fs := http.FileServer(http.Dir(staticDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(staticDir); os.IsNotExist(err) {
			http.Error(w, "Frontend not built. Run `pnpm build` or use the Vite dev server.", http.StatusServiceUnavailable)
			return
		}
		clean := path.Clean(strings.TrimPrefix(r.URL.Path, "/"))
		if clean != "." {
			if info, err := os.Stat(path.Join(staticDir, clean)); err == nil && !info.IsDir() {
				fs.ServeHTTP(w, r)
				return
			}
		}
		http.ServeFile(w, r, path.Join(staticDir, "index.html"))
	})
}
