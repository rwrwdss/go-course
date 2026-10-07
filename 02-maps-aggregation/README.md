# Модуль 02 · Мапы (map) и агрегация данных

## 🎯 Цель
Научиться группировать и считать данные с помощью `map`: суммы по клиентам, счётчики, множества, топ-N. Это основа любой аналитики, отчётов и дашбордов на бэкенде.

## 📚 Теория

**Создание map:**
```go
var m map[string]int          // nil-мапа: читать можно, ПИСАТЬ НЕЛЬЗЯ (будет panic)
m := map[string]int{}         // пустая мапа, писать можно
m := make(map[string]int, 10) // то же, с подсказкой ёмкости
```

**Чтение отсутствующего ключа** возвращает нулевое значение типа (`0`, `""`, `nil`), поэтому `m["x"]++` работает даже без явной инициализации ключа.

**«comma ok»** — способ понять, есть ли ключ на самом деле:
```go
v, ok := m["alice"] // ok == false, если ключа нет
```

**Множество (set).** Отдельного типа set в Go нет, вместо него используют `map[T]struct{}`. Пустая структура `struct{}` занимает 0 байт.
```go
seen := map[string]struct{}{}
seen["books"] = struct{}{}
if _, ok := seen["books"]; ok { /* уже было */ }
```

**Порядок обхода map случайный.** Go специально перемешивает его. Если результат должен быть упорядочен, собери ключи в слайс и отсортируй:
```go
sort.Slice(names, func(i, j int) bool { return names[i] < names[j] })
```

**Деньги не храним во float.** `0.1 + 0.2 != 0.3`. Используй целые числа (рубли или копейки) или decimal-библиотеку.

**Map не потокобезопасна.** Одновременная запись из нескольких горутин приводит к `fatal error: concurrent map writes`. Подробнее в модуле 09.

## 🗂 Файлы
| Файл | Что внутри |
|---|---|
| `orders.go` | **Здесь ты работаешь** |
| `mock_data.go` | 10 mock-заказов (статусы paid / pending / cancelled) |
| `main.go` | Демо-запуск |
| `orders_test.go` | Автотесты. **Не меняй** |

## 📝 Задания

| # | Функция | Тип | Ожидаемый результат на mock-данных |
|---|---|---|---|
| 1 | `TotalByCustomer` | 🐞 panic | `map[alice:3500 bob:45000 carol:12800 dave:3500 eve:3500]` (только `paid`); для пустого входа пустая мапа, не nil |
| 2 | `CountByStatus` | ✍️ | `map[cancelled:2 paid:7 pending:1]` |
| 3 | `UniqueCategories` | 🐞 | `[books electronics food]` |
| 4 | `TopCustomers` | ✍️ | `n=3` → `[bob carol alice]` (при равных суммах сортировка по имени) |
| 5 | `HasCustomer` | 🐞 | клиент с суммой 0 **существует** |

## ▶️ Как запустить
Из корня курса:
```bash
go run ./02-maps-aggregation     # демо (сейчас упадёт с panic, это задание 1)
go run ./tools/ci 02             # 🚦 пайплайн
make ci M=02
```

> 💥 Пока не исправлено задание 1, пайплайн показывает панику, и остальные тесты не запускаются: паника роняет весь тестовый процесс. Начни с задания 1.

## 💡 Подсказки

<details><summary>Подсказка к заданию 1</summary>

Текст паники `assignment to entry in nil map` говорит, что запись идёт в nil-мапу. Как создать мапу, в которую можно писать?
</details>

<details><summary>Подсказка к заданию 3</summary>

Заведи `seen := map[string]struct{}{}` и добавляй категорию в результат, только если её ещё нет в `seen`. В конце отсортируй.
</details>

<details><summary>Подсказка к заданию 4</summary>

1) Собери ключи мапы в слайс. 2) `sort.Slice` с компаратором «сначала по сумме убыв., потом по имени возр.». 3) Обрежь до `n`, но аккуратно: `n` может быть больше длины.
</details>

<details><summary>Подсказка к заданию 5</summary>

`_, ok := totals[name]`
</details>

## ✅ Решения

<details><summary>🔒 Решение задания 1 — TotalByCustomer</summary>

```go
func TotalByCustomer(orders []Order) map[string]int {
	totals := make(map[string]int) // было: var totals map[string]int (nil)
	for _, o := range orders {
		if o.Status != StatusPaid {
			continue
		}
		totals[o.Customer] += o.Amount
	}
	return totals
}
```
</details>

<details><summary>🔒 Решение задания 2 — CountByStatus</summary>

```go
func CountByStatus(orders []Order) map[string]int {
	counts := make(map[string]int)
	for _, o := range orders {
		counts[o.Status]++
	}
	return counts
}
```
</details>

<details><summary>🔒 Решение задания 3 — UniqueCategories</summary>

```go
func UniqueCategories(orders []Order) []string {
	seen := make(map[string]struct{})
	result := []string{}
	for _, o := range orders {
		if _, ok := seen[o.Category]; ok {
			continue
		}
		seen[o.Category] = struct{}{}
		result = append(result, o.Category)
	}
	sort.Strings(result)
	return result
}
```
</details>

<details><summary>🔒 Решение задания 4 — TopCustomers</summary>

```go
func TopCustomers(totals map[string]int, n int) []string {
	names := make([]string, 0, len(totals))
	for name := range totals {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		if totals[names[i]] != totals[names[j]] {
			return totals[names[i]] > totals[names[j]] // сумма по убыванию
		}
		return names[i] < names[j] // имя по алфавиту
	})
	if n < len(names) {
		names = names[:n]
	}
	return names
}
```
Не забудь добавить `import "sort"` в начало файла.
</details>

<details><summary>🔒 Решение задания 5 — HasCustomer</summary>

```go
func HasCustomer(totals map[string]int, name string) bool {
	_, ok := totals[name]
	return ok
}
```
</details>

## 🧠 Вопросы для самопроверки
1. Почему запись в nil-мапу паникует, а чтение нет?
2. Почему порядок обхода map в Go случайный, и чем это полезно?
3. Можно ли взять адрес элемента мапы `&m["k"]`? Почему?
4. Почему `map[string]struct{}` лучше, чем `map[string]bool`, для множества?
