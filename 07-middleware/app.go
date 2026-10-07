package main

import (
	"log"
	"net/http"
)

// NewApp собирает приложение: маршруты + цепочки middleware.
func NewApp(logger *log.Logger) http.Handler {
	mux := http.NewServeMux()
	auth := Auth(mockTokens)

	mux.HandleFunc("GET /public/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("pong")) // WriteHeader не вызываем — Go сам отправит 200
	})

	mux.Handle("GET /whoami", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _ := UserFromContext(r.Context())
		writeJSON(w, http.StatusOK, map[string]string{
			"user":       user.Name,
			"request_id": RequestIDFromContext(r.Context()),
		})
	})))

	mux.Handle("GET /admin/stats", Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]int{"users": len(mockTokens)})
	}), auth, RequireRole("admin")))

	mux.HandleFunc("GET /panic", func(w http.ResponseWriter, r *http.Request) {
		var m map[string]int
		m["boom"] = 1 // panic: assignment to entry in nil map
	})

	// Глобальные middleware: применяются ко всем маршрутам.
	return Chain(mux, RequestID, Logger(logger), Recoverer)
}
