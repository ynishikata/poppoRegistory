package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}
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

	// CORS configuration - allow environment variable to override for production
	allowedOrigins := []string{"http://localhost:5173", "http://127.0.0.1:5173"}
	if corsOrigins := os.Getenv("CORS_ORIGINS"); corsOrigins != "" {
		// Support comma-separated list of origins
		origins := strings.Split(corsOrigins, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
		allowedOrigins = origins
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
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
			r.Get("/plushies/{id}", app.HandleGetPlushie)
			r.Put("/plushies/{id}", app.HandleUpdatePlushie)
			r.Put("/plushies/{id}/conversation", app.HandleUpdateConversation)
			r.Post("/plushies/{id}/chat", app.HandleChat)
			r.Delete("/plushies/{id}", app.HandleDeletePlushie)
		})
	})

	// serve uploaded images
	fileServer := http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads")))
	r.Handle("/uploads/*", fileServer)

	// Serve frontend static files in production (optional)
	// If frontend/dist directory exists, serve it
	if _, err := os.Stat("frontend/dist"); err == nil {
		fs := http.FileServer(http.Dir("frontend/dist"))
		r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Don't serve static files for API routes
			if strings.HasPrefix(r.URL.Path, "/api") || strings.HasPrefix(r.URL.Path, "/uploads") {
				http.NotFound(w, r)
				return
			}
			fs.ServeHTTP(w, r)
		}))
	}

	// Port can be set via environment variable (e.g., for production)
	addr := os.Getenv("PORT")
	if addr == "" {
		addr = ":8080"
	} else if !strings.HasPrefix(addr, ":") {
		addr = ":" + addr
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  60 * time.Second,  // Increased for file uploads
		WriteTimeout: 60 * time.Second,  // Increased for file uploads
		IdleTimeout:  120 * time.Second, // Keep connections alive longer
	}

	log.Printf("Server listening on %s", addr)
	log.Printf("CORS allowed origins: %v", allowedOrigins)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
