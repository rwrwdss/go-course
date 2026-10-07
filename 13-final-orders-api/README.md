# Модуль 13 · Итоговый проект: Orders API (слоистая архитектура + graceful shutdown)

## 🎯 Цель
Собрать всё из курса в одном маленьком, но настоящем сервисе: слои `domain → repo → service → httpapi`, доменные ошибки и их маппинг в HTTP, потокобезопасный репозиторий, бизнес-правила (машина состояний заказа), end-to-end тесты и **graceful shutdown**.

## 📚 Теория

**Структура проекта** (упрощённый [standard Go project layout](https://github.com/golang-standards/project-layout)):
```
13-final-orders-api/
├── main.go               ← точка входа: сигналы ОС, запуск
├── app.go                ← composition root: собираем слои, run() с graceful shutdown
├── mock_data.go          ← каталог товаров
└── internal/             ← код, который НЕЛЬЗЯ импортировать из других модулей (правило Go)
    ├── domain/           ← сущности и доменные ошибки. Ни от чего не зависит
    ├── repo/             ← хранение данных (сейчас in-memory, завтра PostgreSQL)
    ├── service/          ← бизнес-правила. Зависит от ИНТЕРФЕЙСОВ хранилищ
    └── httpapi/          ← HTTP: JSON ↔ вызовы сервиса, ошибки → статусы
```
**Направление зависимостей:** `httpapi → service → domain ← repo`. Сервис ничего не знает про HTTP, а домен ничего не знает ни о ком.

**Машина состояний.** Допустимые переходы статусов удобно описать данными (мапа «откуда → куда»), а не цепочкой `if`:
```
new ──► paid ──► shipped
 │        │
 └────────┴────► cancelled
```

**Внутренние ошибки нельзя показывать клиенту.** Текст вроде `pq: connection refused to 10.0.0.5:5432 (user=admin)` раскрывает инфраструктуру. Клиент получает `500 {"error":"internal error"}`, а детали пишутся в лог.

**Graceful shutdown.** Kubernetes или Docker при деплое шлют процессу `SIGTERM`. Сервис должен:
1. перестать принимать новые соединения;
2. дождаться завершения активных запросов (с таймаутом);
3. корректно выйти.

```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
// ...
go srv.ListenAndServe()   // вернёт http.ErrServerClosed после Shutdown — это НЕ ошибка
<-ctx.Done()
srv.Shutdown(shutdownCtx) // ждёт активные запросы
```

**Деньги в копейках (`int64`)**, а цена фиксируется в позиции заказа: если цена в каталоге поменяется, старый заказ не изменится.

## 📝 Задания

| # | Где | Тип | Ожидаемое |
|---|---|---|---|
| 1 | `internal/repo/memory.go` | 🐞 гонка | 20 горутин × 50 заказов → 1000 уникальных ID, `-race` молчит |
| 2 | `internal/service/orders.go` → `CreateOrder` | ✍️ | валидация (`ErrValidation`), неизвестный товар (`ErrNotFound`), цена из каталога, `Total`, статус `new`; при ошибке заказ **не сохраняется** |
| 3 | `internal/service/orders.go` → `ChangeStatus` | 🐞 | только переходы из `allowedTransitions`, иначе `ErrInvalidTransition`, статус не меняется |
| 4 | `internal/httpapi/handler.go` → `writeServiceError` | 🐞 | 422 / 404 / 409 с текстом ошибки; всё прочее → 500 `{"error":"internal error"}` без утечки деталей |
| 5 | `app.go` → `run` | 🐞 | после отмены ctx сервер останавливается и `run` возвращает `nil` (< 2 с); занятый порт → ошибка |

Сценарий, который проверяет e2e-тест:
```
POST /orders {"customer":"anna","items":[{"product_id":1,"quantity":2},{"product_id":2,"quantity":1}]}
  → 201 {"id":1,"customer":"anna","items":[...],"total":817000,"status":"new"}
POST /orders/1/status {"status":"paid"}     → 200
POST /orders/1/status {"status":"shipped"}  → 200
GET  /orders/1                               → 200 {"status":"shipped",...}
```

## ▶️ Как запустить
```bash
go run ./tools/ci 13                 # 🚦 пайплайн (или make ci M=13)

go run ./13-final-orders-api          # сервер на :8080
curl -s -X POST localhost:8080/orders -d '{"customer":"anna","items":[{"product_id":1,"quantity":2},{"product_id":4,"quantity":1}]}'
curl -s -X POST localhost:8080/orders/1/status -d '{"status":"paid"}'
curl -s -X POST localhost:8080/orders/1/status -d '{"status":"new"}'      # 409
curl -s localhost:8080/orders
# остановить: Ctrl+C → после задания 5 увидишь «👋 сервер остановлен корректно»
```

## 💡 Подсказки

<details><summary>Подсказка к заданию 1</summary>

Тот же приём, что в модуле 09: поле `mu sync.RWMutex`, `Lock` в `Create` и `Update`, `RLock` в `Get` и `List`.
</details>

<details><summary>Подсказка к заданию 2</summary>

Сначала **всё проверь** (покупатель, позиции, количество, товары), собери заказ в локальной переменной, и только в конце вызови `s.store.Create`. Ошибки оборачивай: `fmt.Errorf("customer is required: %w", domain.ErrValidation)`.
</details>

<details><summary>Подсказка к заданию 3</summary>

`slices.Contains(allowedTransitions[order.Status], to)`. У `shipped` и `cancelled` в мапе нет ключа, значит, из них переходов нет (nil-слайс).
</details>

<details><summary>Подсказка к заданию 4</summary>

`switch { case errors.Is(err, domain.ErrValidation): ... }`. Для 500 сначала `log.Printf` с деталями, потом клиенту `"internal error"`.
</details>

<details><summary>Подсказка к заданию 5</summary>

Запусти `ListenAndServe` в горутине, ошибку (кроме `http.ErrServerClosed`) отправь в канал. Потом `select` между этим каналом и `ctx.Done()`. После `ctx.Done()` вызови `srv.Shutdown` с отдельным контекстом-таймаутом.
</details>

## ✅ Решения

<details><summary>🔒 Решение задания 1 — потокобезопасный репозиторий</summary>

```go
type OrderRepo struct {
	mu     sync.RWMutex
	orders map[int]domain.Order
	nextID int
}

func (r *OrderRepo) Create(o domain.Order) domain.Order {
	r.mu.Lock()
	defer r.mu.Unlock()
	// ... как было
}

func (r *OrderRepo) Get(id int) (domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// ...
}

func (r *OrderRepo) Update(o domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// ...
}

func (r *OrderRepo) List() []domain.Order {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// ...
}
```
</details>

<details><summary>🔒 Решение задания 2 — CreateOrder</summary>

```go
// import "fmt", "slices", "strings"

func (s *OrderService) CreateOrder(customer string, items []ItemInput) (domain.Order, error) {
	customer = strings.TrimSpace(customer)
	if customer == "" {
		return domain.Order{}, fmt.Errorf("customer is required: %w", domain.ErrValidation)
	}
	if len(items) == 0 {
		return domain.Order{}, fmt.Errorf("order must contain at least one item: %w", domain.ErrValidation)
	}

	order := domain.Order{Customer: customer, Status: domain.StatusNew, Items: make([]domain.Item, 0, len(items))}
	for _, in := range items {
		if in.Quantity <= 0 {
			return domain.Order{}, fmt.Errorf("product %d: quantity must be positive: %w", in.ProductID, domain.ErrValidation)
		}
		product, err := s.catalog.Product(in.ProductID)
		if err != nil {
			return domain.Order{}, fmt.Errorf("create order: %w", err)
		}
		order.Items = append(order.Items, domain.Item{ProductID: product.ID, Quantity: in.Quantity, Price: product.Price})
		order.Total += product.Price * int64(in.Quantity)
	}
	return s.store.Create(order), nil
}
```
</details>

<details><summary>🔒 Решение задания 3 — ChangeStatus</summary>

```go
func (s *OrderService) ChangeStatus(id int, to domain.Status) (domain.Order, error) {
	order, err := s.store.Get(id)
	if err != nil {
		return domain.Order{}, err
	}
	if !slices.Contains(allowedTransitions[order.Status], to) {
		return domain.Order{}, fmt.Errorf("order %d: %s → %s: %w", id, order.Status, to, domain.ErrInvalidTransition)
	}
	order.Status = to
	if err := s.store.Update(order); err != nil {
		return domain.Order{}, err
	}
	return order, nil
}
```
</details>

<details><summary>🔒 Решение задания 4 — маппинг ошибок</summary>

```go
// import "errors"

func writeServiceError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, domain.ErrValidation):
		status = http.StatusUnprocessableEntity
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, domain.ErrInvalidTransition):
		status = http.StatusConflict
	}
	if status == http.StatusInternalServerError {
		log.Printf("internal error: %v", err) // детали — только в лог
		writeJSON(w, status, map[string]string{"error": "internal error"})
		return
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
```
</details>

<details><summary>🔒 Решение задания 5 — graceful shutdown</summary>

```go
// import "errors", "fmt"

func run(ctx context.Context, addr string) error {
	srv := &http.Server{Addr: addr, Handler: newApp(), ReadHeaderTimeout: 5 * time.Second}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("🛒 Orders API слушает %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err // например, порт занят
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("listen: %w", err)
	case <-ctx.Done():
	}

	log.Println("⏳ получен сигнал остановки, завершаем активные запросы...")
	// Новый контекст: родительский ctx уже отменён и не подходит для ожидания.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}
```
`ReadHeaderTimeout` защищает от атаки Slowloris: клиент открывает соединение и шлёт заголовки по байту в минуту.
</details>

## 🚀 Что дальше (по желанию)
- Заменить in-memory репозиторий на PostgreSQL (`database/sql` + `pgx`). Сервис при этом не меняется, в этом и смысл интерфейсов.
- Перенести middleware из модуля 07 (RequestID, Logger, Recoverer) и подключить их к API.
- Добавить пагинацию в `GET /orders` (модуль 08).
- Структурные логи через `log/slog`, конфиг через переменные окружения.
- Dockerfile и `docker compose up`.

## 🧠 Вопросы для самопроверки
1. Почему сервис зависит от интерфейсов `OrderStore` и `ProductCatalog`, а не от `*repo.OrderRepo`?
2. Что гарантирует директория `internal/`?
3. Почему `srv.Shutdown` вызывается с **новым** контекстом, а не с `ctx` из аргумента?
4. Почему `http.ErrServerClosed` не считается ошибкой?
5. Где в этом проекте должна жить транзакция «создать заказ + списать остатки», если появится БД?
