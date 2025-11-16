package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if err := os.MkdirAll("uploads", 0o755); err != nil {
		log.Fatalf("failed to create uploads dir: %v", err)
	}

	db, err := sql.Open("sqlite3", "./poppo.db?_foreign_keys=on")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := migrate(db); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	sessionStore := NewSessionStore()
	app := &App{
		DB:           db,
		SessionStore: sessionStore,
	}

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api", func(r chi.Router) {
		r.Post("/register", app.HandleRegister)
		r.Post("/login", app.HandleLogin)
		r.Post("/logout", app.HandleLogout)

		r.Group(func(r chi.Router) {
			r.Use(app.AuthMiddleware)
			r.Get("/me", app.HandleMe)

			r.Get("/plushies", app.HandleListPlushies)
			r.Post("/plushies", app.HandleCreatePlushie)
			r.Put("/plushies/{id}", app.HandleUpdatePlushie)
			r.Delete("/plushies/{id}", app.HandleDeletePlushie)
		})
	})

	// serve uploaded images
	fileServer := http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads")))
	r.Handle("/uploads/*", fileServer)

	addr := ":8080"
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("Server listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}


