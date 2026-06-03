package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/CosmiCompile/FlowForge/internal/api"
	"github.com/CosmiCompile/FlowForge/internal/db"
	"github.com/CosmiCompile/FlowForge/internal/migrate"
	"github.com/CosmiCompile/FlowForge/internal/worker"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dsn := os.Getenv("FLOWFORGE_DB_DSN")
	dbConn, err := db.Open(ctx, db.Config{DSN: dsn})
	if err != nil {
		panic(err)
	}
	defer func() { _ = dbConn.Close() }()

	if err := migrate.ApplyAll(ctx, dbConn.SQL); err != nil {
		panic(err)
	}

	// User-first default: single-process mode (server + worker) so "it just works".
	w := worker.New("local", dbConn.SQL)
	go func() {
		_ = w.Run(ctx)
	}()

	apiHandler := api.New(dbConn.SQL)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	// API routes
	r.Mount("/", apiHandler.Router())

	addr := ":8080"
	fmt.Println("flowforge-server listening on", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		panic(err)
	}
}
