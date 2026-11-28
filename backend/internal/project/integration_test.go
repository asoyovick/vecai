package project

import (
	"context"
	"os"
	"testing"
	"time"

	"vecai/backend/internal/db"
)

func TestIntegration_ProjectRepository(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping integration test")
	}
	conn, err := db.InitDB(dsn)
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer conn.Close()

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "../../db/migrations"
	}
	if err := db.RunMigrations(conn, migrationsPath); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	repo := NewPostgresRepository(conn)
	ctx := context.Background()
	p := &Project{
		ID:        "itest-" + time.Now().Format("20060102150405.000000000"),
		OwnerID:   "owner-itest",
		Name:      "Integration Project",
		Location:  "Test",
		Status:    "Planning",
		CreatedAt: time.Now().UTC(),
	}
	if err := repo.CreateProject(ctx, p); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	got, err := repo.GetProjectByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetProjectByID: %v", err)
	}
	if got.Name != p.Name || got.OwnerID != p.OwnerID {
		t.Fatalf("unexpected project: %+v", got)
	}
	if err := repo.DeleteProject(ctx, p.ID); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}
}
