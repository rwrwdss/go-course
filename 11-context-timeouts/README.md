# Модуль 11 · context: таймауты, отмена, дедлайны

## 🎯 Цель
Научиться ограничивать время работы запросов и прекращать бесполезную работу через `context.Context`. Без этого один медленный внешний сервис «съедает» все горутины и соединения бэкенда, и вся система встаёт.

## 📚 Теория

**`context.Context`** передаётся **первым аргументом** во все функции, которые ходят в сеть, БД или делают долгую работу. Он несёт:
- сигнал отмены (`ctx.Done()` — канал, который закрывается при отмене);
- дедлайн (`ctx.Deadline()`);
- причину отмены (`ctx.Err()`: `context.Canceled` или `context.DeadlineExceeded`);
- значения уровня запроса (`ctx.Value`, см. модуль 07).

**Дерево контекстов.** Дочерний контекст отменяется вместе с родителем:
```go
ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
defer cancel() // ОБЯЗАТЕЛЬНО: освобождает таймер и ресурсы, даже если всё завершилось раньше
```
`go vet` проверяет это сам (анализатор `lostcancel`).

**`r.Context()`** — контекст HTTP-запроса. Go отменяет его, когда клиент разорвал соединение. Если родителем сделать `context.Background()`, сервер будет ждать внешний API впустую, хотя ответ уже никому не нужен.

**Как функция «уважает» контекст:**
```go
select {
case res := <-work:
    return res, nil
case <-ctx.Done():
    return 0, ctx.Err()
}
// или в цикле:
for _, item := range items {
    if err := ctx.Err(); err != nil {
        return err
    }
    ...
}
```
Стандартная библиотека (`http.NewRequestWithContext`, `db.QueryContext`) делает это сама, если передать ей ctx.

**Ошибка → статус:**
| Ошибка | Статус | Смысл |
|---|---|---|
| `context.DeadlineExceeded` | **504 Gateway Timeout** | внешний сервис не успел ответить |
| `context.Canceled` | ответ не пишем (у nginx есть код 499) | клиент ушёл сам |

**Правила:** не храни ctx в структуре, передавай его аргументом. `context.Background()` используется только в `main`, в тестах и в корне фоновых задач.

## 🗂 Файлы
| Файл | Что внутри |
|---|---|
| `handler.go` | **Работаешь здесь**: HTTP-хендлер погоды |
| `batch.go` | **И здесь**: пакетная обработка заказов |
| `mock_weather.go` | Mock внешнего API: moscow 20 мс, london 50 мс, tokyo **2 с** |
| `context_test.go` | Автотесты. **Не меняй** |

## 📝 Задания

| # | Что | Тип | Ожидаемое |
|---|---|---|---|
| 1 | `cancel` в хендлере | 🐞 vet | стадия **vet** зелёная |
| 2 | родительский контекст | 🐞 | клиент ушёл через 30 мс → хендлер завершается сразу, запрос к API прерван |
| 3 | маппинг ошибок | 🐞 | без city → 400, неизвестный город → 404, tokyo с таймаутом 100 мс → **504** `{"error":"upstream timeout"}` |
| 4 | `ProcessOrders` | ✍️ | отмена через ~55 мс → обработано ≈5 заказов и `context.Canceled`; ошибка `process` останавливает обработку |

## ▶️ Как запустить
```bash
go run ./tools/ci 11                     # 🚦 пайплайн (или make ci M=11)

go run ./11-context-timeouts             # сервер на :8080, таймаут к API 500 мс
curl -i "localhost:8080/weather?city=moscow"
curl -i "localhost:8080/weather?city=tokyo"      # → 504 через 0.5 с
curl -i "localhost:8080/weather?city=atlantis"
curl -i --max-time 0.1 "localhost:8080/weather?city=tokyo"  # клиент сам рвёт соединение
```

## 💡 Подсказки

<details><summary>Подсказка к заданию 1</summary>

`ctx, cancel := context.WithTimeout(...)` и следующей строкой `defer cancel()`.
</details>

<details><summary>Подсказка к заданию 2</summary>

Поменяй `context.Background()` на `r.Context()`.
</details>

<details><summary>Подсказка к заданию 3</summary>

Сначала проверь `city == ""`. Потом `switch { case errors.Is(err, ErrCityNotFound): ... case errors.Is(err, context.DeadlineExceeded): ... }`.
</details>

<details><summary>Подсказка к заданию 4</summary>

В начале каждой итерации: `if err := ctx.Err(); err != nil { return processed, err }`. Ошибку `process` тоже проверяй и возвращай.
</details>

## ✅ Решения

<details><summary>🔒 Решение заданий 1–3 — хендлер</summary>

```go
// import "errors"

func (h *WeatherHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	if city == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "city is required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.timeout) // родитель — запрос
	defer cancel()

	temp, err := h.client.Temperature(ctx, city)
	switch {
	case err == nil:
	case errors.Is(err, ErrCityNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "city not found"})
		return
	case errors.Is(err, context.DeadlineExceeded):
		writeJSON(w, http.StatusGatewayTimeout, map[string]string{"error": "upstream timeout"})
		return
	case errors.Is(err, context.Canceled):
		return // клиент ушёл, отвечать некому
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"city": city, "temperature": temp})
}
```
</details>

<details><summary>🔒 Решение задания 4 — ProcessOrders</summary>

```go
func ProcessOrders(ctx context.Context, orderIDs []int, process func(ctx context.Context, id int) error) (int, error) {
	processed := 0
	for _, id := range orderIDs {
		if err := ctx.Err(); err != nil {
			return processed, err
		}
		if err := process(ctx, id); err != nil {
			return processed, fmt.Errorf("process order %d: %w", id, err)
		}
		processed++
	}
	return processed, nil
}
```
</details>

## 🧠 Вопросы для самопроверки
1. Зачем вызывать `cancel()`, если таймаут и так сработает сам?
2. Что произойдёт с дочерним контекстом, если отменить родительский? А наоборот?
3. Чем `context.Canceled` отличается от `context.DeadlineExceeded`?
4. Почему контекст не хранят в полях структуры?
