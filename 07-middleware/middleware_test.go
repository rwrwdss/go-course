package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// ⚠️ Автотесты пайплайна. Не меняй их.

func request(t *testing.T, h http.Handler, method, url string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, url, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func quietLogger() *log.Logger { return log.New(io.Discard, "", 0) }

func TestChain_Order(t *testing.T) {
	var calls []string
	mw := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls = append(calls, name)
				next.ServeHTTP(w, r)
			})
		}
	}
	h := Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, "handler")
	}), mw("A"), mw("B"), mw("C"))
	request(t, h, "GET", "/", nil)
	if want := []string{"A", "B", "C", "handler"}; !reflect.DeepEqual(calls, want) {
		t.Errorf("порядок вызова %v, ожидалось %v", calls, want)
	}
}

func TestLogger_ImplicitStatus200(t *testing.T) {
	var buf bytes.Buffer
	h := NewApp(log.New(&buf, "", 0))
	request(t, h, "GET", "/public/ping", nil)
	if !strings.Contains(buf.String(), "GET /public/ping status=200") {
		t.Errorf("в логе ожидалось 'GET /public/ping status=200', получено: %q", buf.String())
	}
}

func TestRecoverer(t *testing.T) {
	var buf bytes.Buffer
	h := NewApp(log.New(&buf, "", 0))

	var rec *httptest.ResponseRecorder
	func() {
		defer func() {
			if p := recover(); p != nil {
				t.Fatalf("паника дошла до сервера (а значит, процесс упал бы): %v", p)
			}
		}()
		rec = request(t, h, "GET", "/panic", nil)
	}()

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("статус = %d, ожидалось 500", rec.Code)
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["error"] != "internal server error" {
		t.Errorf("тело = %q, ожидалось {\"error\":\"internal server error\"}", rec.Body.String())
	}
	if !strings.Contains(buf.String(), "status=500") {
		t.Errorf("Logger должен залогировать status=500, лог: %q", buf.String())
	}
}

func TestRequestID_GeneratedAndInContext(t *testing.T) {
	rec := request(t, NewApp(quietLogger()), "GET", "/whoami", map[string]string{"Authorization": "Bearer token-alice"})
	headerID := rec.Result().Header.Get("X-Request-ID")
	if headerID == "" {
		t.Fatalf("нет заголовка X-Request-ID в ответе")
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["request_id"] != headerID {
		t.Errorf("request_id из контекста = %q, а в заголовке %q — должны совпадать", body["request_id"], headerID)
	}
}

func TestRequestID_ReusesIncoming(t *testing.T) {
	rec := request(t, NewApp(quietLogger()), "GET", "/whoami", map[string]string{
		"Authorization": "Bearer token-alice",
		"X-Request-ID":  "trace-from-gateway",
	})
	if got := rec.Result().Header.Get("X-Request-ID"); got != "trace-from-gateway" {
		t.Errorf("X-Request-ID = %q, ожидалось переиспользовать входящий trace-from-gateway", got)
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["request_id"] != "trace-from-gateway" {
		t.Errorf("request_id в контексте = %q", body["request_id"])
	}
}

func TestAuth(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		wantCode   int
		wantCalled bool
	}{
		{"без заголовка", "", 401, false},
		{"без Bearer", "token-alice", 401, false},
		{"неизвестный токен", "Bearer token-hacker", 401, false},
		{"валидный токен", "Bearer token-bob", 200, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			h := Auth(mockTokens)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				if _, ok := UserFromContext(r.Context()); !ok {
					t.Errorf("пользователь не попал в контекст")
				}
			}))
			rec := request(t, h, "GET", "/", map[string]string{"Authorization": tt.authHeader})
			if rec.Code != tt.wantCode {
				t.Errorf("статус = %d, ожидалось %d", rec.Code, tt.wantCode)
			}
			if called != tt.wantCalled {
				t.Errorf("хендлер вызван = %v, ожидалось %v (после ответа 401 нужен return!)", called, tt.wantCalled)
			}
		})
	}
}

func TestWhoami(t *testing.T) {
	rec := request(t, NewApp(quietLogger()), "GET", "/whoami", map[string]string{"Authorization": "Bearer token-bob"})
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if rec.Code != 200 || body["user"] != "bob" {
		t.Errorf("GET /whoami: статус %d, тело %s", rec.Code, rec.Body.String())
	}
}

func TestRequireRole(t *testing.T) {
	app := NewApp(quietLogger())
	tests := []struct {
		name     string
		token    string
		wantCode int
	}{
		{"админ", "Bearer token-alice", 200},
		{"обычный пользователь", "Bearer token-bob", 403},
		{"аноним", "", 401},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := request(t, app, "GET", "/admin/stats", map[string]string{"Authorization": tt.token})
			if rec.Code != tt.wantCode {
				t.Errorf("статус = %d, ожидалось %d; тело: %s", rec.Code, tt.wantCode, rec.Body.String())
			}
		})
	}

	t.Run("без Auth в цепочке", func(t *testing.T) {
		h := RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		if rec := request(t, h, "GET", "/", nil); rec.Code != 401 {
			t.Errorf("если пользователя нет в контексте, ожидалось 401, получено %d", rec.Code)
		}
	})
}
