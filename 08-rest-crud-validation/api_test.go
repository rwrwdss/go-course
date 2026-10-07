package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ⚠️ Автотесты пайплайна. Не меняй их.

func call(t *testing.T, h http.Handler, method, url, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func newAPI() *API { return NewAPI(NewStore(mockTasks())) }

func TestList_Pagination(t *testing.T) {
	tests := []struct {
		name                 string
		url                  string
		wantIDs              []int
		wantLimit, wantTotal int
	}{
		{"значения по умолчанию", "/tasks", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 10, 12},
		{"limit и offset", "/tasks?limit=3&offset=3", []int{4, 5, 6}, 3, 12},
		{"последняя неполная страница", "/tasks?limit=5&offset=10", []int{11, 12}, 5, 12},
		{"offset за пределами", "/tasks?offset=100", []int{}, 10, 12},
		{"limit обрезается до 50", "/tasks?limit=1000", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, 50, 12},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := call(t, newAPI(), "GET", tt.url, "")
			if rec.Code != 200 {
				t.Fatalf("статус = %d, тело %s", rec.Code, rec.Body.String())
			}
			var resp listResponse
			_ = json.Unmarshal(rec.Body.Bytes(), &resp)
			ids := []int{}
			for _, task := range resp.Items {
				ids = append(ids, task.ID)
			}
			if len(ids) != len(tt.wantIDs) || (len(ids) > 0 && (ids[0] != tt.wantIDs[0] || ids[len(ids)-1] != tt.wantIDs[len(tt.wantIDs)-1])) {
				t.Errorf("ID = %v, ожидалось %v", ids, tt.wantIDs)
			}
			if resp.Limit != tt.wantLimit || resp.Total != tt.wantTotal {
				t.Errorf("limit=%d total=%d, ожидалось limit=%d total=%d", resp.Limit, resp.Total, tt.wantLimit, tt.wantTotal)
			}
		})
	}
}

func TestList_InvalidPagination(t *testing.T) {
	for _, q := range []string{"limit=0", "limit=-5", "limit=abc", "offset=-1", "offset=x"} {
		t.Run(q, func(t *testing.T) {
			if rec := call(t, newAPI(), "GET", "/tasks?"+q, ""); rec.Code != 400 {
				t.Errorf("GET /tasks?%s: статус = %d, ожидалось 400", q, rec.Code)
			}
		})
	}
}

func TestCreateTaskRequest_Validate(t *testing.T) {
	tests := []struct {
		name       string
		req        CreateTaskRequest
		wantFields []string
	}{
		{"валидный", CreateTaskRequest{Title: "Купить молоко", Priority: "low"}, nil},
		{"без приоритета — ок", CreateTaskRequest{Title: "Купить молоко"}, nil},
		{"2 кириллических символа (4 байта)", CreateTaskRequest{Title: "Ёж"}, []string{"title"}},
		{"пробелы не считаются", CreateTaskRequest{Title: "   ab   "}, []string{"title"}},
		{"100 кириллических символов (200 байт)", CreateTaskRequest{Title: strings.Repeat("я", 100)}, nil},
		{"101 символ", CreateTaskRequest{Title: strings.Repeat("я", 101)}, []string{"title"}},
		{"неизвестный приоритет", CreateTaskRequest{Title: "Задача", Priority: "urgent"}, []string{"priority"}},
		{"обе ошибки", CreateTaskRequest{Title: "", Priority: "asap"}, []string{"title", "priority"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.req.Validate()
			if len(errs) != len(tt.wantFields) {
				t.Fatalf("ошибки валидации %v, ожидались поля %v", errs, tt.wantFields)
			}
			for _, f := range tt.wantFields {
				if _, ok := errs[f]; !ok {
					t.Errorf("нет ошибки для поля %q: %v", f, errs)
				}
			}
		})
	}
}

func TestCreate(t *testing.T) {
	api := newAPI()
	rec := call(t, api, "POST", "/tasks", `{"title":"  Новая задача  "}`)
	if rec.Code != 201 {
		t.Fatalf("статус = %d, тело %s", rec.Code, rec.Body.String())
	}
	var task Task
	_ = json.Unmarshal(rec.Body.Bytes(), &task)
	if task.ID != 13 || task.Title != "Новая задача" || task.Priority != "medium" {
		t.Errorf("создана задача %+v", task)
	}

	rec = call(t, api, "POST", "/tasks", `{"title":"Ок","priority":"urgent"}`)
	if rec.Code != 422 {
		t.Fatalf("невалидная задача: статус = %d, ожидалось 422", rec.Code)
	}
	var body struct {
		Errors map[string]string `json:"errors"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if len(body.Errors) != 2 {
		t.Errorf("ожидались ошибки по title и priority, получено %v", body.Errors)
	}
}

func TestCreate_UnknownField(t *testing.T) {
	rec := call(t, newAPI(), "POST", "/tasks", `{"title":"Задача","priorty":"high"}`)
	if rec.Code != 400 {
		t.Errorf("неизвестное поле priorty: статус = %d, ожидалось 400", rec.Code)
	}
}

func TestPatch_PartialUpdate(t *testing.T) {
	api := newAPI()

	// Задача 1 выполнена (done=true). Меняем только title — done должен остаться true.
	rec := call(t, api, "PATCH", "/tasks/1", `{"title":"Настроить CI/CD"}`)
	var task Task
	_ = json.Unmarshal(rec.Body.Bytes(), &task)
	if rec.Code != 200 || task.Title != "Настроить CI/CD" || !task.Done {
		t.Errorf("PATCH title: статус %d, задача %+v — done должен остаться true", rec.Code, task)
	}

	// Явно присланный false должен примениться.
	rec = call(t, api, "PATCH", "/tasks/1", `{"done":false}`)
	_ = json.Unmarshal(rec.Body.Bytes(), &task)
	if task.Done || task.Title != "Настроить CI/CD" {
		t.Errorf("PATCH done=false: задача %+v", task)
	}

	// Пустой PATCH ничего не меняет.
	rec = call(t, api, "PATCH", "/tasks/8", `{}`)
	_ = json.Unmarshal(rec.Body.Bytes(), &task)
	if !task.Done || task.Title != "Ревью PR #42" {
		t.Errorf("пустой PATCH изменил задачу: %+v", task)
	}
}

func TestPatch_Errors(t *testing.T) {
	api := newAPI()
	if rec := call(t, api, "PATCH", "/tasks/999", `{"done":true}`); rec.Code != 404 {
		t.Errorf("несуществующая задача: статус = %d, ожидалось 404", rec.Code)
	}
	if rec := call(t, api, "PATCH", "/tasks/2", `{"title":"ab"}`); rec.Code != 422 {
		t.Errorf("короткий title: статус = %d, ожидалось 422", rec.Code)
	}
}

func TestDelete(t *testing.T) {
	api := newAPI()
	rec := call(t, api, "DELETE", "/tasks/3", "")
	if rec.Code != http.StatusNoContent {
		t.Errorf("статус = %d, ожидалось 204", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("у 204 не должно быть тела, получено %q", rec.Body.String())
	}
	if rec := call(t, api, "DELETE", "/tasks/3", ""); rec.Code != 404 {
		t.Errorf("повторное удаление: статус = %d, ожидалось 404", rec.Code)
	}
}
