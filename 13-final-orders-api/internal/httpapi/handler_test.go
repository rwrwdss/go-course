package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gocourse/13-final-orders-api/internal/domain"
	"gocourse/13-final-orders-api/internal/repo"
	"gocourse/13-final-orders-api/internal/service"
)

// ⚠️ Автотесты пайплайна (end-to-end через настоящий HTTP-сервер). Не меняй их.

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	catalog := repo.NewCatalog([]domain.Product{
		{ID: 1, Name: "Кофе", Price: 159_000},
		{ID: 2, Name: "Кофемолка", Price: 499_000},
	})
	srv := httptest.NewServer(NewHandler(service.NewOrderService(catalog, repo.NewOrderRepo())))
	t.Cleanup(srv.Close)
	return srv
}

type response struct {
	code int
	body map[string]any
	raw  string
}

func send(t *testing.T, method, url, body string) response {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	return response{resp.StatusCode, m, strings.TrimSpace(string(raw))}
}

func TestE2E_OrderLifecycle(t *testing.T) {
	srv := newTestServer(t)

	r := send(t, "POST", srv.URL+"/orders", `{"customer":"anna","items":[{"product_id":1,"quantity":2},{"product_id":2,"quantity":1}]}`)
	if r.code != 201 {
		t.Fatalf("создание заказа: статус %d, тело %s", r.code, r.raw)
	}
	if r.body["total"] != float64(817_000) || r.body["status"] != "new" {
		t.Errorf("созданный заказ: %s", r.raw)
	}

	for _, status := range []string{"paid", "shipped"} {
		r = send(t, "POST", srv.URL+"/orders/1/status", `{"status":"`+status+`"}`)
		if r.code != 200 || r.body["status"] != status {
			t.Fatalf("переход в %s: статус %d, тело %s", status, r.code, r.raw)
		}
	}

	r = send(t, "GET", srv.URL+"/orders/1", "")
	if r.code != 200 || r.body["status"] != "shipped" {
		t.Errorf("GET /orders/1: статус %d, тело %s", r.code, r.raw)
	}
}

func TestE2E_ErrorMapping(t *testing.T) {
	srv := newTestServer(t)
	send(t, "POST", srv.URL+"/orders", `{"customer":"anna","items":[{"product_id":1,"quantity":1}]}`)

	tests := []struct {
		name     string
		method   string
		path     string
		body     string
		wantCode int
	}{
		{"валидация → 422", "POST", "/orders", `{"customer":"","items":[]}`, 422},
		{"неизвестный товар → 404", "POST", "/orders", `{"customer":"bob","items":[{"product_id":42,"quantity":1}]}`, 404},
		{"нет заказа → 404", "GET", "/orders/999", "", 404},
		{"недопустимый переход → 409", "POST", "/orders/1/status", `{"status":"shipped"}`, 409},
		{"битый JSON → 400", "POST", "/orders", `{"customer":`, 400},
		{"кривой id → 400", "GET", "/orders/abc", "", 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := send(t, tt.method, srv.URL+tt.path, tt.body)
			if r.code != tt.wantCode {
				t.Errorf("статус = %d, ожидалось %d; тело %s", r.code, tt.wantCode, r.raw)
			}
			if msg, _ := r.body["error"].(string); msg == "" {
				t.Errorf("в теле ошибки нет поля error: %s", r.raw)
			}
		})
	}
}

// brokenStore имитирует упавшую базу данных.
type brokenStore struct{}

var errDB = errors.New("pq: connection refused to 10.0.0.5:5432 (user=admin)")

func (brokenStore) Create(o domain.Order) domain.Order { return o }
func (brokenStore) Get(int) (domain.Order, error)      { return domain.Order{}, errDB }
func (brokenStore) Update(domain.Order) error          { return errDB }
func (brokenStore) List() []domain.Order               { return nil }

func TestE2E_InternalErrorIsHidden(t *testing.T) {
	srv := httptest.NewServer(NewHandler(service.NewOrderService(repo.NewCatalog(nil), brokenStore{})))
	defer srv.Close()

	r := send(t, "GET", srv.URL+"/orders/1", "")
	if r.code != 500 {
		t.Errorf("статус = %d, ожидалось 500", r.code)
	}
	if r.body["error"] != "internal error" {
		t.Errorf("тело = %s — ожидалось {\"error\":\"internal error\"}", r.raw)
	}
	if strings.Contains(r.raw, "10.0.0.5") || strings.Contains(r.raw, "admin") {
		t.Errorf("🚨 внутренние детали инфраструктуры утекли клиенту: %s", r.raw)
	}
}
