package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// ⚠️ Автотесты пайплайна. Не меняй их.

func loadOrFail(t *testing.T) []Movie {
	t.Helper()
	movies, err := LoadMovies(moviesJSON)
	if err != nil {
		t.Fatalf("LoadMovies вернул ошибку: %v", err)
	}
	if len(movies) != 5 {
		t.Fatalf("ожидалось 5 фильмов, получено %d", len(movies))
	}
	return movies
}

// toMap кодирует значение в JSON и обратно в map — так проверяем реальные ключи.
func toMap(t *testing.T, v any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	return m
}

func TestLoadMovies_ParsesAllFields(t *testing.T) {
	m := loadOrFail(t)[0]
	out := toMap(t, m)
	if out["title"] != "Inception" {
		t.Errorf("title = %v, ожидалось Inception", out["title"])
	}
	if out["year"] != 2010.0 {
		t.Errorf("year = %v, ожидалось 2010 (поле year должно читаться из JSON и попадать в ответ)", out["year"])
	}
	if out["rating"] != 8.8 {
		t.Errorf("rating = %v, ожидалось 8.8 (читаем из imdb_rating, отдаём как rating)", out["rating"])
	}
	if _, ok := out["imdb_rating"]; ok {
		t.Errorf("в ответе не должно быть ключа imdb_rating — только rating")
	}
	dir, _ := out["director"].(map[string]any)
	if dir["name"] != "Christopher Nolan" {
		t.Errorf("director.name = %v", dir["name"])
	}
}

func TestMovie_HidesInternalNote(t *testing.T) {
	m := Movie{ID: 1, Title: "X", InternalNote: "secret"}
	raw, _ := json.Marshal(m)
	if strings.Contains(string(raw), "secret") || strings.Contains(string(raw), "InternalNote") {
		t.Errorf("InternalNote утёк в JSON: %s", raw)
	}
}

func TestMovie_OmitsEmptyGenres(t *testing.T) {
	movies := loadOrFail(t)
	if _, ok := toMap(t, movies[2])["genres"]; ok {
		t.Errorf("у фильма без жанров ключ genres должен отсутствовать")
	}
	if g, ok := toMap(t, movies[0])["genres"]; !ok || len(g.([]any)) != 2 {
		t.Errorf("у Inception должно быть 2 жанра, получено %v", g)
	}
}

func TestLoadMovies_InvalidJSON(t *testing.T) {
	for _, body := range []string{`[{"id": 1,`, `{"id": "not-a-list"}`, `not json`} {
		if _, err := LoadMovies([]byte(body)); err == nil {
			t.Errorf("для невалидного JSON %q ожидалась ошибка", body)
		}
	}
}

func TestBuildResponse(t *testing.T) {
	movies := loadOrFail(t)
	raw, err := BuildResponse(movies, 8.0)
	if err != nil {
		t.Fatalf("BuildResponse вернул ошибку: %v", err)
	}
	var resp struct {
		Success bool             `json:"success"`
		Data    []map[string]any `json:"data"`
		Meta    struct {
			Total int `json:"total"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("ответ не является валидным JSON: %v\n%s", err, raw)
	}
	if !resp.Success {
		t.Errorf("success должен быть true")
	}
	var titles []string
	for _, m := range resp.Data {
		titles = append(titles, m["title"].(string))
	}
	want := "Inception,Interstellar,Ирония судьбы"
	if strings.Join(titles, ",") != want {
		t.Errorf("фильмы с рейтингом >= 8.0: %v, ожидалось %s", titles, want)
	}
	if resp.Meta.Total != 3 {
		t.Errorf("meta.total = %d, ожидалось 3", resp.Meta.Total)
	}
}

func TestBuildResponse_EmptyIsArray(t *testing.T) {
	raw, err := BuildResponse(loadOrFail(t), 9.9)
	if err != nil {
		t.Fatalf("ошибка: %v", err)
	}
	s := string(raw)
	if !strings.Contains(s, `"data":[]`) {
		t.Errorf("пустой результат должен кодироваться как \"data\":[], получено %s", s)
	}
	if !strings.Contains(s, `"total":0`) {
		t.Errorf("ожидалось \"total\":0, получено %s", s)
	}
}

func TestParseFilter(t *testing.T) {
	f, err := ParseFilter([]byte(`{"min_rating": 8.5, "director": "Christopher Nolan"}`))
	if err != nil {
		t.Fatalf("валидный фильтр вернул ошибку: %v", err)
	}
	if f.MinRating != 8.5 || f.Director != "Christopher Nolan" {
		t.Errorf("фильтр разобран неверно: %+v", f)
	}
	for _, body := range []string{`{"min_ratng": 9}`, `{"min_rating": "high"}`, ``, `{`} {
		if _, err := ParseFilter([]byte(body)); err == nil {
			t.Errorf("для тела %q ожидалась ошибка", body)
		}
	}
}
