# Модуль 04 · Интерфейсы, методы и паттерн Repository

## 🎯 Цель
Понять интерфейсы в Go (неявная реализация, полиморфизм, `fmt.Stringer`), разницу между value- и pointer-ресиверами и паттерн **Repository + Service**. На нём держится почти любой Go-бэкенд: хендлер → сервис → репозиторий.

## 📚 Теория

**Интерфейс** — это набор методов. Тип реализует интерфейс **неявно**: достаточно иметь все нужные методы, слово `implements` не нужно.
```go
type Stringer interface { String() string }
func (p Product) String() string { ... } // Product теперь fmt.Stringer
```

**Проверка реализации на этапе компиляции** — идиома, которая ловит ошибку сразу, а не в рантайме:
```go
var _ ProductRepository = (*InMemoryProductRepo)(nil)
```

**Value receiver и pointer receiver:**
```go
func (r Repo)  A() { r.count++ } // r — КОПИЯ, изменение потеряется
func (r *Repo) B() { r.count++ } // r — указатель, меняем оригинал
```
Правило: если метод меняет состояние или структура большая, используй указатель. Внутри одного типа лучше не смешивать оба вида ресиверов.

⚠️ Коварный момент: map внутри структуры — ссылочный тип. Поэтому при value receiver запись в `r.items` **сохранится**, а `r.nextID++` потеряется. Получается «полуработающий» баг.

**Repository** прячет хранение данных за интерфейсом. **Service** содержит бизнес-правила и зависит от интерфейса, а не от конкретной БД:
```
HTTP handler → ProductService → ProductRepository (interface)
                                   ├── InMemoryProductRepo (сейчас)
                                   └── PostgresProductRepo (завтра)
```
Плюс для тестов: сервис можно проверять со «стабом» вместо настоящей БД (см. `stubRepo` в `service_test.go`).

**Принцип:** «Accept interfaces, return structs» — принимай интерфейсы, возвращай конкретные типы.

## 🗂 Файлы
| Файл | Что внутри |
|---|---|
| `repository.go` | **Работаешь здесь**: интерфейс и in-memory реализация |
| `service.go` | **И здесь**: бизнес-логика покупки, Stringer |
| `mock_data.go` | 4 товара; у «Мыши» Stock = 0 |
| `repository_test.go`, `service_test.go` | Автотесты. **Не меняй** |

## 📝 Задания

| # | Где | Тип | Ожидаемое |
|---|---|---|---|
| 1 | `InMemoryProductRepo` | 🐞 не компилируется | пайплайн падает на стадии **build**: `missing method Delete`. Реализуй `Delete(id int) error` (нет товара → обёрнутая `ErrNotFound`) |
| 2 | `List` | 🐞 | всегда отсортирован по ID: `[1 2 3 4]` |
| 3 | `Create` | 🐞 | два новых товара получают ID `5` и `6` |
| 4 | — | 🧠 | разберись, почему в задании 3 товар всё же «сохраняется» в map, а ID нет (см. теорию) |
| 5 | `ProductService.Buy` | ✍️ | `Buy(1,3)` → Stock 7 и вызов `repo.Update`; ошибки `ErrInvalidQuantity` / `ErrOutOfStock` / `ErrNotFound`; при ошибке `Update` не вызывается |
| 6 | `Product.String` | ✍️ | `"Клавиатура (3500 ₽, остаток 10)"` |

Ожидаемый вывод `go run ./04-interfaces-repository` после решения:
```
== Каталог ==
  Клавиатура (3500 ₽, остаток 10)
  Мышь (1200 ₽, остаток 0)
  Монитор (18000 ₽, остаток 3)
  Наушники (5600 ₽, остаток 7)
== Созданы товары с ID 5 и 6 ==
Куплено! Остаток: 7
Покупка мыши: buy "Мышь": want 1, have 0: not enough items in stock
Стоимость склада: 150700
```

## ▶️ Как запустить
```bash
go run ./04-interfaces-repository
go run ./tools/ci 04        # или: make ci M=04
```

## 💡 Подсказки

<details><summary>Подсказка к заданию 1</summary>

Компилятор прямо пишет, чего не хватает: `missing method Delete`. Удалить ключ из мапы можно встроенной функцией `delete(m, key)`. Перед удалением проверь, что ключ существует.
</details>

<details><summary>Подсказка к заданию 2</summary>

Обход map идёт в случайном порядке. После сборки слайса отсортируй его через `sort.Slice` по `ID`.
</details>

<details><summary>Подсказка к заданию 3</summary>

Сравни ресивер `Create` с ресиверами остальных методов. Не хватает одного символа.
</details>

<details><summary>Подсказка к заданию 5</summary>

Порядок: проверить qty → `GetByID` → проверить Stock → уменьшить → `Update` → вернуть. Ошибки оборачивай через `%w`, тогда тесты на `errors.Is` пройдут.
</details>

## ✅ Решения

<details><summary>🔒 Решение задания 1 — Delete</summary>

```go
func (r *InMemoryProductRepo) Delete(id int) error {
	if _, ok := r.items[id]; !ok {
		return fmt.Errorf("delete product %d: %w", id, ErrNotFound)
	}
	delete(r.items, id)
	return nil
}
```
</details>

<details><summary>🔒 Решение задания 2 — List</summary>

```go
import "sort"

func (r *InMemoryProductRepo) List() []Product {
	result := make([]Product, 0, len(r.items))
	for _, p := range r.items {
		result = append(result, p)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}
```
</details>

<details><summary>🔒 Решение задания 3 — Create</summary>

```go
func (r *InMemoryProductRepo) Create(p Product) Product { // * — pointer receiver
	p.ID = r.nextID
	r.nextID++
	r.items[p.ID] = p
	return p
}
```
С value receiver метод получал копию структуры. `nextID++` менял копию, а map общая (внутри неё указатель), поэтому запись в `items` проходила, но под одним и тем же ID.
</details>

<details><summary>🔒 Решение задания 5 — Buy</summary>

```go
func (s *ProductService) Buy(id, qty int) (Product, error) {
	if qty <= 0 {
		return Product{}, ErrInvalidQuantity
	}
	p, err := s.repo.GetByID(id)
	if err != nil {
		return Product{}, fmt.Errorf("buy: %w", err)
	}
	if p.Stock < qty {
		return Product{}, fmt.Errorf("buy %q: want %d, have %d: %w", p.Name, qty, p.Stock, ErrOutOfStock)
	}
	p.Stock -= qty
	if err := s.repo.Update(p); err != nil {
		return Product{}, fmt.Errorf("buy: %w", err)
	}
	return p, nil
}
```
</details>

<details><summary>🔒 Решение задания 6 — String</summary>

```go
func (p Product) String() string {
	return fmt.Sprintf("%s (%d ₽, остаток %d)", p.Name, p.Price, p.Stock)
}
```
</details>

## 🧠 Вопросы для самопроверки
1. Что значит «интерфейсы в Go реализуются неявно»? Какие у этого плюсы?
2. Почему `var _ I = (*T)(nil)` — хорошая практика?
3. Если метод объявлен с pointer receiver, реализует ли интерфейс значение `T` (не `*T`)?
4. Что такое nil-интерфейс и почему `var p *T = nil; var i I = p; i != nil` — это true?
5. Зачем сервису зависеть от интерфейса репозитория, а не от конкретной структуры?
