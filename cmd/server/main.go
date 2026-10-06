// Command server runs the VECAI web application: the REST API under /api
// plus the built frontend (static files with SPA fallback) when dist/ exists.
package main

import (
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"vecai/internal/config"
	"vecai/internal/db"
	"vecai/internal/handlers"
	"vecai/internal/routes"
)

func main() {
	cfg := config.Load()
	database, err := db.Open(cfg.DatabasePath, cfg.MigrationsDir)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer database.Close()
	if err := os.MkdirAll(cfg.UploadsDir, 0o755); err != nil {
		log.Fatalf("create uploads dir: %v", err)
	}

	h := handlers.New(database, cfg)
	api := routes.New(h)

	mux := http.NewServeMux()
	mux.Handle("/api/", api)
	mux.Handle("/", spaHandler(cfg.StaticDir))

	addr := ":" + cfg.Port
	log.Printf("vecai listening on http://localhost%s (db: %s)", addr, cfg.DatabasePath)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// spaHandler serves the built frontend and falls back to index.html for
// client-side routes. Missing static dir just means the API-only mode
// (the Vite dev server serves the frontend in development).
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
