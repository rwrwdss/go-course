package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ⚠️ Автотесты пайплайна. Не меняй их.

// do выполняет запрос к серверу без реальной сети — через httptest.ResponseRecorder.
func do(t *testing.T, srv http.Handler, method, url, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("тело ответа не JSON: %v\n%s", err, rec.Body.String())
	}
	return v
}

func TestHealth_ContentType(t *testing.T) {
	rec := do(t, NewServer(mockBooks()), "GET", "/health", "")
	if rec.Code != 200 {
		t.Errorf("статус = %d, ожидалось 200", rec.Code)
	}
	if ct := rec.Result().Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, ожидалось application/json", ct)
	}
	if body := decode[map[string]string](t, rec); body["status"] != "ok" {
		t.Errorf("тело = %v", body)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	srv := NewServer(mockBooks())
	for _, m := range []string{"DELETE", "PUT", "PATCH"} {
		if rec := do(t, srv, m, "/books/1", ""); rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s /books/1: статус = %d, ожидалось 405", m, rec.Code)
		}
	}
}

func TestListBooks(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantIDs []int
	}{
		{"все книги", "/books", []int{1, 2, 3, 4, 5}},
		{"фильтр без учёта регистра", "/books?author=булгаков", []int{3, 5}},
		{"фильтр по части имени", "/books?author=Толкин", []int{1, 2}},
		{"ничего не найдено", "/books?author=Пушкин", []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, NewServer(mockBooks()), "GET", tt.url, "")
			if rec.Code != 200 {
				t.Fatalf("статус = %d", rec.Code)
			}
			if strings.TrimSpace(rec.Body.String()) == "null" {
				t.Fatalf("тело ответа null — должен быть пустой массив []")
			}
			books := decode[[]Book](t, rec)
			ids := []int{}
			for _, b := range books {
				ids = append(ids, b.ID)
			}
			if len(ids) != len(tt.wantIDs) {
				t.Fatalf("получены ID %v, ожидалось %v", ids, tt.wantIDs)
			}
			for i := range ids {
				if ids[i] != tt.wantIDs[i] {
					t.Fatalf("получены ID %v, ожидалось %v", ids, tt.wantIDs)
				}
			}
		})
	}
}

func TestGetBook(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantCode  int
		wantError string
		wantTitle string
	}{
		{"существующая", "/books/3", 200, "", "Мастер и Маргарита"},
		{"не найдена", "/books/999", 404, "book not found", ""},
		{"id не число", "/books/abc", 400, "invalid id", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, NewServer(mockBooks()), "GET", tt.url, "")
			if rec.Code != tt.wantCode {
				t.Fatalf("статус = %d, ожидалось %d; тело: %s", rec.Code, tt.wantCode, rec.Body.String())
			}
			body := decode[map[string]any](t, rec)
			if tt.wantError != "" && body["error"] != tt.wantError {
				t.Errorf("error = %v, ожидалось %q", body["error"], tt.wantError)
			}
			if tt.wantTitle != "" && body["title"] != tt.wantTitle {
				t.Errorf("title = %v, ожидалось %q", body["title"], tt.wantTitle)
			}
		})
	}
}

func TestCreateBook_OK(t *testing.T) {
	srv := NewServer(mockBooks())
	rec := do(t, srv, "POST", "/books", `{"title":"Чистый код","author":"Роберт Мартин","year":2008}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("статус = %d, ожидалось 201; тело: %s", rec.Code, rec.Body.String())
	}
	if loc := rec.Result().Header.Get("Location"); loc != "/books/6" {
		t.Errorf("Location = %q, ожидалось /books/6", loc)
	}
	created := decode[Book](t, rec)
	if created.ID != 6 || created.Title != "Чистый код" || created.Year != 2008 {
		t.Errorf("создана книга %+v", created)
	}

	// Книга должна реально сохраниться.
	rec = do(t, srv, "GET", "/books/6", "")
	if rec.Code != 200 {
		t.Errorf("после создания GET /books/6 вернул %d", rec.Code)
	}
}

func TestCreateBook_Errors(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantCode int
		wantErr  string
	}{
		{"битый JSON", `{"title":`, 400, "invalid json"},
		{"нет автора", `{"title":"Книга"}`, 422, "title and author are required"},
		{"пустой title", `{"title":"","author":"Кто-то"}`, 422, "title and author are required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, NewServer(mockBooks()), "POST", "/books", tt.body)
			if rec.Code != tt.wantCode {
				t.Fatalf("статус = %d, ожидалось %d", rec.Code, tt.wantCode)
			}
			if body := decode[map[string]string](t, rec); body["error"] != tt.wantErr {
				t.Errorf("error = %q, ожидалось %q", body["error"], tt.wantErr)
			}
		})
	}
}
