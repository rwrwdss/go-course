# Модуль 03 · Обработка ошибок

## 🎯 Цель
Освоить идиоматичную обработку ошибок в Go: sentinel-ошибки, свои типы ошибок, обёртывание через `%w`, `errors.Is` / `errors.As`. Плюс научиться превращать ошибки бизнес-логики в HTTP-статусы. На бэкенде этим занимается каждый хендлер.

## 📚 Теория

**Ошибка в Go — это обычное значение**, которое реализует интерфейс:
```go
type error interface { Error() string }
```
Исключений нет. Функция возвращает `error` последним значением, а вызывающий код **обязан** её проверить:
```go
if err := doSomething(); err != nil {
    return fmt.Errorf("doing something: %w", err)
}
```
`_ = doSomething()` — проглоченная ошибка. Это один из самых опасных багов.

**Sentinel-ошибка** — заранее созданное значение, с которым сравнивают:
```go
var ErrNotFound = errors.New("not found")
```

**Обёртывание `%w`** добавляет контекст и сохраняет исходную ошибку внутри:
```go
err := fmt.Errorf("get user 42: %w", ErrNotFound)
err.Error()                    // "get user 42: not found"
errors.Is(err, ErrNotFound)    // true — Is разворачивает цепочку
err == ErrNotFound             // false! Поэтому сравнивай через errors.Is
```
С `%v` вместо `%w` текст будет тот же, но цепочка разорвётся и `errors.Is` вернёт false.

**Свой тип ошибки** нужен, когда в ошибке надо передать данные (поле, код и т.п.):
```go
type ValidationError struct{ Field, Message string }
func (e *ValidationError) Error() string { ... }

var vErr *ValidationError
if errors.As(err, &vErr) {   // As ищет в цепочке ошибку нужного ТИПА
    fmt.Println(vErr.Field)
}
```

**Атомарность операций.** Сначала проверь всё, что может пойти не так, и только потом меняй состояние. Иначе после ошибки данные останутся в полуизменённом виде (в нашем банке так сгорают деньги). В реальной БД эту задачу решают транзакции.

**Ошибка → HTTP-статус.** Слой бизнес-логики ничего не знает про HTTP и возвращает доменные ошибки, а хендлер маппит их в коды 400, 404, 409, 500.

## 🗂 Файлы
| Файл | Что внутри |
|---|---|
| `bank.go` | **Здесь ты работаешь** |
| `mock_data.go` | 3 счёта: acc-1 (100 000 коп.), acc-2 (5 000), acc-3 (0) |
| `main.go` | Демо-сценарии |
| `bank_test.go` | Автотесты. **Не меняй** |

## 📝 Задания

| # | Что | Ожидаемое поведение |
|---|---|---|
| 1 | ✍️ `ValidationError.Error()` | `"validation failed: amount: must be positive"` |
| 2 | 🐞 `Withdraw` | `amount<=0` → `*ValidationError{Field:"amount"}`; нет счёта → `errors.Is(err, ErrAccountNotFound)`; мало денег → `errors.Is(err, ErrInsufficientFunds)`, в тексте есть id счёта |
| 3 | 🐞 `Transfer` | при любой ошибке балансы не меняются; ошибка возвращается наверх; `from == to` → `*ValidationError{Field:"to"}` |
| 4 | ✍️ `HTTPStatus` | nil→200, Validation→400, NotFound→404, Insufficient→409, прочее→500 (с учётом обёрток) |

Ожидаемый вывод `go run ./03-errors` после решения:
```
Withdraw acc-1 1000                      → err=<nil>, http=200
Withdraw acc-2 999999                    → err=withdraw acc-2: insufficient funds, http=409
Withdraw acc-404 10                      → err=withdraw acc-404: account not found, http=404
Withdraw acc-1 -5                        → err=validation failed: amount: must be positive, http=400
Transfer acc-1 → acc-404 500             → err=transfer to acc-404: account not found, http=404
Transfer acc-3 → acc-1 500 (пустой счёт) → err=transfer: withdraw acc-3: insufficient funds, http=409
Баланс acc-1: 99000 коп.
Баланс acc-2: 5000 коп.
Баланс acc-3: 0 коп.
```
(Текст ошибок может отличаться. Тесты проверяют `errors.Is`/`errors.As`, а не точные строки.)

## ▶️ Как запустить
```bash
go run ./03-errors
go run ./tools/ci 03      # или: make ci M=03
```

## 💡 Подсказки

<details><summary>Подсказка к заданию 1</summary>

`fmt.Sprintf("validation failed: %s: %s", e.Field, e.Message)`
</details>

<details><summary>Подсказка к заданию 2</summary>

Не создавай новую ошибку через `errors.New("account not found")`: это **другое значение**, хоть и с тем же текстом. Оборачивай sentinel: `fmt.Errorf("withdraw %s: %w", id, ErrAccountNotFound)`. Проверку `amount` делай самой первой.
</details>

<details><summary>Подсказка к заданию 3</summary>

Порядок: 1) `from == to`? 2) получатель существует? 3) `Withdraw` и проверка ошибки. 4) `Deposit`. После шага 2 `Deposit` уже не упадёт из-за отсутствия счёта.
</details>

<details><summary>Подсказка к заданию 4</summary>

`switch { case errors.As(err, &vErr): ... case errors.Is(err, ErrAccountNotFound): ... }`. Сначала обработай `err == nil`.
</details>

## ✅ Решения

<details><summary>🔒 Решение задания 1</summary>

```go
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %s: %s", e.Field, e.Message)
}
```
</details>

<details><summary>🔒 Решение задания 2 — Withdraw</summary>

```go
func (b *Bank) Withdraw(id string, amount int64) error {
	if amount <= 0 {
		return &ValidationError{Field: "amount", Message: "must be positive"}
	}
	acc, ok := b.accounts[id]
	if !ok {
		return fmt.Errorf("withdraw %s: %w", id, ErrAccountNotFound)
	}
	if acc.Balance < amount {
		return fmt.Errorf("withdraw %s: %w", id, ErrInsufficientFunds)
	}
	acc.Balance -= amount
	return nil
}
```
</details>

<details><summary>🔒 Решение задания 3 — Transfer</summary>

```go
func (b *Bank) Transfer(from, to string, amount int64) error {
	if from == to {
		return &ValidationError{Field: "to", Message: "cannot transfer to the same account"}
	}
	if _, ok := b.accounts[to]; !ok { // проверяем ДО списания
		return fmt.Errorf("transfer to %s: %w", to, ErrAccountNotFound)
	}
	if err := b.Withdraw(from, amount); err != nil { // ошибку НЕ игнорируем
		return fmt.Errorf("transfer: %w", err)
	}
	if err := b.Deposit(to, amount); err != nil {
		return fmt.Errorf("transfer: %w", err)
	}
	return nil
}
```
</details>

<details><summary>🔒 Решение задания 4 — HTTPStatus</summary>

```go
func HTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	var vErr *ValidationError
	switch {
	case errors.As(err, &vErr):
		return http.StatusBadRequest
	case errors.Is(err, ErrAccountNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrInsufficientFunds):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
```
</details>

## 🧠 Вопросы для самопроверки
1. Чем `%w` отличается от `%v` в `fmt.Errorf`?
2. Когда использовать `errors.Is`, а когда `errors.As`?
3. Почему метод `Error()` объявлен на `*ValidationError`, а не на `ValidationError`? Что будет с `errors.As`, если перепутать?
4. Почему не стоит отдавать клиенту текст внутренней ошибки (например, ошибки БД) в ответе 500?
