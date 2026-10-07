# Модуль 09 · Горутины, WaitGroup, Mutex, sync.Once, race detector

## 🎯 Цель
Научиться безопасно работать с общими данными из многих горутин. На бэкенде это происходит постоянно: каждый HTTP-запрос обрабатывается в своей горутине. Разберём `sync.WaitGroup`, `sync.Mutex`/`RWMutex`, `sync.Once` и детектор гонок `-race`.

## 📚 Теория

**Горутина** — лёгкий поток, который запускается словом `go f()`. `main` (и тест) **не ждёт** горутины: если не синхронизироваться, функция вернётся раньше, чем горутины закончат работу.

**`sync.WaitGroup`** — счётчик «сколько горутин ещё работает»:
```go
var wg sync.WaitGroup
for i, u := range urls {
    wg.Add(1)              // ДО запуска горутины, не внутри неё!
    go func() {
        defer wg.Done()
        results[i] = fetch(u) // каждая горутина пишет в СВОЙ индекс — гонки нет
    }()
}
wg.Wait()
```

**Data race** — две горутины обращаются к одной памяти, хотя бы одна пишет, и синхронизации нет. Результат непредсказуем. Для map Go может аварийно завершить процесс: `fatal error: concurrent map writes`. `append` в общий слайс из нескольких горутин — тоже гонка.

**Детектор гонок:** `go test -race`, `go run -race`. Пайплайн курса всегда запускает тесты с `-race`.

**`sync.Mutex`** — «только один внутри»:
```go
c.mu.Lock()
defer c.mu.Unlock()
c.views[page]++
```
**`sync.RWMutex`** пускает много читателей (`RLock`) или одного писателя (`Lock`). Подходит, когда чтений намного больше, чем записей.

Правила: мьютекс хранится в структуре рядом с данными, которые он защищает. Структуру с мьютексом нельзя копировать (передавай по указателю, `go vet` за этим следит). Не отдавай наружу ссылку на защищённую map, отдавай копию (`maps.Clone`).

**`sync.Once`** выполняет функцию ровно один раз, даже при одновременных вызовах. Подходит для ленивой инициализации конфига, подключений, кэшей.

**`sync/atomic`** (`atomic.Int64` и т.п.) — быстрые счётчики без мьютекса, но только для простых чисел.

**Как распараллелить вычисление:** разбей данные на куски, каждой горутине дай свой кусок и свою ячейку для результата, дождись всех и сложи.

## 🗂 Файлы
| Файл | Задание |
|---|---|
| `counter.go` | 1, 2 — потокобезопасный счётчик |
| `fetch.go` | 3 — параллельный опрос сервисов |
| `sum.go` | 4 — параллельная сумма |
| `config.go` | 5 — ленивая загрузка конфига |
| `concurrency_test.go` | Автотесты с `-race`. **Не меняй** |

## 📝 Задания

| # | Где | Тип | Ожидаемое |
|---|---|---|---|
| 1 | `ViewCounter.Inc/Get` | 🐞 гонка | 50 горутин × 300 инкрементов → по 5000 на страницу, `-race` молчит |
| 2 | `Snapshot` | 🐞 | возвращает копию: изменение снапшота не портит счётчик |
| 3 | `FetchAll` | 🐞 | все 20 результатов, порядок как у `urls`, 10 запросов по 100 мс < 500 мс |
| 4 | `SumParallel` | ✍️ | сумма 1..100000 = 5000050000 при любом `workers` (в т.ч. 0, -1, 200000) |
| 5 | `ConfigProvider.Get` | 🐞 | `load` вызывается ровно 1 раз из 50 одновременных `Get` |

> 💥 Пока не сделано задание 1, тесты падают с `fatal error: concurrent map writes` и остальные не запускаются. **Начни с задания 1.**

## ▶️ Как запустить
```bash
go run ./tools/ci 09                         # 🚦 пайплайн (или make ci M=09)
go run -race ./09-goroutines-mutex           # демо с детектором гонок
go test -race -run TestFetchAll -v ./09-goroutines-mutex   # отдельный тест
```

## 💡 Подсказки

<details><summary>Подсказка к заданиям 1–2</summary>

Добавь поле `mu sync.RWMutex` в структуру. В `Inc` используй `Lock`/`Unlock`, в `Get` и `Snapshot` — `RLock`/`RUnlock`. Копию map можно сделать через `maps.Clone(c.views)` (Go 1.21+) или циклом.
</details>

<details><summary>Подсказка к заданию 3</summary>

Создай `results := make([]Result, len(urls))` (именно длину, а не ёмкость), пиши в `results[i]`, используй `WaitGroup`.
</details>

<details><summary>Подсказка к заданию 4</summary>

`chunk := (len(nums) + workers - 1) / workers` — деление с округлением вверх. Каждая горутина пишет в `partial[w]`, а после `wg.Wait()` частичные суммы складываются. Не забудь про граничные случаи: `workers < 1` и `workers > len(nums)`.
</details>

<details><summary>Подсказка к заданию 5</summary>

Поле `once sync.Once` и `p.once.Do(func() { p.cfg = p.load() })`.
</details>

## ✅ Решения

<details><summary>🔒 Решение заданий 1–2 — ViewCounter</summary>

```go
import (
	"maps"
	"sync"
)

type ViewCounter struct {
	mu    sync.RWMutex
	views map[string]int
}

func (c *ViewCounter) Inc(page string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.views[page]++
}

func (c *ViewCounter) Get(page string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.views[page]
}

func (c *ViewCounter) Snapshot() map[string]int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return maps.Clone(c.views)
}
```
</details>

<details><summary>🔒 Решение задания 3 — FetchAll</summary>

```go
func FetchAll(urls []string, fetch func(url string) (int, error)) []Result {
	results := make([]Result, len(urls))
	var wg sync.WaitGroup
	for i, u := range urls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status, err := fetch(u)
			results[i] = Result{URL: u, Status: status, Err: err}
		}()
	}
	wg.Wait()
	return results
}
```
Начиная с Go 1.22 у каждой итерации цикла свои `i` и `u`, поэтому их можно использовать в замыкании без копирования. В старых версиях писали `i, u := i, u`.
</details>

<details><summary>🔒 Решение задания 4 — SumParallel</summary>

```go
func SumParallel(nums []int, workers int) int {
	if workers < 1 {
		workers = 1
	}
	if workers > len(nums) {
		workers = max(len(nums), 1)
	}
	chunk := (len(nums) + workers - 1) / workers
	partial := make([]int, workers)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		start := w * chunk
		end := min(start+chunk, len(nums))
		if start >= end {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, n := range nums[start:end] {
				partial[w] += n
			}
		}()
	}
	wg.Wait()

	total := 0
	for _, p := range partial {
		total += p
	}
	return total
}
```
Можно и через канал: каждая горутина отправляет свою сумму в `chan int`. Это тема следующего модуля.
</details>

<details><summary>🔒 Решение задания 5 — sync.Once</summary>

```go
type ConfigProvider struct {
	load func() Config
	once sync.Once
	cfg  Config
}

func (p *ConfigProvider) Get() Config {
	p.once.Do(func() {
		p.cfg = p.load()
	})
	return p.cfg
}
```
</details>

## 🧠 Вопросы для самопроверки
1. Почему `wg.Add(1)` нельзя вызывать внутри горутины?
2. Чем `Mutex` отличается от `RWMutex`? Когда `RWMutex` может оказаться медленнее?
3. Почему структуру с мьютексом нельзя копировать?
4. Ловит ли `-race` все гонки? (Подсказка: только те, что реально случились во время запуска.)
5. Что такое deadlock? Приведи пример с двумя мьютексами.
