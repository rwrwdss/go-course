# Модуль 08 · REST CRUD: валидация, пагинация, PATCH, статус-коды

## 🎯 Цель
Довести CRUD API до «боевого» качества: строгий разбор тела запроса, валидация с ошибками по полям, пагинация с защитой от слишком больших `limit`, частичное обновление (PATCH) через указатели, правильные коды ответа (201 / 204 / 400 / 404 / 422).

## 📚 Теория

**CRUD ↔ HTTP:**
| Операция | Метод | Успех |
|---|---|---|
| Create | `POST /tasks` | 201 + созданный объект |
| Read | `GET /tasks`, `GET /tasks/{id}` | 200 |
| Update (частично) | `PATCH /tasks/{id}` | 200 + обновлённый объект |
| Update (целиком) | `PUT /tasks/{id}` | 200 |
| Delete | `DELETE /tasks/{id}` | **204 No Content**, тела нет |

**Пагинация limit/offset.** Значения по умолчанию плюс **верхняя граница**, иначе клиент запросит `limit=1000000` и положит базу. Вместе с данными возвращаем `total`, чтобы фронт мог нарисовать страницы.

**Строки в Go — это байты в UTF-8.** `len("Ёж") == 4`, потому что каждая кириллическая буква занимает 2 байта. Символы (руны) считают так: `utf8.RuneCountInString(s)`.

**Валидация с ошибками по полям** удобна фронтенду, который подсвечивает конкретные поля формы:
```json
422 {"errors": {"title": "must be between 3 and 100 characters", "priority": "must be one of: low, medium, high"}}
```

**PATCH и zero values.** В Go нельзя отличить «поле не прислали» от «прислали `false`»: в обоих случаях получится `false`. Решение — указатели:
```go
type UpdateTaskRequest struct {
    Title *string `json:"title"` // nil → поле не присылали
    Done  *bool   `json:"done"`  // &false → прислали false
}
```

**Защита при разборе тела:**
```go
r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // не больше 1 МБ
dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields()                     // {"priorty": "high"} → 400, а не тихий игнор
```

**204** — ответ без тела. Вызываешь `w.WriteHeader(http.StatusNoContent)` и ничего не пишешь.

## 🗂 Файлы
| Файл | Что внутри |
|---|---|
| `validation.go` | **Работаешь здесь**: пагинация, валидация, decodeJSON |
| `handlers.go` | **И здесь**: PATCH и DELETE |
| `store.go` | Потокобезопасное хранилище (готово, не трогай) |
| `mock_data.go` | 12 задач |
| `api_test.go` | Автотесты. **Не меняй** |

## 📝 Задания

| # | Где | Тип | Ожидаемое |
|---|---|---|---|
| 1 | `parsePagination` | ✍️ | `/tasks` → 10 штук, `limit=10`, `total=12`; `limit=1000` → `limit=50`; `limit=0`, `limit=abc`, `offset=-1` → **400** |
| 2 | `CreateTaskRequest.Validate` | 🐞 | `"Ёж"` → ошибка title; 100 букв «я» → ок; `priority: "urgent"` → ошибка priority |
| 3 | `decodeJSON` | ✍️ | `{"title":"Задача","priorty":"high"}` → **400** |
| 4 | `UpdateTaskRequest` / `updateTask` | 🐞 | `PATCH {"title":"..."}` не трогает `done`; `PATCH {"done":false}` ставит false; `PATCH {}` ничего не меняет |
| 5 | `deleteTask` | 🐞 | **204** без тела; повторный DELETE → 404 |

## ▶️ Как запустить
```bash
go run ./tools/ci 08            # 🚦 пайплайн (или make ci M=08)

go run ./08-rest-crud-validation
curl -s "localhost:8080/tasks?limit=3&offset=3"
curl -s -X POST localhost:8080/tasks -d '{"title":"Ёж","priority":"urgent"}'
curl -s -X PATCH localhost:8080/tasks/1 -d '{"title":"Настроить CI/CD"}'
curl -i -X DELETE localhost:8080/tasks/3
```

## 💡 Подсказки

<details><summary>Подсказка к заданию 1</summary>

Начни с `limit, offset = 10, 0`. Если параметр в query не пустой, разбери его через `strconv.Atoi` и проверь границы. В конце `min(limit, maxLimit)` (встроенная функция `min` есть с Go 1.21).
</details>

<details><summary>Подсказка к заданию 2</summary>

`utf8.RuneCountInString(strings.TrimSpace(title))`. Для priority удобно написать `switch r.Priority { case "", "low", "medium", "high": default: errs["priority"] = ... }`.
</details>

<details><summary>Подсказка к заданию 4</summary>

Сделай поля указателями. Тогда `if req.Done != nil { task.Done = *req.Done }`. Валидацию title удобно вынести в функцию `validTitle` и использовать и в Validate, и в PATCH.
</details>

<details><summary>Подсказка к заданию 5</summary>

`w.WriteHeader(http.StatusNoContent)`, и всё.
</details>

## ✅ Решения

<details><summary>🔒 Решение задания 1 — parsePagination</summary>

```go
func parsePagination(q url.Values) (limit, offset int, err error) {
	limit, offset = defaultLimit, 0
	if v := q.Get("limit"); v != "" {
		limit, err = strconv.Atoi(v)
		if err != nil || limit < 1 {
			return 0, 0, errInvalidPagination
		}
	}
	if v := q.Get("offset"); v != "" {
		offset, err = strconv.Atoi(v)
		if err != nil || offset < 0 {
			return 0, 0, errInvalidPagination
		}
	}
	return min(limit, maxLimit), offset, nil
}
```
</details>

<details><summary>🔒 Решение заданий 2 и 3 — Validate, validTitle, decodeJSON</summary>

```go
// import "strconv", "unicode/utf8"

func (r CreateTaskRequest) Validate() map[string]string {
	errs := map[string]string{}
	if !validTitle(r.Title) {
		errs["title"] = "must be between 3 and 100 characters"
	}
	switch r.Priority {
	case "", "low", "medium", "high":
	default:
		errs["priority"] = "must be one of: low, medium, high"
	}
	return errs
}

// validTitle проверяет длину в СИМВОЛАХ (рунах), а не в байтах.
func validTitle(title string) bool {
	n := utf8.RuneCountInString(strings.TrimSpace(title))
	return n >= minTitleLength && n <= maxTitleLength
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
```
</details>

<details><summary>🔒 Решение задания 4 — PATCH</summary>

```go
type UpdateTaskRequest struct {
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}

// внутри updateTask, после decodeJSON:
	if req.Title != nil {
		if !validTitle(*req.Title) {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
				"errors": map[string]string{"title": "must be between 3 and 100 characters"},
			})
			return
		}
		task.Title = strings.TrimSpace(*req.Title)
	}
	if req.Done != nil {
		task.Done = *req.Done
	}
	a.store.Update(task)
	writeJSON(w, http.StatusOK, task)
```
Если `unicode/utf8` больше не используется в `handlers.go`, удали импорт, иначе код не скомпилируется.
</details>

<details><summary>🔒 Решение задания 5 — DELETE</summary>

```go
	if !a.store.Delete(id) {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
```
</details>

## 🧠 Вопросы для самопроверки
1. Чем PATCH отличается от PUT?
2. Почему `len(s)` для строки возвращает байты? Как пройтись по строке посимвольно?
3. Какие минусы у пагинации limit/offset на больших таблицах? Что такое cursor (keyset) пагинация?
4. Зачем ограничивать размер тела запроса?
