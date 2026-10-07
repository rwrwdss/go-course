package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

// API — HTTP-слой таск-трекера.
type API struct {
	store *Store
	mux   *http.ServeMux
}

func NewAPI(store *Store) *API {
	a := &API{store: store, mux: http.NewServeMux()}
	a.mux.HandleFunc("GET /tasks", a.listTasks)
	a.mux.HandleFunc("POST /tasks", a.createTask)
	a.mux.HandleFunc("PATCH /tasks/{id}", a.updateTask)
	a.mux.HandleFunc("DELETE /tasks/{id}", a.deleteTask)
	return a
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) { a.mux.ServeHTTP(w, r) }

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func pathID(r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	return id, err == nil && id > 0
}

type listResponse struct {
	Items  []Task `json:"items"`
	Total  int    `json:"total"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// GET /tasks?limit=10&offset=0
func (a *API) listTasks(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePagination(r.URL.Query())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	all := a.store.List()
	start := min(offset, len(all))
	end := min(offset+limit, len(all))
	writeJSON(w, http.StatusOK, listResponse{Items: all[start:end], Total: len(all), Limit: limit, Offset: offset})
}

// POST /tasks
func (a *API) createTask(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}
	if errs := req.Validate(); len(errs) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
		return
	}
	priority := req.Priority
	if priority == "" {
		priority = "medium"
	}
	task := a.store.Create(Task{Title: strings.TrimSpace(req.Title), Priority: priority})
	writeJSON(w, http.StatusCreated, task)
}

// UpdateTaskRequest — тело PATCH /tasks/{id}. Частичное обновление: меняем ТОЛЬКО присланные поля.
// 🐞 ЗАДАНИЕ 4: PATCH {"title":"..."} сбрасывает done в false. Как отличить
// «поле не прислали» от «прислали false»?
type UpdateTaskRequest struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// PATCH /tasks/{id}
//   - нет задачи → 404; если title прислан, он валидируется так же, как при создании → 422.
func (a *API) updateTask(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	task, found := a.store.Get(id)
	if !found {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	var req UpdateTaskRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}
	if req.Title != "" {
		title := strings.TrimSpace(req.Title)
		if n := utf8.RuneCountInString(title); n < minTitleLength || n > maxTitleLength {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
				"errors": map[string]string{"title": "must be between 3 and 100 characters"},
			})
			return
		}
		task.Title = title
	}
	task.Done = req.Done
	a.store.Update(task)
	writeJSON(w, http.StatusOK, task)
}

// DELETE /tasks/{id}
//   - успех → 204 No Content БЕЗ тела; нет задачи → 404.
//
// 🐞 ЗАДАНИЕ 5: сейчас возвращается 200 с телом.
func (a *API) deleteTask(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if !a.store.Delete(id) {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}
