package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

// Middleware оборачивает обработчик и добавляет поведение до и/или после него.
type Middleware func(http.Handler) http.Handler

// Chain применяет middleware так, чтобы ПЕРВЫЙ в списке был ВНЕШНИМ (выполнялся первым):
//
//	Chain(h, A, B, C)  ==  A(B(C(h)))   → запрос идёт A → B → C → h
//
// 🐞 ЗАДАНИЕ 1: сейчас порядок перевёрнут. Исправь.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for _, mw := range mws {
		h = mw(h)
	}
	return h
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// ---------- Logger ----------

// statusRecorder запоминает статус ответа, чтобы его можно было залогировать.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Logger пишет строку вида: "GET /path status=200 duration=1.2ms".
// 🐞 ЗАДАНИЕ 2: если хендлер не вызывает WriteHeader явно (просто пишет тело),
// в лог попадает status=0, хотя клиент получил 200.
func Logger(l *log.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w}
			next.ServeHTTP(rec, r)
			l.Printf("%s %s status=%d duration=%s", r.Method, r.URL.Path, rec.status, time.Since(start))
		})
	}
}

// ---------- Recoverer ----------

// Recoverer перехватывает панику в хендлере, логирует её и отвечает
// 500 {"error":"internal server error"}, чтобы сервер не упал и клиент получил ответ.
// ✍️ ЗАДАНИЕ 3: реализуй через defer + recover().
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO
		next.ServeHTTP(w, r)
	})
}

// ---------- Request ID ----------

// ctxKey — приватный тип ключа контекста. Так ключи из разных пакетов не столкнутся.
type ctxKey int

const (
	requestIDKey ctxKey = iota
	userKey
)

func newRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// RequestID:
//   - если клиент прислал заголовок X-Request-ID — используем его, иначе генерируем newRequestID();
//   - кладём id в контекст запроса (ключ requestIDKey);
//   - возвращаем id в заголовке ответа X-Request-ID.
//
// ✍️ ЗАДАНИЕ 4: сейчас id только ставится в заголовок и всегда новый. Доделай.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := newRequestID()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}

// RequestIDFromContext достаёт id запроса из контекста ("" если нет).
// ✍️ ЗАДАНИЕ 4 (продолжение).
func RequestIDFromContext(ctx context.Context) string {
	// TODO
	return ""
}

// ---------- Auth ----------

// User — аутентифицированный пользователь.
type User struct {
	ID   int
	Name string
	Role string
}

// UserFromContext достаёт пользователя, которого положил Auth.
func UserFromContext(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userKey).(User)
	return u, ok
}

// Auth проверяет заголовок "Authorization: Bearer <token>".
// Нет заголовка / не Bearer / неизвестный токен → 401 {"error":"unauthorized"} и хендлер НЕ вызывается.
// Иначе кладёт User в контекст и передаёт запрос дальше.
// 🐞 ЗАДАНИЕ 5: неавторизованный запрос всё равно доходит до хендлера! Найди баг.
func Auth(tokens map[string]User) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			user, found := tokens[token]
			if !ok || !found {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			}
			ctx := context.WithValue(r.Context(), userKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole пропускает только пользователей с нужной ролью, иначе 403 {"error":"forbidden"}.
// Используется ПОСЛЕ Auth (пользователь уже в контексте). Если пользователя в контексте нет — 401.
// ✍️ ЗАДАНИЕ 6: реализуй.
func RequireRole(role string) Middleware {
	return func(next http.Handler) http.Handler {
		// TODO
		return next
	}
}
