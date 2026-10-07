# Модуль 06 · HTTP-сервер на net/http

## 🎯 Цель
Написать REST-эндпоинты на стандартной библиотеке: маршрутизация Go 1.22+ (метод и `{id}` прямо в шаблоне), path- и query-параметры, статус-коды, заголовки, JSON-ответы. Плюс научиться тестировать хендлеры через `httptest` без запуска сервера.

## 📚 Теория

**Хендлер** — функция `func(w http.ResponseWriter, r *http.Request)` или тип с методом `ServeHTTP`.

**Маршрутизация в Go 1.22+:**
```go
mux := http.NewServeMux()
mux.HandleFunc("GET /books", list)          // только GET
mux.HandleFunc("GET /books/{id}", get)      // path-параметр
mux.HandleFunc("POST /books", create)
// r.PathValue("id") → "42" (строка!)
// на DELETE /books/1 mux сам ответит 405 Method Not Allowed
```

**Query-параметры:** `r.URL.Query().Get("author")` для `/books?author=x`. Если параметра нет, вернётся `""`.

**Порядок записи ответа очень важен:**
```go
w.Header().Set("Content-Type", "application/json") // 1. заголовки
w.WriteHeader(http.StatusCreated)                  // 2. статус (после него заголовки «заморожены»)
json.NewEncoder(w).Encode(v)                       // 3. тело
```
Если вызвать `Header().Set` после `WriteHeader`, заголовок **молча проигнорируется**. Если не вызвать `WriteHeader` вовсе, первый `Write` отправит 200.

**Шпаргалка по статусам:**
| Код | Когда |
|---|---|
| 200 OK | успешный GET |
| 201 Created | ресурс создан (+ заголовок `Location`) |
| 204 No Content | успешно, тела нет (DELETE) |
| 400 Bad Request | кривой запрос: битый JSON, `id=abc` |
| 404 Not Found | ресурса нет |
| 405 Method Not Allowed | метод не поддерживается |
| 409 Conflict | конфликт состояния |
| 422 Unprocessable Entity | JSON валиден, но данные не проходят валидацию |
| 500 Internal Server Error | баг или сбой на сервере |

**Не забывай `return`** после записи ошибки, иначе код пойдёт дальше и допишет второй ответ.

**Конкурентность.** Каждый запрос обрабатывается в **своей горутине**, поэтому общие данные сервера защищены `sync.RWMutex`: `RLock` для чтения, `Lock` для записи.

**`httptest`** позволяет тестировать хендлер без сети:
```go
req := httptest.NewRequest("GET", "/books/1", nil)
rec := httptest.NewRecorder()
srv.ServeHTTP(rec, req)
// rec.Code, rec.Body, rec.Result().Header
```

## 🗂 Файлы
| Файл | Что внутри |
|---|---|
| `server.go` | **Работаешь здесь** |
| `mock_data.go` | 5 книг (Толкин ×2, Булгаков ×2, Donovan) |
| `main.go` | Запуск на `:8080` |
| `server_test.go` | Автотесты. **Не меняй** |

## 📝 Задания

| # | Где | Тип | Ожидаемое |
|---|---|---|---|
| 1 | `writeJSON` | 🐞 | `Content-Type: application/json` у всех ответов |
| 2 | `routes` | 🐞 | `DELETE/PUT/PATCH /books/1` → **405** |
| 3 | `handleListBooks` | ✍️ | `?author=булгаков` → ID `[3,5]`; ничего не найдено → `[]` |
| 4 | `handleGetBook` | 🐞 | `/books/abc` → 400 `{"error":"invalid id"}`; `/books/999` → 404 `{"error":"book not found"}` |
| 5 | `handleCreateBook` | ✍️ | 201 + `Location: /books/6` + книга; битый JSON → 400 `invalid json`; пустые поля → 422 `title and author are required` |

## ▶️ Как запустить
```bash
go run ./tools/ci 06            # 🚦 пайплайн (или make ci M=06)

go run ./06-http-handlers        # запустить сервер на :8080
# в другом терминале:
curl -i localhost:8080/health
curl -s localhost:8080/books
curl -s "localhost:8080/books?author=%D0%B1%D1%83%D0%BB%D0%B3%D0%B0%D0%BA%D0%BE%D0%B2"   # author=булгаков (URL-encoded)
curl -i localhost:8080/books/abc
curl -i -X DELETE localhost:8080/books/1
curl -i -X POST localhost:8080/books -d '{"title":"Чистый код","author":"Роберт Мартин","year":2008}'
```
До исправлений `curl -i localhost:8080/health` показывает `Content-Type: text/plain; charset=utf-8`: Go сам «угадал» тип, потому что наш заголовок потерялся.

## 💡 Подсказки

<details><summary>Подсказка к заданию 1</summary>

Поменяй местами две строки. Сначала заголовки, потом `WriteHeader`.
</details>

<details><summary>Подсказка к заданию 2</summary>

`"GET /books/{id}"`
</details>

<details><summary>Подсказка к заданию 3</summary>

`strings.ToLower` для обеих строк + `strings.Contains`. Результат начни с `result := []Book{}`.
</details>

<details><summary>Подсказка к заданию 4</summary>

`id, err := strconv.Atoi(...)`. При `err != nil` вызови `writeError(w, 400, "invalid id")` и сразу `return`.
</details>

<details><summary>Подсказка к заданию 5</summary>

`json.NewDecoder(r.Body).Decode(&b)` → проверки → `s.mu.Lock()` → присвоить ID, `append`, `nextID++` → `Unlock` → `w.Header().Set("Location", ...)` → `writeJSON(w, 201, b)`.
</details>

## ✅ Решения

<details><summary>🔒 Решение заданий 1 и 2</summary>

```go
func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /books", s.handleListBooks)
	s.mux.HandleFunc("GET /books/{id}", s.handleGetBook)
	s.mux.HandleFunc("POST /books", s.handleCreateBook)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json") // сначала заголовки
	w.WriteHeader(status)                              // потом статус
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}
```
</details>

<details><summary>🔒 Решение задания 3 — фильтр</summary>

```go
func (s *Server) handleListBooks(w http.ResponseWriter, r *http.Request) {
	author := strings.ToLower(r.URL.Query().Get("author"))

	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []Book{}
	for _, b := range s.books {
		if author == "" || strings.Contains(strings.ToLower(b.Author), author) {
			result = append(result, b)
		}
	}
	writeJSON(w, http.StatusOK, result)
}
```
</details>

<details><summary>🔒 Решение задания 4 — GET по id</summary>

```go
func (s *Server) handleGetBook(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, b := range s.books {
		if b.ID == id {
			writeJSON(w, http.StatusOK, b)
			return
		}
	}
	writeError(w, http.StatusNotFound, "book not found")
}
```
</details>

<details><summary>🔒 Решение задания 5 — POST</summary>

```go
func (s *Server) handleCreateBook(w http.ResponseWriter, r *http.Request) {
	var b Book
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if strings.TrimSpace(b.Title) == "" || strings.TrimSpace(b.Author) == "" {
		writeError(w, http.StatusUnprocessableEntity, "title and author are required")
		return
	}

	s.mu.Lock()
	b.ID = s.nextID // ID задаёт сервер, а не клиент
	s.nextID++
	s.books = append(s.books, b)
	s.mu.Unlock()

	w.Header().Set("Location", "/books/"+strconv.Itoa(b.ID))
	writeJSON(w, http.StatusCreated, b)
}
```
</details>

## 🧠 Вопросы для самопроверки
1. В каком порядке нужно вызывать `Header().Set`, `WriteHeader` и `Write`? Почему?
2. Чем 400 отличается от 422?
3. Почему хендлеру нужен мьютекс, хотя мы не запускали ни одной горутины?
4. Почему ID новой книги назначает сервер, а не берёт из тела запроса?
