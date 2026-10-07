package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// Book — книга.
type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   int    `json:"year"`
}

// Server — HTTP API библиотеки. Реализует http.Handler.
type Server struct {
	mu     sync.RWMutex // хендлеры вызываются из разных горутин — данные защищаем мьютексом
	books  []Book
	nextID int
	mux    *http.ServeMux
}

// NewServer создаёт сервер и регистрирует маршруты.
func NewServer(seed []Book) *Server {
	s := &Server{books: seed, nextID: len(seed) + 1, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// routes регистрирует маршруты.
// 🐞 ЗАДАНИЕ 2: маршрут "/books/{id}" отвечает на ЛЮБОЙ метод (DELETE, PUT...).
// В Go 1.22+ метод указывается прямо в шаблоне: "GET /books/{id}". Тогда на чужой метод
// сервер сам ответит 405 Method Not Allowed. Поправь шаблоны.
func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /books", s.handleListBooks)
	s.mux.HandleFunc("/books/{id}", s.handleGetBook)
	s.mux.HandleFunc("POST /books", s.handleCreateBook)
}

// writeJSON пишет ответ в формате JSON с нужным статусом.
// 🐞 ЗАДАНИЕ 1: клиент получает Content-Type: text/plain вместо application/json. Почему?
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}

// writeError пишет ошибку в формате {"error": "..."}.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// GET /health → 200 {"status":"ok"}
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GET /books — список книг.
// GET /books?author=булгаков — фильтр по автору: подстрока, БЕЗ учёта регистра.
// Если ничего не найдено — пустой массив [] (не null).
// ✍️ ЗАДАНИЕ 3: добавь фильтрацию по query-параметру author.
func (s *Server) handleListBooks(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	writeJSON(w, http.StatusOK, s.books)
}

// GET /books/{id}
//   - id не число        → 400 {"error":"invalid id"}
//   - книги нет          → 404 {"error":"book not found"}
//   - иначе              → 200 и книга
//
// 🐞 ЗАДАНИЕ 4: сейчас ошибка парсинга игнорируется, а «не найдено» возвращает 200 и null.
func (s *Server) handleGetBook(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))

	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, b := range s.books {
		if b.ID == id {
			writeJSON(w, http.StatusOK, b)
			return
		}
	}
	writeJSON(w, http.StatusOK, nil)
}

// POST /books — создание книги. Тело: {"title": "...", "author": "...", "year": 2020}
//   - невалидный JSON                    → 400 {"error":"invalid json"}
//   - пустой title или author            → 422 {"error":"title and author are required"}
//   - успех → 201, заголовок Location: /books/<id>, в теле созданная книга с новым ID
//
// ✍️ ЗАДАНИЕ 5: реализуй хендлер. Не забудь s.mu.Lock() — ты меняешь общие данные.
func (s *Server) handleCreateBook(w http.ResponseWriter, r *http.Request) {
	// TODO
	writeError(w, http.StatusNotImplemented, "not implemented")
}
