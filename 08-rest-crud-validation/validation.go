package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

const (
	defaultLimit   = 10
	maxLimit       = 50
	minTitleLength = 3
	maxTitleLength = 100
	maxBodyBytes   = 1 << 20 // 1 МБ
)

var errInvalidPagination = errors.New("invalid pagination")

// parsePagination читает ?limit= и ?offset= из query.
//   - limit по умолчанию 10, offset по умолчанию 0;
//   - limit > 50 → обрезать до 50 (не ошибка);
//   - limit < 1, offset < 0 или не число → errInvalidPagination.
//
// ✍️ ЗАДАНИЕ 1: реализуй.
func parsePagination(q url.Values) (limit, offset int, err error) {
	// TODO
	return 0, 0, nil
}

// CreateTaskRequest — тело POST /tasks.
type CreateTaskRequest struct {
	Title    string `json:"title"`
	Priority string `json:"priority"`
}

// Validate возвращает ошибки по полям (пустая мапа — всё ок):
//   - title: после TrimSpace от 3 до 100 СИМВОЛОВ → иначе "must be between 3 and 100 characters";
//   - priority: "", "low", "medium" или "high" → иначе "must be one of: low, medium, high".
//
// 🐞 ЗАДАНИЕ 2: валидация длины некорректна для кириллицы, а priority не проверяется вовсе.
func (r CreateTaskRequest) Validate() map[string]string {
	errs := map[string]string{}
	title := strings.TrimSpace(r.Title)
	if len(title) < minTitleLength || len(title) > maxTitleLength {
		errs["title"] = "must be between 3 and 100 characters"
	}
	return errs
}

// decodeJSON читает тело запроса в dst.
// Требования: тело не больше 1 МБ, НЕИЗВЕСТНЫЕ ПОЛЯ — ошибка.
// ✍️ ЗАДАНИЕ 3: добавь запрет неизвестных полей.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	return dec.Decode(dst)
}
