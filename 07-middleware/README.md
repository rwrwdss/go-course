# Модуль 07 · Middleware: логирование, recover, request-id, аутентификация

## 🎯 Цель
Понять, как устроены middleware в Go. Это функции-обёртки над `http.Handler`, на которых держатся логирование, трассировка, авторизация и защита от паник. Плюс научиться передавать данные запроса через `context`.

## 📚 Теория

**Middleware** — функция, которая принимает хендлер и возвращает новый:
```go
type Middleware func(http.Handler) http.Handler

func Logger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ДО хендлера
        next.ServeHTTP(w, r)
        // ПОСЛЕ хендлера
    })
}
```

**Порядок — «луковица»:** `A(B(C(h)))`. Запрос проходит A → B → C → h, ответ возвращается в обратном порядке. Поэтому Recoverer должен стоять **внутри** Logger: тогда Logger увидит статус 500, который выставил Recoverer.

**Прервать цепочку** — значит записать ответ и **не вызывать** `next`. Забытый `return` после `401` — классическая дыра в безопасности.

**Перехват статуса.** У `ResponseWriter` нет метода «какой статус я отправил». Его оборачивают структурой, которая запоминает код. Если хендлер не вызвал `WriteHeader`, Go отправит **200**, значит и по умолчанию надо считать 200.

**`recover()`** работает только внутри `defer` и только в той же горутине:
```go
defer func() {
    if p := recover(); p != nil { /* лог + 500 */ }
}()
```
Паника в горутине, которую запустил хендлер, этим `recover` **не ловится** и роняет весь процесс.

**`context.Value`** — способ передать по цепочке данные **уровня запроса**: request-id, текущего пользователя. Ключ должен быть **приватного типа**, иначе ключи из разных пакетов могут совпасть:
```go
type ctxKey int
const userKey ctxKey = 0
ctx := context.WithValue(r.Context(), userKey, user)
next.ServeHTTP(w, r.WithContext(ctx))
// ...
user, ok := r.Context().Value(userKey).(User)
```
⚠️ Через контекст не передают зависимости (БД, логгер) и обязательные параметры функций.

**401 и 403:** 401 Unauthorized означает «кто ты? (не аутентифицирован)», 403 Forbidden — «я знаю, кто ты, но тебе сюда нельзя».

**X-Request-ID** нужен, чтобы по одному id найти все логи запроса во всех сервисах. Если id уже пришёл от API-gateway, его переиспользуют.

## 🗂 Файлы
| Файл | Что внутри |
|---|---|
| `middleware.go` | **Работаешь здесь** |
| `app.go` | Маршруты и сборка цепочек (менять не нужно) |
| `mock_data.go` | Токены: `token-alice` (admin), `token-bob` (user) |
| `main.go` | Запуск на `:8080` |
| `middleware_test.go` | Автотесты. **Не меняй** |

## 📝 Задания

| # | Что | Тип | Ожидаемое |
|---|---|---|---|
| 1 | `Chain` | 🐞 | `Chain(h, A, B, C)` → вызовы `A → B → C → handler` |
| 2 | `Logger` | 🐞 | для `/public/ping` в логе `status=200` (а не 0) |
| 3 | `Recoverer` | ✍️ | `/panic` → 500 `{"error":"internal server error"}`, сервер жив, в логе `status=500` |
| 4 | `RequestID` + `RequestIDFromContext` | ✍️ | id в заголовке и в контексте совпадают; входящий `X-Request-ID` переиспользуется |
| 5 | `Auth` | 🐞 | без токена → 401 и хендлер **не вызывается** |
| 6 | `RequireRole` | ✍️ | `/admin/stats`: alice → 200, bob → 403, без токена → 401 |

## ▶️ Как запустить
```bash
go run ./tools/ci 07           # 🚦 пайплайн (или make ci M=07)

go run ./07-middleware          # сервер на :8080, логи пишутся в терминал
curl -i localhost:8080/public/ping
curl -i localhost:8080/whoami -H "Authorization: Bearer token-alice" -H "X-Request-ID: my-trace-1"
curl -i localhost:8080/admin/stats -H "Authorization: Bearer token-bob"
curl -i localhost:8080/panic
```

## 💡 Подсказки

<details><summary>Подсказка к заданию 1</summary>

Чтобы первый middleware оказался снаружи, оборачивать надо **с конца**: `for i := len(mws)-1; i >= 0; i--`.
</details>

<details><summary>Подсказка к заданию 2</summary>

Какое значение у `status` до вызова `WriteHeader`? Задай значение по умолчанию при создании `statusRecorder`.
</details>

<details><summary>Подсказка к заданию 3</summary>

`defer func() { if p := recover(); p != nil { ... } }()` поставь **перед** `next.ServeHTTP`.
</details>

<details><summary>Подсказка к заданию 4</summary>

`r.Header.Get("X-Request-ID")` → если пусто, сгенерируй → `context.WithValue(r.Context(), requestIDKey, id)` → `next.ServeHTTP(w, r.WithContext(ctx))`. Достать из контекста: `id, _ := ctx.Value(requestIDKey).(string)`.
</details>

<details><summary>Подсказка к заданию 5</summary>

Что происходит после `writeJSON(... 401 ...)`?
</details>

## ✅ Решения

<details><summary>🔒 Решение задания 1 — Chain</summary>

```go
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
```
</details>

<details><summary>🔒 Решение задания 2 — Logger</summary>

```go
rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
```
</details>

<details><summary>🔒 Решение задания 3 — Recoverer</summary>

```go
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				log.Printf("panic: %v (%s %s)", p, r.Method, r.URL.Path)
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
```
Клиенту не отдаём текст паники: он может раскрыть внутренности сервера. Подробности пишем только в лог.
</details>

<details><summary>🔒 Решение задания 4 — RequestID</summary>

```go
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}
```
</details>

<details><summary>🔒 Решение задания 5 — Auth</summary>

```go
if !ok || !found {
	writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	return // ← без него запрос шёл дальше в хендлер
}
```
</details>

<details><summary>🔒 Решение задания 6 — RequireRole</summary>

```go
func RequireRole(role string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			if user.Role != role {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
```
</details>

## 🧠 Вопросы для самопроверки
1. В каком порядке выполняется код «до» и «после» `next.ServeHTTP` в цепочке из трёх middleware?
2. Почему `recover()` не поймает панику из горутины, запущенной внутри хендлера?
3. Почему ключ контекста нельзя делать обычной строкой?
4. Чем 401 отличается от 403?
